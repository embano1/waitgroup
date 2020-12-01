package waitgroup_test

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/embano1/waitgroup"
)

func ExampleWaitGroup_WaitTimeout() {
	var wg waitgroup.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// long running computation
		time.Sleep(time.Second)
	}()

	err := wg.WaitTimeout(time.Millisecond * 100)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Output: timed out
}

func ExampleAwait() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// long running computation
		time.Sleep(time.Second)
	}()

	err := waitgroup.Await(&wg, time.Millisecond*100)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Output: timed out
}

func ExampleAwaitWithError_timeout() {
	var eg errgroup.Group

	eg.Go(func() error {
		// long running computation
		time.Sleep(time.Second)
		return nil
	})

	err := waitgroup.AwaitWithError(&eg, time.Millisecond*100)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Output: timed out
}

func ExampleAwaitWithError_error() {
	var eg errgroup.Group

	eg.Go(func() error {
		return errors.New("did not work")
	})

	err := waitgroup.AwaitWithError(&eg, time.Millisecond*100)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Output: did not work
}
