// package implement a simple primitive to wait for readiness of concurrent modules
package ready

import (
	"golang.org/x/xerrors"
	"sync"
	"time"
)

type Ready struct {
	wg sync.WaitGroup
}

// New creates new ready object in 'not ready state'
func New() *Ready {
	r := &Ready{}
	r.wg.Add(1)
	return r
}

// Wait waits for a timeout until set ready
func (r *Ready) Wait(timeout ...time.Duration) error {
	t := 10 * time.Second
	if len(timeout) > 0 {
		t = timeout[0]
	}
	c := make(chan struct{})
	go func() {
		defer close(c)
		r.wg.Wait()
	}()
	select {
	case <-c:
		return nil
	case <-time.After(t):
		return xerrors.New("not ready after timeout")
	}
}

// SetReady sets the object ready
func (r *Ready) SetReady() {
	r.wg.Done()
}
