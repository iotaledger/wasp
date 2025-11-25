// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type SyncStateMgr struct {
	c *Consensus
	//
	// Query for a proposal.
	proposedBaseAnchor         *isc.StateAnchor
	proposedBaseAnchorReceived bool
	stateProposalReceived      bool
	//
	// Query for a decided Virtual State.
	decidedBaseAnchor    *isc.StateAnchor
	decidedStateReceived bool
	//
	// Save the produced block.
	producedBlock         state.StateDraft // In the case of rotation the block will be nil.
	producedBlockReceived bool
	saveProducedBlockDone bool
}

func NewSyncStateMgr(
	c *Consensus,
) *SyncStateMgr {
	return &SyncStateMgr{c: c}
}

func (s *SyncStateMgr) ProposedBaseAnchorReceived(baseAnchor *isc.StateAnchor) []gpa.MessageOut {
	if s.proposedBaseAnchorReceived {
		return nil
	}
	s.proposedBaseAnchor = baseAnchor
	s.proposedBaseAnchorReceived = true
	return s.c.uponStateMgrStateProposalQueryInputsReady(s.proposedBaseAnchor)
}

func (s *SyncStateMgr) StateProposalConfirmedByStateMgr() []gpa.MessageOut {
	if s.stateProposalReceived {
		return nil
	}
	s.stateProposalReceived = true
	return s.c.uponStateMgrStateProposalReceived(s.proposedBaseAnchor)
}

func (s *SyncStateMgr) DecidedVirtualStateNeeded(decidedBaseAnchor *isc.StateAnchor) []gpa.MessageOut {
	if s.decidedBaseAnchor != nil {
		return nil
	}
	s.decidedBaseAnchor = decidedBaseAnchor
	return s.c.uponStateMgrDecidedStateQueryInputsReady(s.decidedBaseAnchor)
}

func (s *SyncStateMgr) DecidedVirtualStateReceived(
	chainState state.State,
) []gpa.MessageOut {
	if s.decidedStateReceived {
		return nil
	}
	s.decidedStateReceived = true
	return s.c.uponStateMgrDecidedStateReceived(chainState)
}

func (s *SyncStateMgr) BlockProduced(block state.StateDraft) []gpa.MessageOut {
	if s.producedBlockReceived {
		return nil
	}
	s.producedBlock = block
	s.producedBlockReceived = true
	return s.c.uponStateMgrSaveProducedBlockInputsReady(s.producedBlock)
}

func (s *SyncStateMgr) BlockSaved(block state.Block) []gpa.MessageOut {
	if s.saveProducedBlockDone {
		return nil
	}
	s.saveProducedBlockDone = true
	return s.c.uponStateMgrSaveProducedBlockDone(block)
}

// String tries to provide useful human-readable compact status.
func (s *SyncStateMgr) String() string {
	str := "SM"
	if s.stateProposalReceived && s.decidedStateReceived {
		return str + statusStrOK
	}
	if s.stateProposalReceived {
		str += "/proposal=OK"
	} else if !s.proposedBaseAnchorReceived {
		str += "/proposal=WAIT[BaseAnchor]"
	} else {
		str += "/proposal=WAIT[RespFromStateMgr]"
	}
	if s.decidedStateReceived {
		str += "/state=OK"
	} else if s.decidedBaseAnchor == nil {
		str += "/state=WAIT[AcsDecision]"
	} else {
		str += "/state=WAIT[RespFromStateMgr]"
	}
	if s.saveProducedBlockDone {
		str += "/state=OK"
	} else if s.producedBlock == nil {
		str += "/state=WAIT[BlockFromVM]"
	} else {
		str += "/state=WAIT[RespFromStateMgr]"
	}
	return str
}
