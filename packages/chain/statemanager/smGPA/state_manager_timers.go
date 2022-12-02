// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPA

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
)

type StateManagerTimers struct {
	// How long should the block stay in block cache before being deleted
	BlockCacheBlocksInCacheDuration time.Duration
	// How often should the block cache be cleaned
	BlockCacheBlockCleaningPeriod time.Duration
	// How often get block requests should be repeated
	StateManagerGetBlockRetry time.Duration
	// How often requests waiting for response should be checked for expired context
	StateManagerRequestCleaningPeriod time.Duration

	TimeProvider smGPAUtils.TimeProvider
}

func NewStateManagerTimers(tpOpt ...smGPAUtils.TimeProvider) StateManagerTimers {
	var tp smGPAUtils.TimeProvider
	if len(tpOpt) > 0 {
		tp = tpOpt[0]
	} else {
		tp = smGPAUtils.NewDefaultTimeProvider()
	}
	return StateManagerTimers{
		BlockCacheBlocksInCacheDuration:   1 * time.Hour,
		BlockCacheBlockCleaningPeriod:     1 * time.Minute,
		StateManagerGetBlockRetry:         3 * time.Second,
		StateManagerRequestCleaningPeriod: 1 * time.Second,
		TimeProvider:                      tp,
	}
}
