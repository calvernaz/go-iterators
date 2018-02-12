package iterator

import (
	"testing"
	
	"fmt"
	
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSimpleIterator(t *testing.T) {
	items := generateItems(0, 10)
	computeNext := next(items)
	
	iterator := NewDefaultIterator(computeNext)
	total := 0
	for range items {
		assert.True(t, iterator.HasNext())
		
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		
		// sum up the ids
		i := next.(*Item)
		total += i.ID
	}
	assert.Equal(t, 45, total)
}

func TestCloseHandler(t *testing.T) {
	items := generateItems(0, 10)
	computeNext, idx := nextAndIndex(items)
	iterator := NewCloseableIterator(computeNext, func() error {
		*idx = -1
		return nil
	})
	
	for range items {
		iterator.HasNext()
		next, err := iterator.Next()
		assert.NotNil(t, next)
		assert.Nil(t, err)
		
	}
	
	// idx was incremented
	assert.Equal(t, 10, *idx)
	// after close, it should reset to -1
	err := iterator.Close()
	assert.Nil(t, err)
	assert.Equal(t, -1, *idx)
}

func TestTransform_WhenErrorOccurs(t *testing.T) {
	items := generateItems(0, 10)
	iterator := Items(items).Iterator()
	transformedIterator := Transform(iterator, func(item interface{}) (interface{}, error) {
		return nil, errors.Errorf("Failed transforming value: %+v", item)
	})
	
	for range items {
		hasNext := iterator.HasNext()
		assert.True(t, hasNext)
		
		next, err := transformedIterator.Next()
		assert.Nil(t, next)
		assert.NotNil(t, err)
	}
}

func TestTransform(t *testing.T) {
	items := generateItems(0, 10)
	iterator := Items(items).Iterator()
	iterator = Transform(iterator, func(item interface{}) (interface{}, error) {
		it := item.(*Item)
		return fmt.Sprintf("%s : %d", it.Name, it.ID), nil
	})
	
	for i := 1; i < len(items); i++ {
		hasNext := iterator.HasNext()
		assert.True(t, hasNext)
		
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		
		// assume transformation
		fnItem := next.(string)
		
		expected := fmt.Sprintf("%s : %d", items[i-1].Name, items[i-1].ID)
		assert.Equal(t, expected, fnItem)
	}
	iterator.Close()
}

func TestSkip(t *testing.T) {
	items := generateItems(0, 10)
	iterator := Items(items).Iterator()
	iterator = Skip(iterator, 4)
	
	for i := 4; i < len(items); i++ {
		hasNext := iterator.HasNext()
		assert.True(t, hasNext)
		
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		
		item := next.(*Item)
		assert.Equal(t, i, item.ID, fmt.Sprintf("Expected '%d' but got '%d'", i, item.ID))
	}
	
	iterator.Close()
}

func TestFilter(t *testing.T) {
	items := generateItems(0, 10)
	iterator := Items(items).Iterator()
	
	iterator = Filter(iterator, func(item interface{}) (bool, error) {
		myItem := item.(*Item)
		if (myItem.ID % 2) == 0 {
			return true, nil
		}
		return false, nil
	})
	
	for range items[:len(items)/2]{
		hasNext := iterator.HasNext()
		assert.True(t, hasNext)
		
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
		
		item := next.(*Item)
		assert.True(t, item.ID% 2 == 0)
	}
	iterator.Close()
}

func TestFilterNonNil(t *testing.T) {
	index := 0
	iterator := NewDefaultIterator(func() (interface{}, bool, error) {
		for index < 10 {
			for index%2 == 0 {
				index++
				return &Item{index, fmt.Sprintf("%d", index)}, false, nil
			}
			index++
			return nil, false, nil
		}
		return nil, true, nil
	})
	iterator = FilterNonNil(iterator)
	
	for {
		hasNext := iterator.HasNext()
		if !hasNext {
			return
		}
		
		next, err := iterator.Next()
		assert.Nil(t, err)
		assert.NotNil(t, next)
	}
}


func TestLimit(t *testing.T) {
	items := generateItems(0, 10)
	iterator := Items(items).Iterator()
	iterator = Limit(iterator, 6)
	
	i := 0
	for iterator.HasNext() {
		hasNext := iterator.HasNext()
		if !hasNext {
			break
		}
		iterator.Next()
		i++
	}
	iterator.Close()
	assert.Equal(t, 6, i)
}

func TestConcat(t *testing.T) {
	items0To10 := generateItems(0, 10)
	items10To15 := generateItems(10, 15)
	
	iterator0To10 := Items(items0To10).Iterator()
	iterator10To15 := Items(items10To15).Iterator()
	
	iterator := Concat(iterator0To10, iterator10To15)
	
	i := 0
	for iterator.HasNext() {
		hasNext := iterator.HasNext()
		if !hasNext {
			break
		}
		iterator.Next()
		i++
	}
	iterator.Close()
	assert.Equal(t, 15, i)
}


func TestDedup(t *testing.T) {
	items := itemsFromIds(1, 2, 2, 3, 4, 4, 5, 5, 5, 6, 7, 8, 8, 9, 10, 10, 10)
	iterator := Items(items).Iterator()
	iterator = Dedup(iterator, func(item1 interface{}, item2 interface{}) bool {
		var myItem1 = item1.(*Item)
		var myItem2 = item2.(*Item)
		return myItem1.ID == myItem2.ID
	})
	
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	i := 0
	for iterator.HasNext() {
		hasNext := iterator.HasNext()
		if !hasNext {
			break
		}
		item, _ := iterator.Next()
		
		mitem := item.(*Item)
		assert.Equal(t, expected[i], mitem.ID)
		i++
	}
	iterator.Close()
	assert.Equal(t, i, len(expected))
}

func TestMerge(t *testing.T) {
	
	items1 := itemsFromIds(2, 4, 7, 9)
	items2 := itemsFromIds(1, 5, 8, 10)
	items3 := itemsFromIds(3, 6, 11, 12)
	
	iterator1 := Items(items1).Iterator()
	iterator2 := Items(items2).Iterator()
	iterator3 := Items(items3).Iterator()
	
	iterator := Merge(func(item1 interface{}, item2 interface{}) int {
		var myItem1 = item1.(*Item)
		var myItem2 = item2.(*Item)
		if myItem1.ID == myItem2.ID {
			return 0
		} else if myItem1.ID > myItem2.ID {
			return 1
		} else {
			return -1
		}
	}, iterator3, iterator1, iterator2)
	
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	i := 0
	for iterator.HasNext() {
		hasNext := iterator.HasNext()
		if !hasNext {
			break
		}
		item, _ := iterator.Next()
		
		mitem := item.(*Item)
		assert.Equal(t, expected[i], mitem.ID)
		i++
	}
	
	iterator.Close()
	assert.Equal(t, i, len(expected))
}
