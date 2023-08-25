// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sm_gpa

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
)

type StateManagerParameters struct {
	// How many blocks may be stored in cache before old ones start being deleted
	BlockCacheMaxSize int
	// How long should the block stay in block cache before being deleted
	BlockCacheBlocksInCacheDuration time.Duration
	// How often should the block cache be cleaned
	BlockCacheBlockCleaningPeriod time.Duration
	// How often get block requests should be repeated
	StateManagerGetBlockRetry time.Duration
	// How often requests waiting for response should be checked for expired context
	StateManagerRequestCleaningPeriod time.Duration
	// How often timer tick fires in state manager
	StateManagerTimerTickPeriod time.Duration
	// This number of states will always be available in the database
	PruningMinStatesToKeep int
	// On single store pruning attempt at most this number of states will be deleted
	PruningMaxStatesToDelete int

	TimeProvider sm_gpa_utils.TimeProvider
}

func NewStateManagerParameters(tpOpt ...sm_gpa_utils.TimeProvider) StateManagerParameters {
	var tp sm_gpa_utils.TimeProvider
	if len(tpOpt) > 0 {
		tp = tpOpt[0]
	} else {
		tp = sm_gpa_utils.NewDefaultTimeProvider()
	}
	return StateManagerParameters{
		BlockCacheMaxSize:                 1000,
		BlockCacheBlocksInCacheDuration:   1 * time.Hour,
		BlockCacheBlockCleaningPeriod:     1 * time.Minute,
		StateManagerGetBlockRetry:         3 * time.Second,
		StateManagerRequestCleaningPeriod: 1 * time.Second,
		StateManagerTimerTickPeriod:       1 * time.Second,
		PruningMinStatesToKeep:            10000,
		PruningMaxStatesToDelete:          1000,
		TimeProvider:                      tp,
	}
}
