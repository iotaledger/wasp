// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"
)

const (
	defaultPullStateRetryConst         = 2 * time.Second
	defaultPullStateNewBlockDelayConst = 10 * time.Second
	defaultGetBlockRetryConst          = 3 * time.Second
)

type Timers struct {
	PullStateRetry         *time.Duration
	PullStateNewBlockDelay *time.Duration
	GetBlockRetry          *time.Duration
}

func (tT Timers) setPullStateRetry(pullStateRetry time.Duration) Timers {
	tT.PullStateRetry = &pullStateRetry
	return tT
}

func (tT Timers) setPullStateNewBlockDelay(pullStateNewBlockDelay time.Duration) Timers {
	tT.PullStateNewBlockDelay = &pullStateNewBlockDelay
	return tT
}

func (tT Timers) setGetBlockRetry(getBlockRetry time.Duration) Timers {
	tT.GetBlockRetry = &getBlockRetry
	return tT
}

func (tT Timers) getPullStateRetry() time.Duration {
	if tT.PullStateRetry == nil {
		return defaultPullStateRetryConst
	}
	return *tT.PullStateRetry
}

func (tT Timers) getPullStateNewBlockDelay() time.Duration {
	if tT.PullStateNewBlockDelay == nil {
		return defaultPullStateNewBlockDelayConst
	}
	return *tT.PullStateNewBlockDelay
}

func (tT Timers) getGetBlockRetry() time.Duration {
	if tT.GetBlockRetry == nil {
		return defaultGetBlockRetryConst
	}
	return *tT.GetBlockRetry
}
