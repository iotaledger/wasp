// package implement a simple primitive to wait for readiness of concurrent modules
package ready

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/atomic"
)

type Ready struct {
	name  string
	wg    sync.WaitGroup
	ready atomic.Bool
}

// New creates new ready object in 'not ready state'
func New(name string) *Ready {
	r := &Ready{name: name}
	r.wg.Add(1)
	return r
}

// Wait waits for a timeout until set ready
func (r *Ready) Wait(timeout ...time.Duration) error {
	if r.ready.Load() {
		return nil
	}
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
		r.ready.Store(true)
		return nil
	case <-time.After(t):
		return fmt.Errorf("'%s' not ready after timeout %v", r.name, t)
	}
}

func (r *Ready) MustWait(timeout ...time.Duration) {
	if err := r.Wait(timeout...); err != nil {
		panic(fmt.Sprintf("can't get ready '%s': %v", r.name, err))
	}
}

// SetReady sets the object ready
func (r *Ready) SetReady() {
	r.wg.Done()
}

func (r *Ready) IsReady() bool {
	return r.ready.Load()
}
