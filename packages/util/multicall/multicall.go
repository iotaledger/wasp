package multicall

import (
	"errors"
	"sync"
	"time"
)

type Response struct {
	Result interface{}
	Err    error
}

var ErrTimeout = errors.New("timeout occurred")

// MultiCall call functions is parallel goroutines with overall timeout.
// returns array of results and error value
func MultiCall(funs []func() (interface{}, error), timeout time.Duration) ([]Response, bool) {
	results := make([]Response, len(funs))
	for i := range results {
		results[i].Err = ErrTimeout
	}
	mutex := &sync.Mutex{}
	counter := 0
	var wg sync.WaitGroup
	chNormal := make(chan struct{})

	wg.Add(len(funs))

	go func() {
		for i, f := range funs {
			go func(i int, f func() (interface{}, error)) {
				res, err := f()
				mutex.Lock()
				defer mutex.Unlock()

				results[i].Err = err
				results[i].Result = res

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
			// if the function blocks, it will leak a goroutine and the channel
			wg.Wait()
			close(chNormal)
		}()
	}

	// in any case it returns a copy of the result array

	ret := make([]Response, len(funs))

	mutex.Lock()
	copy(ret, results)
	mutex.Unlock()

	success := true
	for i := range ret {
		if ret[i].Err != nil {
			success = false
		}
	}
	return ret, success
}
