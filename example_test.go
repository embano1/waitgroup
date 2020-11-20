package waitgroup_test

import (
	"fmt"
	"time"

	"github.com/embano1/waitgroup"
)

func ExampleWaitGroup_WaitTimeout() {
	var wg waitgroup.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// long running computation
		time.Sleep(time.Second * 3)
	}()

	err := wg.WaitTimeout(time.Second)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Output: timed out
}
