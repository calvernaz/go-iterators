// A forward only immutable iterator over a collection of items.
// The iterator is a powerful mechanism to serve and transform data lazily reducing the memory requirements and
// improving the time to first result performances.

package iterator

import (
	"io"

	"github.com/pkg/errors"
)

type ComputeNext func() (interface{}, bool, error)
type Closer func() error

// An iterator over a stream of data
type Iterator interface {
	HasNext() bool
	Next() (interface{}, error)
	Peek() (item interface{}, e error)
	io.Closer
}

type State int

const (
	// We haven't yet computed or have already returned the element
	NotReady State = iota
	// We have computed the next element and haven't returned it yet.
	Ready
	// We have reached the end of the data and are finished.
	Done
	// We've suffered an error, kaput !!.
	Failed
)

var _ Iterator = (*DefaultIterator)(nil)
var _ io.Closer = (*DefaultIterator)(nil)

type DefaultIterator struct {
	state State
	next  interface{}
	err   error

	ComputeNext ComputeNext

	closer Closer
}

// Given a way to compute next, returns an iterator
func NewDefaultIterator(computeNext ComputeNext) Iterator {
	return &DefaultIterator{
		ComputeNext: computeNext,
	}
}

// Given a way to compute next and a close handler, return a closeable iterator
func NewCloseableIterator(computeNext ComputeNext, closer Closer) Iterator {
	return &DefaultIterator{
		ComputeNext: computeNext,
		closer:      closer,
	}
}

// Returns true if the iterator can be continued or false if the end of data has been reached.
// It returns an error if the check fails.
func (it *DefaultIterator) HasNext() bool {
	switch it.state {
	case Ready:
		return true
	case Done:
		return false
	case Failed:
		return false
		//, errors.New("metadata iterator in an error state")
	}
	return it.tryToComputeNext()
}

// Returns the next item in the iteration.
// This method should be always called in combination with the HasNext.
// If the iterator reached the end of data, the method will return an error
func (it *DefaultIterator) Next() (interface{}, error) {
	hasNext := it.HasNext()
	if it.err != nil {
		return nil, it.err
	}
	if !hasNext {
		return nil, errors.New("no such element") // is eof an error?
	}

	it.state = NotReady
	nextItem := it.next
	it.next = nil
	return nextItem, nil
}

//
func (it *DefaultIterator) tryToComputeNext() bool {
	it.state = Failed // temporary pessimism

	next, eod, err := it.ComputeNext()
	if err != nil { // we got an err, stated
		it.state = Failed
		it.err = err
		return false
	}

	if eod {
		it.state = Done
		return false
	}

	it.state = Ready
	it.next = next
	return true
}

// Returns the next element without continuing the iteration.
func (it *DefaultIterator) Peek() (interface{}, error) {
	hasNext := it.HasNext()
	// an error getting the next item
	if it.err != nil {
		return nil, it.err
	}

	// no more items
	if !hasNext {
		return nil, io.EOF
	}
	next := it.next
	return next, nil
}

func (it *DefaultIterator) Close() error {
	it.state = Done
	if it.closer != nil {
		err := it.closer()
		if err != nil {
			return err
		}
	}
	return nil
}
