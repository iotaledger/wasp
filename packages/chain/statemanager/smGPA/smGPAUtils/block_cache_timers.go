// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"time"
)

type BlockCacheTimers struct {
	// How long should the block stay in cache before being deleted
	BlocksInCacheDuration time.Duration
	// How often should the block cache be cleened
	BlockCleaningPeriod time.Duration
}

func NewBlockCacheTimers() BlockCacheTimers {
	return BlockCacheTimers{
		BlocksInCacheDuration: 1 * time.Hour,
		BlockCleaningPeriod:   1 * time.Minute,
	}
}
