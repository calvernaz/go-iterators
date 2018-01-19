package main

import (
	"database/sql"
	
	
	"github.com/calvernaz/go-iterators"
	_ "github.com/proullon/ramsql/driver"
	"fmt"
	"github.com/pkg/errors"
)

type address struct {
	number int
	street string
}

const selectAddressWhereUser = `SELECT address.street_number, address.street FROM address
							JOIN user_addresses ON address.id=user_addresses.address_id
							WHERE user_addresses.user_id = $1;`

func LoadUserAddresses(db *sql.DB, userID int64) error {
	// runs the query
	rows, err := db.Query(selectAddressWhereUser, userID)
	if err != nil {
		return err
	}
	
	// iterator next function
	next := func() (next interface{}, endOfData bool, err error) {
		if rows.Next() {
			var addr address
			if err := rows.Scan(&addr.number, &addr.street); err != nil {
				return nil, false, err
			}
			return &addr, false, nil
		}
		return nil, true, nil
	}
	
	// close function
	closeFn := func() error {
		if rows != nil {
			return rows.Close()
		}
		return nil
	}
	
	iter := iterator.NewCloseableIterator(next, closeFn)
	for {
		hasNext  := iter.HasNext()
		if !hasNext {
			break
		}
		
		elem, err := iter.Next()
		if err != nil {
			break
		}
		fmt.Println(elem)
	}
	
	defer iter.Close()
	
	return nil
}

//func main() {
//	batch := []string{
//		`CREATE TABLE address (id BIGSERIAL PRIMARY KEY, street TEXT, street_number INT);`,
//		`CREATE TABLE user_addresses (address_id INT, user_id INT);`,
//		`INSERT INTO address (street, street_number) VALUES ('rue Victor Hugo', 32);`,
//		`INSERT INTO address (street, street_number) VALUES ('boulevard de la République', 23);`,
//		`INSERT INTO address (street, street_number) VALUES ('rue Charles Martel', 5);`,
//		`INSERT INTO address (street, street_number) VALUES ('chemin du bout du monde ', 323);`,
//		`INSERT INTO address (street, street_number) VALUES ('boulevard de la liberté', 2);`,
//		`INSERT INTO address (street, street_number) VALUES ('avenue des champs', 12);`,
//		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 1);`,
//		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 1);`,
//		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 2);`,
//		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 3);`,
//		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 4);`,
//		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 5);`,
//	}
//
//	db, err := sql.Open("ramsql", "TestLoadUserAddresses")
//	if err != nil {
//		log.Fatalf("sql.Open : Error : %s\n", err)
//	}
//	defer db.Close()
//
//	for _, b := range batch {
//		_, err = db.Exec(b)
//		if err != nil {
//			log.Fatalf("sql.Exec: Error: %s\n", err)
//		}
//	}
//
//	LoadUserAddresses(db, 1)
//}

func main() {

	var tr iterator.TransformFunc = func(item interface{}) (interface{}, error) {
		i, ok := item.(*MyItem)
		if !ok {
			return nil, errors.New("failed casting item to type *MyItem")
		}
		i.Name = i.Name + "Tr"
		return i, nil
	}
	
	items := itemsArray(1, 10)
	iterator := MyItemArray(items).Iterator()
	
	iter := TeeIterator(iterator, tr)
	
	for {
		hasNext := iter.HasNext()
		if !hasNext {
			return
		}
		
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
	
	return t.TransformFunc(n), nil
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
