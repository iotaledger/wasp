// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"time"
)

type StateManagerTimers struct {
	// period of state pull retry
	PullStateRetry time.Duration
	// how long delay state pull after state candidate received
	PullStateAfterStateCandidateDelay time.Duration
	GetBlockRetry                     time.Duration
}

func NewStateManagerTimers() StateManagerTimers {
	return StateManagerTimers{
		PullStateRetry:                    1 * time.Second,
		PullStateAfterStateCandidateDelay: 1 * time.Second,
		GetBlockRetry:                     3 * time.Second,
	}
}
