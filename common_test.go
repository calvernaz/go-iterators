package iterator

import "fmt"

type Item struct {
	Id   int
	Name string
}

type Items []Item

func (a Items) Iterator() Iterator {
	return NewCloseableIterator(next(a), func() error {
		return nil
	})
}


func next(items []Item) ComputeNext {
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

func nextAndIndex(items []Item) (ComputeNext, *int) {
	index := 0
	
	return func() (interface{}, bool, error) {
		if index >= len(items) {
			return nil, true, nil
		}
		
		n := &items[index]
		index++
		return n, false, nil
	}, &index
}



func generateItems(from int, to int) []Item {
	var items []Item
	for from < to {
		items = append(items, Item{
			Id:   from,
			Name: fmt.Sprintf("item_%04d", from),
		})
		from++
	}
	return items
}


func itemsFromIds(ids ...int) []Item {
	var items []Item
	for _, id := range ids {
		items = append(items, Item{
			Id: id,
			Name: fmt.Sprintf("item_%04d", id),
		})
	}
	return items
}

