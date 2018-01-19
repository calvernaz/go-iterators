package example

import (
	"github.com/calvernaz/go-iterators"
	"io"
	"errors"
	"encoding/csv"
	"log"
	"fmt"
	"strings"
)


// This example shows how to use an iterator applying the skip and transform operations over a csv reader.
// It creates a csv iterator where the next function wraps the csv `Read` function.
// Then it skips the header using the `Skip` operation and transforms the slice of fields (csv record)
// combining into a string using the `Transform` operation.

const csvRows = `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
func ExampleCsv() {
	iter, err := NewCsvIterator()
	if err != nil {
		log.Printf("error opening file")
		return
	}
	
	// skip header
	iter = iterator.Skip(iter, 1)
	
	// transform function that transforms a record into string
	iter = iterator.Transform(iter, func(item interface{}) (interface{}, error) {
		it := item.([]string)
		if len(it) >= 2 {
			return fmt.Sprintf("%s : %s", it[0], it[1]), nil
		}
		return nil, errors.New("some error")
	})
	
	// iterates over the transformed records
	for iter.HasNext() {
		record, err := iter.Next()
		if err != nil {
			continue
		}
		
		fmt.Printf("%s\n", record)
	}
	
	// Output:
	// Rob : Pike
	// Ken : Thompson
	// Robert : Griesemer
}


func NewCsvIterator() (iterator.Iterator, error) {
	var errorFile error

	reader := csv.NewReader(strings.NewReader(csvRows))
	
	// the next function
	iter := iterator.NewCloseableIterator(func() (interface{}, bool, error) {
		record, err := reader.Read()
		if err == io.EOF {
			return nil, true, nil
		} else if err != nil {
			if err, ok := err.(*csv.ParseError); ok && err.Err == csv.ErrFieldCount {
				log.Printf("error parsing")
				errorFile = errors.New("error parsing")
				return nil, false, errorFile
			}
		}
		return record, false, nil
	}, func() error { // close function
		return nil
	})
	
	return iter, nil
}

