// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type SyncStateMgr interface {
	//
	// State proposal.
	ProposedBaseAnchorReceived(baseAnchor *isc.StateAnchor) gpa.OutMessages
	StateProposalConfirmedByStateMgr() gpa.OutMessages
	//
	// Decided state.
	DecidedVirtualStateNeeded(decidedBaseAnchor *isc.StateAnchor) gpa.OutMessages
	DecidedVirtualStateReceived(chainState state.State) gpa.OutMessages
	//
	// Save the block.
	BlockProduced(producedBlock state.StateDraft) gpa.OutMessages
	BlockSaved(savedBlock state.Block) gpa.OutMessages
	//
	// Supporting stuff.
	String() string
}

type syncStateMgrImpl struct {
	//
	// Query for a proposal.
	proposedBaseAnchor              *isc.StateAnchor
	proposedBaseAnchorReceived      bool
	stateProposalQueryInputsReadyCB func(baseAnchor *isc.StateAnchor) gpa.OutMessages
	stateProposalReceived           bool
	stateProposalReceivedCB         func(proposedAnchor *isc.StateAnchor) gpa.OutMessages
	//
	// Query for a decided Virtual State.
	decidedBaseAnchor              *isc.StateAnchor
	decidedStateQueryInputsReadyCB func(decidedBaseAnchor *isc.StateAnchor) gpa.OutMessages
	decidedStateReceived           bool
	decidedStateReceivedCB         func(chainState state.State) gpa.OutMessages
	//
	// Save the produced block.
	producedBlock                  state.StateDraft // In the case of rotation the block will be nil.
	producedBlockReceived          bool
	saveProducedBlockInputsReadyCB func(producedBlock state.StateDraft) gpa.OutMessages
	saveProducedBlockDone          bool
	saveProducedBlockDoneCB        func(savedBlock state.Block) gpa.OutMessages
}

func NewSyncStateMgr(
	stateProposalQueryInputsReadyCB func(baseAnchor *isc.StateAnchor) gpa.OutMessages,
	stateProposalReceivedCB func(proposedAnchor *isc.StateAnchor) gpa.OutMessages,
	decidedStateQueryInputsReadyCB func(decidedBaseAnchor *isc.StateAnchor) gpa.OutMessages,
	decidedStateReceivedCB func(chainState state.State) gpa.OutMessages,
	saveProducedBlockInputsReadyCB func(producedBlock state.StateDraft) gpa.OutMessages,
	saveProducedBlockDoneCB func(savedBlock state.Block) gpa.OutMessages,
) SyncStateMgr {
	return &syncStateMgrImpl{
		stateProposalQueryInputsReadyCB: stateProposalQueryInputsReadyCB,
		stateProposalReceivedCB:         stateProposalReceivedCB,
		decidedStateQueryInputsReadyCB:  decidedStateQueryInputsReadyCB,
		decidedStateReceivedCB:          decidedStateReceivedCB,
		saveProducedBlockInputsReadyCB:  saveProducedBlockInputsReadyCB,
		saveProducedBlockDoneCB:         saveProducedBlockDoneCB,
	}
}

func (s *syncStateMgrImpl) ProposedBaseAnchorReceived(baseAnchor *isc.StateAnchor) gpa.OutMessages {
	if s.proposedBaseAnchorReceived {
		return nil
	}
	s.proposedBaseAnchor = baseAnchor
	s.proposedBaseAnchorReceived = true
	return s.stateProposalQueryInputsReadyCB(s.proposedBaseAnchor)
}

func (s *syncStateMgrImpl) StateProposalConfirmedByStateMgr() gpa.OutMessages {
	if s.stateProposalReceived {
		return nil
	}
	s.stateProposalReceived = true
	return s.stateProposalReceivedCB(s.proposedBaseAnchor)
}

func (s *syncStateMgrImpl) DecidedVirtualStateNeeded(decidedBaseAnchor *isc.StateAnchor) gpa.OutMessages {
	if s.decidedBaseAnchor != nil {
		return nil
	}
	s.decidedBaseAnchor = decidedBaseAnchor
	return s.decidedStateQueryInputsReadyCB(decidedBaseAnchor)
}

func (s *syncStateMgrImpl) DecidedVirtualStateReceived(
	chainState state.State,
) gpa.OutMessages {
	if s.decidedStateReceived {
		return nil
	}
	s.decidedStateReceived = true
	return s.decidedStateReceivedCB(chainState)
}

func (s *syncStateMgrImpl) BlockProduced(block state.StateDraft) gpa.OutMessages {
	if s.producedBlockReceived {
		return nil
	}
	s.producedBlock = block
	s.producedBlockReceived = true
	return s.saveProducedBlockInputsReadyCB(s.producedBlock)
}

func (s *syncStateMgrImpl) BlockSaved(block state.Block) gpa.OutMessages {
	if s.saveProducedBlockDone {
		return nil
	}
	s.saveProducedBlockDone = true
	return s.saveProducedBlockDoneCB(block)
}

// Try to provide useful human-readable compact status.
func (s *syncStateMgrImpl) String() string {
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
