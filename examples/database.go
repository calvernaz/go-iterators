package main

import (
	"database/sql"
	
	
	"github.com/calvernaz/go-iterators"
	_ "github.com/proullon/ramsql/driver"
	"fmt"
)

// A typical database access usage where it executes a query that returns rows.
// Then we take advantage of the iterator creating the `next` function, which in the sql package
// fits quite well given that `Rows` implements the same concept, calling `Next()` before it call `Scan`.

type address struct {
	number int
	street string
}

const selectQueryStmt = `SELECT address.street_number, address.street FROM address
							JOIN user_addresses ON address.id=user_addresses.address_id
							WHERE user_addresses.user_id = $1;`

func LoadUserAddresses(db *sql.DB, userID int64) error {
	// runs the query
	rows, err := db.Query(selectQueryStmt, userID)
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
	defer iter.Close()
	
	for  iter.HasNext() {
		
		elem, err := iter.Next()
		if err != nil {
			break
		}
		fmt.Println(elem)
	}
	
	return nil
}
