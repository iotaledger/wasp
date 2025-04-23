package timeutil

import (
	"time"
)

type TimeProvider interface {
	SetNow(time.Time)
	GetNow() time.Time
	After(time.Duration) <-chan time.Time
}
