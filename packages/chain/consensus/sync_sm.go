// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type SyncSM interface {
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

type syncSMImpl struct {
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

func NewSyncSM(
	stateProposalQueryInputsReadyCB func(baseAnchor *isc.StateAnchor) gpa.OutMessages,
	stateProposalReceivedCB func(proposedAnchor *isc.StateAnchor) gpa.OutMessages,
	decidedStateQueryInputsReadyCB func(decidedBaseAnchor *isc.StateAnchor) gpa.OutMessages,
	decidedStateReceivedCB func(chainState state.State) gpa.OutMessages,
	saveProducedBlockInputsReadyCB func(producedBlock state.StateDraft) gpa.OutMessages,
	saveProducedBlockDoneCB func(savedBlock state.Block) gpa.OutMessages,
) SyncSM {
	return &syncSMImpl{
		stateProposalQueryInputsReadyCB: stateProposalQueryInputsReadyCB,
		stateProposalReceivedCB:         stateProposalReceivedCB,
		decidedStateQueryInputsReadyCB:  decidedStateQueryInputsReadyCB,
		decidedStateReceivedCB:          decidedStateReceivedCB,
		saveProducedBlockInputsReadyCB:  saveProducedBlockInputsReadyCB,
		saveProducedBlockDoneCB:         saveProducedBlockDoneCB,
	}
}

func (sub *syncSMImpl) ProposedBaseAnchorReceived(baseAnchor *isc.StateAnchor) gpa.OutMessages {
	if sub.proposedBaseAnchorReceived {
		return nil
	}
	sub.proposedBaseAnchor = baseAnchor
	sub.proposedBaseAnchorReceived = true
	return sub.stateProposalQueryInputsReadyCB(sub.proposedBaseAnchor)
}

func (sub *syncSMImpl) StateProposalConfirmedByStateMgr() gpa.OutMessages {
	if sub.stateProposalReceived {
		return nil
	}
	sub.stateProposalReceived = true
	return sub.stateProposalReceivedCB(sub.proposedBaseAnchor)
}

func (sub *syncSMImpl) DecidedVirtualStateNeeded(decidedBaseAnchor *isc.StateAnchor) gpa.OutMessages {
	if sub.decidedBaseAnchor != nil {
		return nil
	}
	sub.decidedBaseAnchor = decidedBaseAnchor
	return sub.decidedStateQueryInputsReadyCB(decidedBaseAnchor)
}

func (sub *syncSMImpl) DecidedVirtualStateReceived(
	chainState state.State,
) gpa.OutMessages {
	if sub.decidedStateReceived {
		return nil
	}
	sub.decidedStateReceived = true
	return sub.decidedStateReceivedCB(chainState)
}

func (sub *syncSMImpl) BlockProduced(block state.StateDraft) gpa.OutMessages {
	if sub.producedBlockReceived {
		return nil
	}
	sub.producedBlock = block
	sub.producedBlockReceived = true
	return sub.saveProducedBlockInputsReadyCB(sub.producedBlock)
}

func (sub *syncSMImpl) BlockSaved(block state.Block) gpa.OutMessages {
	if sub.saveProducedBlockDone {
		return nil
	}
	sub.saveProducedBlockDone = true
	return sub.saveProducedBlockDoneCB(block)
}

// Try to provide useful human-readable compact status.
func (sub *syncSMImpl) String() string {
	str := "SM"
	if sub.stateProposalReceived && sub.decidedStateReceived {
		return str + statusStrOK
	}
	if sub.stateProposalReceived {
		str += "/proposal=OK"
	} else if !sub.proposedBaseAnchorReceived {
		str += "/proposal=WAIT[BaseAnchor]"
	} else {
		str += "/proposal=WAIT[RespFromStateMgr]"
	}
	if sub.decidedStateReceived {
		str += "/state=OK"
	} else if sub.decidedBaseAnchor == nil {
		str += "/state=WAIT[AcsDecision]"
	} else {
		str += "/state=WAIT[RespFromStateMgr]"
	}
	if sub.saveProducedBlockDone {
		str += "/state=OK"
	} else if sub.producedBlock == nil {
		str += "/state=WAIT[BlockFromVM]"
	} else {
		str += "/state=WAIT[RespFromStateMgr]"
	}
	return str
}
