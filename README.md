# `WaitGroup`
[![GoDoc](https://godoc.org/github.com/embano1/waitgroup?status.svg)](https://godoc.org/github.com/embano1/waitgroup)
[![Go Report
Card](https://goreportcard.com/badge/github.com/embano1/waitgroup)](https://goreportcard.com/report/github.com/embano1/waitgroup)

A simple wrapper around sync.WaitGroup with support for specifying a timeout to
return in bounded time, e.g. during shutdown.

For convenience two additional functions are provided in case existing
`sync.WaitGroup` or `errgroup.Group` should be used directly instead of wrapping.


# Why would I want this?

In case you use `sync.WaitGroup` (`"wg"`) to manage goroutines you might run into an
issue where `wg.Wait()` could block very long, err... forever, if one or more
goroutines do not finish their work in time/livelock, e.g. due to a missing
`wg.Done()`. ¯\\\_(ツ)\_/¯

# How to use

## Example with `waitgroup.WaitGroup{}`

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/embano1/waitgroup"
)

func main() {
	var wg waitgroup.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("I'm slow")
		time.Sleep(time.Second * 5)
	}()

	if err := wg.WaitTimeout(time.Second); err != nil {
		fmt.Printf("an error occurred: %v\n", err)
		os.Exit(1)
	}
}
```

Run this program:

```
go run cmd/main.go
I'm slow
an error occurred: timed out
exit status 1
```

## Example with `waitgroup.Await()`

```go
package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/embano1/waitgroup"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 5)
	}()

	if err := waitgroup.Await(&wg, time.Second); err != nil {
		fmt.Printf("an error occurred: %v\n", err)
		os.Exit(1)
	}
}
```

Run this program:

```
go run cmd/main.go
I'm slow
an error occurred: timed out
exit status 1
```

See [GoDoc](https://godoc.org/github.com/embano1/waitgroup) for more API details
and more examples.