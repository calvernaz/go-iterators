package main

import (
	"github.com/calvernaz/go-iterators"
	"fmt"
	"github.com/pkg/errors"
)

// This example shows how to use an iterator to implement the unix tee pipe.

func main() {
	
	// transform function
	var tr iterator.TransformFunc = func(item interface{}) (interface{}, error) {
		i, ok := item.(*MyItem)
		if !ok {
			return nil, errors.New("failed casting item to type *MyItem")
		}
		i.Name = i.Name + "Tr"
		return i, nil
	}
	
	// slice of ints with an iterator
	items := itemsArray(1, 10)
	sliceIter := MyItemArray(items).Iterator()
	
	// the tee iterator
	iter := TeeIterator(sliceIter, tr)
	
	// iterate over the transformed values
	for iter.HasNext() {
		trItem, err := iter.Next()
		if err != nil {
			return
		}
		
		fmt.Println(trItem)
	}
}


type teeIterator struct {
	iterator.Iterator
	iterator.TransformFunc
}

func TeeIterator(iterator iterator.Iterator, fn iterator.TransformFunc) iterator.Iterator {
	return &teeIterator{ iterator, fn }
}

func (t *teeIterator) Next() (next interface{}, e error) {
	n, err := t.Iterator.Next()
	if  err != nil {
		return nil, err
	}
	
	return t.TransformFunc(n)
}

// Helpers
//
type MyItem struct {
	Id   int
	Name string
}

type MyItemArray []MyItem

func (a MyItemArray) Iterator() iterator.Iterator {
	return iterator.NewDefaultIterator(next(a))
}

func next(items []MyItem) iterator.ComputeNext {
	index := 0
	
	return func() (interface{}, bool, error) {
		if index >= len(items) {
			return nil, true, nil
		}
		
		n := &items[index]
		index++
		return n, false, nil
	}
}
// Adds the iterator behavior to a slice

func itemsArray(from int, to int) []MyItem {
	var items []MyItem
	for i := from; i <= to; i++ {
		items = append(items, MyItem{
			Id:   i,
			Name: fmt.Sprintf("item_%04d", i),
		})
	}
	return items
}

