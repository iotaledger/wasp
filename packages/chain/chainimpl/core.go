// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Provides implementations for chain.ChainCore methods except the Enqueue*,
// which are provided in eventproc.go
package chainimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func (c *chainObj) ID() *iscp.ChainID {
	return c.chainID
}

func (c *chainObj) GetCommitteeInfo() *chain.CommitteeInfo {
	cmt := c.getCommittee()
	if cmt == nil {
		return nil
	}
	return &chain.CommitteeInfo{
		Address:       cmt.DKShare().Address,
		Size:          cmt.Size(),
		Quorum:        cmt.Quorum(),
		QuorumIsAlive: cmt.QuorumIsAlive(),
		PeerStatus:    cmt.PeerStatus(),
	}
}

func (c *chainObj) StateCandidateToStateManager(virtualState state.VirtualStateAccess, outputID ledgerstate.OutputID) {
	c.stateMgr.EnqueueStateCandidateMsg(virtualState, outputID)
}

func (c *chainObj) TriggerChainTransition(data *chain.ChainTransitionEventData) {
	c.eventChainTransition.Trigger(data)
}

func (c *chainObj) Processors() *processors.Cache {
	return c.procset
}

func (c *chainObj) GlobalStateSync() coreutil.ChainStateSync {
	return c.chainStateSync
}

// GetStateReader returns a new copy of the optimistic state reader, with own baseline
func (c *chainObj) GetStateReader() state.OptimisticStateReader {
	return state.NewOptimisticStateReader(c.db, c.chainStateSync)
}

func (c *chainObj) GetChainNodes() []peering.PeerStatusProvider {
	return c.chainPeers.PeerStatus()
}

func (c *chainObj) GetCandidateNodes() []*governance.AccessNodeInfo {
	return c.candidateNodes
}

func (c *chainObj) Log() *logger.Logger {
	return c.log
}
