package ready

import (
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
	"time"
)

type Ready struct {
	r *atomic.Bool
}

func New() Ready {
	return Ready{atomic.NewBool(false)}

}

func (r *Ready) Wait(timeout ...time.Duration) error {
	t := 10 * time.Second
	if len(timeout) > 0 {
		t = timeout[0]
	}
	deadline := time.Now().Add(t)
	for !r.r.Load() {
		if time.Now().After(deadline) {
			return xerrors.New("not ready after timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (r *Ready) SetReady() {
	r.r.Store(true)
}
