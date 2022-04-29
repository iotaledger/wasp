package util

import "time"

type Timer time.Time

func NewTimer() Timer {
	return Timer(time.Now())
}

func (t Timer) Duration() time.Duration {
	return time.Since(time.Time(t))
}
