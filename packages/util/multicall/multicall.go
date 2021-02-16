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
func MultiCall(funs []func() error, timeout time.Duration) []error {
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
	return results
}

func WrapErrors(errs []error) error {
	return WrapErrorsWithQuorum(errs, len(errs))
}

func WrapErrorsWithQuorum(errs []error, quorum int) error {
	ret := ""
	numSuccess := 0
	for i, err := range errs {
		ret += fmt.Sprintf("#%d: %v\n", i, err)
		if err == nil {
			numSuccess++
		}
	}
	if quorum >= len(errs) {
		quorum = len(errs)
	}
	if numSuccess >= quorum {
		return nil
	}
	return errors.New(ret)
}
