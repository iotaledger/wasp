// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Provides implementations for chain.ChainRequests methods
package chainimpl

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func (c *chainObj) GetRequestProcessingStatus(reqID iscp.RequestID) chain.RequestProcessingStatus {
	if c.IsDismissed() {
		return chain.RequestProcessingStatusUnknown
	}
	if c.consensus != nil {
		if c.mempool.HasRequest(reqID) {
			return chain.RequestProcessingStatusBacklog
		}
	}
	c.stateReader.SetBaseline()
	processed, err := blocklog.IsRequestProcessed(c.stateReader.KVStoreReader(), &reqID)
	if err != nil || !processed {
		return chain.RequestProcessingStatusUnknown
	}
	return chain.RequestProcessingStatusCompleted
}

func (c *chainObj) AttachToRequestProcessed(handler func(iscp.RequestID)) *events.Closure {
	closure := events.NewClosure(handler)
	c.eventRequestProcessed.Attach(closure)
	return closure
}

func (c *chainObj) DetachFromRequestProcessed(attachID *events.Closure) {
	c.eventRequestProcessed.Detach(attachID)
}
