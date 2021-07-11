// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"time"
)

type ConsensusTimers struct {
	VMRunRetryToWaitForReadyRequests time.Duration
	BroadcastSignedResultRetry       time.Duration
	PostTxSequenceStep               time.Duration
	PullInclusionStateRetry          time.Duration
	ProposeBatchRetry                time.Duration
	ProposeBatchDelayForNewState     time.Duration
}

func NewConsensusTimers() ConsensusTimers {
	return ConsensusTimers{
		VMRunRetryToWaitForReadyRequests: 500 * time.Millisecond,
		BroadcastSignedResultRetry:       1 * time.Second,
		PostTxSequenceStep:               1 * time.Second,
		PullInclusionStateRetry:          1 * time.Second,
		ProposeBatchRetry:                500 * time.Millisecond,
		ProposeBatchDelayForNewState:     1 * time.Second, // experimental !!!!!
	}
}
