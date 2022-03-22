package util

import "time"

type timer time.Time

func NewTimer() timer {
	return timer(time.Now())
}

func (t timer) Duration() time.Duration {
	return time.Now().Sub(time.Time(t))
}
