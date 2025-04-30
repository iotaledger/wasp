// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"time"

	"github.com/iotaledger/wasp/packages/util/timeutil"
)

type StateManagerParameters struct {
	// How many blocks may be stored in cache before old ones start being deleted
	BlockCacheMaxSize int
	// How long should the block stay in block cache before being deleted
	BlockCacheBlocksInCacheDuration time.Duration
	// How often should the block cache be cleaned
	BlockCacheBlockCleaningPeriod time.Duration
	// How many nodes should get block request be sent to
	StateManagerGetBlockNodeCount int
	// How often get block requests should be repeated
	StateManagerGetBlockRetry time.Duration
	// How often requests waiting for response should be checked for expired context
	StateManagerRequestCleaningPeriod time.Duration
	// How often state manager status information should be written to log
	StateManagerStatusLogPeriod time.Duration
	// How often timer tick fires in state manager
	StateManagerTimerTickPeriod time.Duration
	// This number of states will always be available in the database
	PruningMinStatesToKeep int
	// On single store pruning attempt at most this number of states will be deleted
	PruningMaxStatesToDelete int

	TimeProvider timeutil.TimeProvider
}

func NewStateManagerParameters(tpOpt ...timeutil.TimeProvider) StateManagerParameters {
	var tp timeutil.TimeProvider
	if len(tpOpt) > 0 {
		tp = tpOpt[0]
	} else {
		tp = timeutil.NewDefaultTimeProvider()
	}
	return StateManagerParameters{
		BlockCacheMaxSize:                 1000,
		BlockCacheBlocksInCacheDuration:   1 * time.Hour,
		BlockCacheBlockCleaningPeriod:     1 * time.Minute,
		StateManagerGetBlockNodeCount:     5,
		StateManagerGetBlockRetry:         3 * time.Second,
		StateManagerRequestCleaningPeriod: 5 * time.Minute,
		StateManagerStatusLogPeriod:       1 * time.Minute,
		StateManagerTimerTickPeriod:       1 * time.Second,
		PruningMinStatesToKeep:            10000,
		PruningMaxStatesToDelete:          10,
		TimeProvider:                      tp,
	}
}
