package multicall

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrTimeout = errors.New("timeout occurred")

// MultiCall call functions is parallel goroutines with overall timeout.
// returns array of results and error value
func MultiCall(funs []func() error, timeout time.Duration) (bool, []error) {
	results := make([]error, len(funs))
	for i := range results {
		results[i] = ErrTimeout
	}
	mutex := &sync.Mutex{}
	counter := 0
	var wg sync.WaitGroup
	chNormal := make(chan struct{})

	wg.Add(len(funs))

	go func() {
		for i, f := range funs {
			go func(i int, f func() error) {
				err := f()
				mutex.Lock()
				defer mutex.Unlock()

				results[i] = err

				wg.Done()
				counter++
				if counter == len(funs) {
					chNormal <- struct{}{}
				}
			}(i, f)
		}
	}()
	select {
	case <-chNormal:
		close(chNormal)

	case <-time.After(timeout):
		go func() {
			// wait for all to finish and then cleanup
			// if some function blocks it will leak the goroutine and the channel
			wg.Wait()
			mutex.Lock()
			defer mutex.Unlock()
			close(chNormal)
		}()
	}

	// in any case it returns a copy of the result array
	errs := make([]error, len(funs))

	mutex.Lock()
	copy(errs, results)
	mutex.Unlock()

	success := true
	for i := range errs {
		if errs[i] != nil {
			success = false
		}
	}
	return success, errs
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
