// Package timeutil provides utilities for time-related operations,
// including implementations of time providers for testing and production use.
package timeutil

import (
	"time"
)

type defaultTimeProvider struct{}

var _ TimeProvider = &defaultTimeProvider{}

func NewDefaultTimeProvider() TimeProvider                          { return &defaultTimeProvider{} }
func (*defaultTimeProvider) SetNow(time.Time)                       {}
func (*defaultTimeProvider) GetNow() time.Time                      { return time.Now() }
func (*defaultTimeProvider) After(d time.Duration) <-chan time.Time { return time.After(d) }
