package iterator

import (
	"testing"
	
	"fmt"
	
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func next(items []MyItem) ComputeNext {
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

func nextAndIndex(items []MyItem) (ComputeNext, *int) {
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

func TestSimpleIterator(t *testing.T) {
	
	items := itemsArray(1, 10)
	computeNext := next(items)
	
	iterator := NewDefaultIterator(computeNext)
	for i := 1; i <= 10; i++ {
		hasNext, err := iterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, i, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", i, nextMyItem.Id))
	}
	
	hasNext, err := iterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
}

func TestFromArray(t *testing.T) {
	
	items := itemsArray(1, 10)
	
	iterator := MyItemArray(items).Iterator()
	
	for i := 1; i <= 10; i++ {
		hasNext, err := iterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, i, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", i, nextMyItem.Id))
	}
	hasNext, err := iterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
}

func TestCloseHandler(t *testing.T) {
	
	items := itemsArray(1, 10)
	
	next, idx := nextAndIndex(items)
	iterator := NewCloseableIterator(next, func() error {
		*idx = -1
		return nil
	})
	
	for i := 1; i <= 10; i++ {
		hasNext, err := iterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, i, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", i, nextMyItem.Id))
	}
	hasNext, err := iterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 10, *idx)
	err = iterator.Close()
	assert.Nil(t, err)
	assert.Equal(t, -1, *idx)
}


func TestToArray(t *testing.T) {
	items := itemsArray(1, 10)
	iterator := MyItemArray(items).Iterator()
	expected, err := ToArray(iterator)
	assert.Nil(t, err)
	assert.NotNil(t, expected)
}



func TestFilter(t *testing.T) {
	
	items := itemsArray(1, 10)
	iterator := MyItemArray(items).Iterator()
	filteredIterator := Filter(iterator, func(item interface{}) (returnIt bool) {
		nextMyItem := item.(*MyItem)
		if (nextMyItem.Id % 2) == 0 {
			return true
		}
		return false
	})
	
	for i:=1; i<= len(items); i++ {
		
		if (items[i-1].Id % 2) != 0 {
			continue
		}
		
		hasNext, err := iterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		
		next, err := filteredIterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, i, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", i, nextMyItem.Id))
	}
	hasNext, err := iterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
	
}

func TestFilterNonNil(t *testing.T) {
	index := 0
	iterator := &DefaultIterator{
		ComputeNext: func() (next interface{}, endOfData bool, e error) {
			index ++
			if index > 10 {
				return nil, true, nil
			}
			switch index {
			case 1:
				return nil, false, nil
			case 2:
				return &MyItem{Id: 2, Name: "item_0002"}, false, nil
			case 3:
				return &MyItem{Id: 3, Name: "item_0003"}, false, nil
			case 4:
				return nil, false, nil
			case 5:
				return &MyItem{Id: 5, Name: "item_0005"}, false, nil
			case 6:
				return nil, false, nil
			case 7:
				return nil, false, nil
			case 8:
				return &MyItem{Id: 8, Name: "item_0008"}, false, nil
			case 9:
				return &MyItem{Id: 9, Name: "item_0009"}, false, nil
			case 10:
				return nil, false, nil
			}
			return nil, true, errors.New(fmt.Sprintf("invalid index %d", index))
		},
	}
	
	nonNilIterator := FilterNonNil(iterator)
	expectedIds := []int{2, 3, 5, 8, 9}
	
	for _, id := range expectedIds {
		hasNext, err := nonNilIterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		
		next, err := nonNilIterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, id, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", id, nextMyItem.Id))
	}
	hasNext, err := nonNilIterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
	
}

func TestTransform_WhenErrorOccurs(t *testing.T) {
	items := itemsArray(1, 10)
	iterator := MyItemArray(items).Iterator()
	transformedIterator := Transform(iterator, func(item interface{}) (interface{}, error) {
		return nil, errors.Errorf("Failed transforming value: %+v", item)
	})
	
	for i:=1; i<= len(items); i++ {
		hasNext, err := iterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		
		next, err := transformedIterator.Next()
		assert.Nil(t, next)
		assert.NotNil(t, err)
	}
}

func TestTransform(t *testing.T) {
	
	items := itemsArray(1, 10)
	iterator := MyItemArray(items).Iterator()
	transformedIterator := Transform(iterator, func(item interface{}) (interface{}, error) {
		nextMyItem := item.(*MyItem)
		return fmt.Sprintf("%s : %d", nextMyItem.Name, nextMyItem.Id), nil
	})
	
	for i:=1; i<= len(items); i++ {
		
		expected := fmt.Sprintf("%s : %d", items[i-1].Name, items[i-1].Id)
		
		hasNext, err := iterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		
		next, err := transformedIterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(string)
		assert.Equal(t, nextMyItem, expected, fmt.Sprintf("Expected '%s' but got '%s'", expected, nextMyItem))
	}
	hasNext, err := iterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
	
}

func TestConcat(t *testing.T){
	items1To10 := itemsArray(1, 10)
	items11To15 := itemsArray(11, 15)
	
	iterator1To10 := MyItemArray(items1To10).Iterator()
	iterator11To15 := MyItemArray(items11To15).Iterator()
	
	concatIterator := Concat(iterator1To10, iterator11To15)
	
	for i:=1; i<= 15; i++ {
		hasNext, err := concatIterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		next, err := concatIterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, i, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", i, nextMyItem.Id))
	}
	hasNext, err := concatIterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
}

func TestSkip(t *testing.T){
	items := itemsArray(1, 10)
	iterator := MyItemArray(items).Iterator()
	skipIterator := Skip(iterator, 4)
	for i:=5; i<= len(items); i++ {
		hasNext, err := skipIterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		next, err := skipIterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, i, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", i, nextMyItem.Id))
	}
	hasNext, err := skipIterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
}

func TestLimit(t *testing.T){
	items := itemsArray(1, 10)
	iterator := MyItemArray(items).Iterator()
	limitIterator := Limit(iterator, 6)
	for i:=1; i<= 6; i++ {
		hasNext, err := limitIterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		next, err := limitIterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, i, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", i, nextMyItem.Id))
	}
	hasNext, err := limitIterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
}

func TestMerge(t *testing.T){
	
	items1 := itemsArrayFromIds(2, 4, 7, 9)
	items2 := itemsArrayFromIds(1, 5, 8, 10)
	items3 := itemsArrayFromIds(3, 6, 11, 12)
	
	iterator1 := MyItemArray(items1).Iterator()
	iterator2 := MyItemArray(items2).Iterator()
	iterator3 := MyItemArray(items3).Iterator()
	
	mergedIt := Merge(func(item1 interface{}, item2 interface{}) (ret int) {
		var myItem1 = item1.(*MyItem)
		var myItem2 = item2.(*MyItem)
		if myItem1.Id == myItem2.Id {
			return 0
		}else if myItem1.Id > myItem2.Id {
			return 1
		}else {
			return -1
		}
	}, iterator3, iterator1, iterator2)
	
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	for i, id := range expected {
		hasNext, err := mergedIt.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext, "expected HasNext to return true, but got false. Iteration number %d", i)
		next, err := mergedIt.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, id, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", id, nextMyItem.Id))
	}
	hasNext, err := mergedIt.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
	
}

func TestDedup(t *testing.T){
	items := itemsArrayFromIds(1, 2, 2, 3, 4, 4, 5, 5, 5, 6, 7, 8, 8, 9, 10, 10, 10)
	iterator := MyItemArray(items).Iterator()
	dedupedIterator := Dedup(iterator, func(item1 interface{}, item2 interface{}) (equal bool) {
		var myItem1 = item1.(*MyItem)
		var myItem2 = item2.(*MyItem)
		return myItem1.Id == myItem2.Id
	})
	
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, id := range expected {
		hasNext, err := dedupedIterator.HasNext()
		assert.Nil(t, err)
		assert.True(t, hasNext)
		next, err := dedupedIterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		nextMyItem := next.(*MyItem)
		assert.Equal(t, id, nextMyItem.Id, fmt.Sprintf("Expected '%d' but got '%d'", id, nextMyItem.Id))
	}
	hasNext, err := dedupedIterator.HasNext()
	assert.Nil(t, err)
	assert.False(t, hasNext)
	
}

// Saves the iterator into an array
func ToArray(it Iterator) ([]interface{}, error) {
	
	var list []interface{}
	for {
		
		hasNext, errHasNext := it.HasNext()
		if errHasNext != nil {
			return nil, errHasNext
		}
		if !hasNext {
			break
		}
		
		item, errNext := it.Next()
		
		if errNext != nil {
			return nil, errNext
		}
		list = append(list, item)
		
	}
	
	return list, nil
}
// Helpers
//
type MyItem struct {
	Id   int
	Name string
}

type MyItemArray []MyItem

// Adds the iterator behavior to a slice


func (a MyItemArray) Iterator() Iterator {
	return NewCloseableIterator(next(a), func() error {
		return nil
	})
}


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


func itemsArrayFromIds(ids ...int) []MyItem {
	var items []MyItem
	for _, id := range ids {
		items = append(items, MyItem{
			Id: id,
			Name: fmt.Sprintf("item_%04d", id),
		})
	}
	return items
}
