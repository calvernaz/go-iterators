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

```

### Create an iterator that needs to close resources

Especially when databases and networking are involved, a programmer has to make sure that all the resources are dismissed after a particular task is completed.
For this reason an ```Iterator``` can be created specifying a ```Closer``` which will take care of closing the resources.
The programmer needs to make sure that the ```Close()``` method is always called, for example deferring the call just after creating the iterator.

```go
iter := NewCloseableIterator(
	func() (next interface{}, endOfData bool, e error) {
    // Here put the logic that is computing the next element.
    // If there is a next element return: next, false, nil
    // If an error occurs computin the next element return: nil, true, error
    // If there is no next element return: nil, true, nil
    },
    func() error {
        // Close the resources.
    },
)

defer iter.Close()
```
