package iterator

import "github.com/pkg/errors"

// Helper Functions

type PredicateFunc func(item interface{}) bool

// Creates a wrapper-iterator over the original that will filter elements according to the filter function specified
func Filter(iter Iterator, test PredicateFunc) Iterator {
	return &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			for {
				hasNext, err := iter.HasNext()
				if err != nil {
					return nil, false, err
					
				}
				if !hasNext {
					return nil, true, nil
				}
				var ret interface{}
				ret, err = iter.Next()
				if err != nil {
					return nil, true, err
				}
				if test(ret) {
					return ret, false, nil
				}
			}
		},
		closer: func() error {
			return iter.Close()
		},
	}
}

// Specific case of Filter that returns a wrapper-iterator over the original that will return only the non nil items
func FilterNonNil(it Iterator) Iterator {
	return Filter(it, func(item interface{}) bool {
		return item != nil
	})
}

type TransformFunc func(item interface{}) interface{}

// Creates an wrapper-iterator over the original that will transform elements according to the filter function specified
func Transform(iter Iterator, fn TransformFunc) Iterator {
	return &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			for {
				hasNext, err := iter.HasNext()
				if err != nil {
					return nil, false, err
					
				}
				if !hasNext {
					return nil, true, nil
				}
				var ret interface{}
				ret, err = iter.Next()
				if err != nil {
					return nil, true, err
				}
				
				return fn(ret), false, nil
			}
		},
		closer: func() (e error) {
			return iter.Close()
		},
	}
}

// Creates an wrapper-iterator over the original that will skip the first 'numberOfElementsToSkip' items
func Skip(it Iterator, skipNumber int) Iterator {
	skippedCountDown := skipNumber
	return &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			for skippedCountDown > 0 {
				hasNext, err := it.HasNext()
				if err != nil {
					return nil, false, err
				}
				if !hasNext {
					return nil, true, nil
				}
				_, _ = it.Next()
				skippedCountDown = skippedCountDown - 1
			}
			
			hasNext, err := it.HasNext()
			if err != nil {
				return nil, false, err
			}
			
			if !hasNext {
				return nil, true, nil
			}
			var ret interface{}
			ret, err = it.Next()
			if err != nil {
				return nil, true, err
			}
			return ret, false, nil
		},
		closer: func() (e error) {
			return it.Close()
		},
	}
}

// Creates an wrapper-iterator over the original that will iterate until there are no more items or the 'upperBound' is reached.
func Limit(it Iterator, upperBound int) Iterator {
	servedItems := 0
	return &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			if servedItems == upperBound {
				return nil, true, nil
			}
			
			hasNext, err := it.HasNext()
			if err != nil {
				return nil, false, err
			}
			if !hasNext {
				return nil, true, nil
			}
			var ret interface{}
			ret, err = it.Next()
			if err != nil {
				return nil, true, err
			}
			servedItems = servedItems + 1
			return ret, false, nil
		},
		closer: func() (e error) {
			return it.Close()
		},
	}
}

// Appends multiple iterators together exposing them as a single virtual iterator.
func Concat(iterators ...Iterator) Iterator {
	var currentIteratorIdx = 0
	var currentIterator = iterators[0]
	return &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			for {
				hasNext, err := currentIterator.HasNext()
				if err != nil {
					return nil, false, err
				}
				if !hasNext {
					currentIterator.Close()
					currentIteratorIdx ++
					if currentIteratorIdx < len(iterators) {
						currentIterator = iterators[currentIteratorIdx]
						continue
					}
					return nil, true, nil
				}
				var next interface{}
				next, err = currentIterator.Next()
				if err != nil {
					return nil, true, err
				}
				return next, false, nil
			}
		},
		closer: func() (e error) {
			var err error
			for _, it := range iterators {
				tmpErr := it.Close()
				if tmpErr != nil {
					if err != nil {
						err = tmpErr
					} else {
						err = errors.Wrap(err, tmpErr.Error())
					}
				}
			}
			return err
		},
	}
}

type CompareFunc func(item1 interface{}, item2 interface{}) int

// Merges multiple sorted iterators into a single sorted iterator.
func Merge(compareFn CompareFunc, iterators ...Iterator) Iterator {
	return &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			for {
				ret, err := selectMin(compareFn, iterators...)
				if err != nil {
					return nil, true, err
				}
				if ret == nil {
					return nil, true, nil
				}
				return ret, false, nil
			}
		},
		closer: func() (e error) {
			var err error
			for _, it := range iterators {
				tmpErr := it.Close()
				if tmpErr != nil {
					if err != nil {
						err = tmpErr
					} else {
						err = errors.Wrap(err, tmpErr.Error())
					}
				}
			}
			return err
		},
	}
}

type EqualsFunc func(item1 interface{}, item2 interface{}) bool

func Dedup(it Iterator, equalsFn EqualsFunc) Iterator {
	var prev interface{}
	return &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			for {
				hasNext, err := it.HasNext()
				if err != nil {
					return nil, false, err
				}
				if !hasNext {
					return nil, true, nil
				}
				ret, err := it.Next()
				if err != nil {
					return nil, true, err
				}
				if prev == nil || !equalsFn(prev, ret) {
					prev = ret
					return ret, false, nil
				}
			}
		},
		closer: func() (e error) {
			return it.Close()
		},
	}
}

func selectMin(compareFn CompareFunc, iterators ...Iterator) (interface{}, error) {
	var err error
	var hasNext bool
	var selected int
	var peek interface{}
	var currentSelection interface{}
	for i, it := range iterators {
		hasNext, err = it.HasNext()
		if err != nil {
			return nil, err
		}
		if hasNext {
			peek, err = it.Peek()
			if currentSelection == nil {
				//log.Printf("Set current selection to %v", peek)
				currentSelection = peek
				selected = i
			} else if compareFn(currentSelection, peek) > 0 { // The peek is lower than the current selection
				//log.Printf("Switch current selection from %v to %v", currentSelection, peek)
				currentSelection = peek
				selected = i
			}
		}
	}
	if currentSelection != nil {
		_, _ = iterators[selected].Next()
		return currentSelection, nil
	} else {
		return nil, nil
	}
}

