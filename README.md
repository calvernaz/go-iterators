[![Build Status](https://travis-ci.org/calvernaz/go-iterators.svg?branch=master)](https://travis-ci.org/calvernaz/go-iterators)
[![Coverage Status](https://coveralls.io/repos/github/calvernaz/go-iterators/badge.svg?branch=master)](https://coveralls.io/github/calvernaz/go-iterators?branch=master)

# go-iterators

The go-iterators project is a library offering the iterator pattern for Golang.

Why iterators are useful?

* They can be lazy and the data will be fetched just when needed.
* They fit a lots of use cases. From a simple slice iteration to data transformation to tree traversals.

## Usage examples

[Examples](examples/)

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

### Create an iterator from a slice

```go
func FisIterator(fis []os.FileInfo) iterator.Iterator {
	i := 0
	return iterator.NewDefaultIterator(func() (interface{}, bool, error) {
		if i >= len(fis) {
			return nil, true, nil
		}
		
		file := &fis[i]
		i++
		return file, false, nil
	})
}

```

## Credits

* [Silvano Riz](https://github.com/melozzola)
* [Gheorghe Prelipcean](https://gitlab.com/prelipceang)
