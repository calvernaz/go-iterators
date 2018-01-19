package iterator

import "github.com/pkg/errors"

// Helper Functions

type PredicateFunc func(item interface{}) bool

// Creates a wrapper-iterator over the original that will filter elements according to the filter function specified
func Filter(iter Iterator, test PredicateFunc) Iterator {
	return &DefaultIterator{
		ComputeNext: func() (interface{}, bool, error) {
			for iter.HasNext() {
				ret, err := iter.Next()
				if err != nil {
					return nil, true, err
				}
				if test(ret) {
					return ret, false, nil
				}
			}
			return nil, true, nil
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

type TransformFunc func(item interface{}) (interface{}, error)

// Creates an wrapper-iterator over the original that will transform elements according to the filter function specified
func Transform(iter Iterator, fn TransformFunc) Iterator {
	return &DefaultIterator{
		ComputeNext: func() (interface{}, bool, error) {
			for iter.HasNext() {
				ret, err := iter.Next()
				if err != nil {
					return nil, false, err
				}
				
				nextFn, err := fn(ret)
				return nextFn, false, err
			}
			return nil, true, nil
		},
		closer: func() (e error) {
			return iter.Close()
		},
	}
}

// Creates an wrapper-iterator over the original that will skip the first 'numberOfElementsToSkip' items
func Skip(it Iterator, howMany int) Iterator {
	return &DefaultIterator{
		ComputeNext: func() (interface{}, bool, error) {
			for howMany > 0 {
				hasNext:= it.HasNext()
				if !hasNext {
					return nil, true, nil
				}
				_, _ = it.Next()
				howMany -= 1
			}
			
			hasNext := it.HasNext()
			if !hasNext {
				return nil, true, nil
			}
			
			ret, err := it.Next()
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
	items := 0
	return &DefaultIterator{
		ComputeNext: func() (interface{}, bool, error) {
			if items == upperBound {
				return nil, true, nil
			}
			
			hasNext := it.HasNext()
			if !hasNext {
				return nil, true, nil
			}
			
			ret, err := it.Next()
			if err != nil {
				return nil, true, err
			}
			items = items + 1
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
	var iterator = iterators[0]
	return &DefaultIterator{
		ComputeNext: func() (interface{}, bool, error) {
			for {
				hasNext := iterator.HasNext()
				if !hasNext {
					iterator.Close()
					currentIteratorIdx ++
					if currentIteratorIdx < len(iterators) {
						iterator = iterators[currentIteratorIdx]
						continue
					}
					return nil, true, nil
				}
				
				next, err := iterator.Next()
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
		ComputeNext: func() (interface{}, bool, error) {
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
		ComputeNext: func() (interface{}, bool, error) {
			for it.HasNext() {
				ret, err := it.Next()
				if err != nil {
					return nil, true, err
				}
				
				if prev == nil || !equalsFn(prev, ret) {
					prev = ret
					return ret, false, nil
				}
			}
			return nil, true, nil
		},
		closer: func() (e error) {
			return it.Close()
		},
	}
}

func selectMin(compareFn CompareFunc, iterators ...Iterator) (interface{}, error) {
	var selected int
	var current interface{}
	for i, it := range iterators {
		hasNext := it.HasNext()
		if hasNext {
			peek, err := it.Peek()
			if err != nil {
				return nil, err
			}
			
			if current == nil {
				current = peek
				selected = i
			} else if compareFn(current, peek) > 0 { // The peek is lower than the current selection
				current = peek
				selected = i
			}
		}
	}
	if current != nil {
		_, _ = iterators[selected].Next()
		return current, nil
	} else {
		return nil, nil
	}
}
