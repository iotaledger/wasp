// package implements simple semaphore with timeout
package sema

import "time"

type Lock struct {
	ch chan struct{}
}

func New() *Lock {
	ret := &Lock{
		ch: make(chan struct{}, 1),
	}
	return ret
}

func (sem *Lock) Dispose() {
	close(sem.ch)
}

func (sem *Lock) Acquire(timeout time.Duration) bool {
	if timeout < 0 {
		sem.ch <- struct{}{}
		return true
	}
	select {
	case sem.ch <- struct{}{}:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (sem *Lock) Release() bool {
	select {
	case <-sem.ch:
		return true
	default:
		return false
	}
}
