# GO ITERATORS

The go-iterators project is a library offering the iterator pattern for Golang.

Why iterators are useful?

* They can be lazy and the data will be fetched just when needed.
* They fit a lots of use cases. From a simple slice iteration to data transformation to tree traversals.

## Usage examples

[Examples](examples/database.go)

### Create an iterator

Creating an iterator is as simple as defining a function, the function will have to compute the next item in the iteration.

```go
iter := NewDefaultIterator(func() (next interface{}, eod bool, err error) { 
    // Here put the logic that is computing the next element.
    // 1. If there is a next element return: next, false, nil
    // 2. If an error occurs computing the next element return: nil, false, error
    // 3. If there is no next element return: nil, true, nil 
})


defer iter.Close()
```


## Credits

* [Silvano Riz](https://github.com/melozzola)
* [Gheorghe Prelipcean](https://gitlab.com/prelipceang)
