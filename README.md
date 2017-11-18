# GO ITERATORS

The go-iterators project is a library offering the iterator pattern for Golang.

Why iterators are useful?

* They can be lazy and the data will be fetched just when needed.
* They fit a lots of use cases. From a simple slice iteration to data transformation to tree traversals.

## Usage examples

### Create an iterator

Creating an iterator is as simple as defining a function, the function will have to compute the next item in the iteration.

```go
iter := NewDefaultIterator(func() (next interface{}, endOfData bool, e error) { 
    // Here put the logic that is computing the next element.
    // If there is a next element return: next, false, nil
    // If an error occurs computin the next element return: nil, true, error
    // If there is no next element return: nil, true, nil 
})


defer iter.Close()