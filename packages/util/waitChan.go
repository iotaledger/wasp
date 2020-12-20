// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"time"
)

// WaitChan is something similar to sync.WaitGroup, just is based on channels
// and supports waiting with timeout.
type WaitChan struct {
	ch chan bool
}

func NewWaitChan() *WaitChan {
	return &WaitChan{ch: make(chan bool, 1)}
}

func (c *WaitChan) Done() {
	select {
	case <-c.ch:
		c.ch <- true
	default:
		c.ch <- true
	}
}

func (c *WaitChan) Wait() {
	<-c.ch
	c.ch <- true
}

func (c *WaitChan) WaitTimeout(timeout time.Duration) bool {
	select {
	case <-c.ch:
		c.ch <- true
		return true
	case <-time.After(timeout):
		return false
	}
}

func (c *WaitChan) Reset() {
	select {
	case <-c.ch:
	default:
	}
}
