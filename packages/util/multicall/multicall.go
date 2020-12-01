package multicall

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrTimeout = errors.New("MultiCall: timeout")

// MultiCall call functions is parallel goroutines with overall timeout.
// returns array of results and error value
func MultiCall(funs []func() error, timeout time.Duration) (bool, []error) {
	results := make([]error, len(funs))

	var wg sync.WaitGroup
	wg.Add(len(funs))

	for i, f := range funs {
		go func(i int, f func() error) {
			defer wg.Done()
			done := make(chan error)
			go func() {
				done <- f()
			}()
			select {
			case err := <-done:
				results[i] = err
			case <-time.After(timeout):
				results[i] = ErrTimeout
			}
		}(i, f)
	}

	wg.Wait()

	success := true
	for i := range results {
		if results[i] != nil {
			success = false
		}
	}
	return success, results
}

func WrapErrors(errs []error) error {
	ret := ""
	numErrors := 0
	for i, err := range errs {
		if err != nil {
			ret += fmt.Sprintf("#%d: %v\n", i, err)
			numErrors++
		}
	}
	if numErrors == 0 {
		return nil
	}
	return errors.New(ret)
}
