// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/state"
)

type SyncSM interface {
	//
	// State proposal.
	ProposedBaseAliasOutputReceived(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	StateProposalConfirmedByStateMgr() gpa.OutMessages
	//
	// Decided state.
	DecidedVirtualStateNeeded(decidedBaseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	DecidedVirtualStateReceived(stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages
	//
	// Save the block.
	BlockProduced(block state.Block) gpa.OutMessages
	BlockSaved() gpa.OutMessages
	//
	// Supporting stuff.
	String() string
}

type syncSMImpl struct {
	//
	// Query for a proposal.
	ProposedBaseAliasOutput         *isc.AliasOutputWithID
	stateProposalQueryInputsReadyCB func(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	stateProposalReceived           bool
	stateProposalReceivedCB         func(proposedAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	//
	// Query for a decided Virtual State.
	decidedBaseAliasOutput         *isc.AliasOutputWithID
	decidedStateQueryInputsReadyCB func(decidedBaseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	decidedStateReceived           bool
	decidedStateReceivedCB         func(stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages
	//
	// Save the produced block.
	producedBlock                  state.Block
	saveProducedBlockInputsReadyCB func(producedBlock state.Block) gpa.OutMessages
	saveProducedBlockDone          bool
	saveProducedBlockDoneCB        func() gpa.OutMessages
}

func NewSyncSM(
	stateProposalQueryInputsReadyCB func(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	stateProposalReceivedCB func(proposedAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	decidedStateQueryInputsReadyCB func(decidedBaseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	decidedStateReceivedCB func(stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages,
	saveProducedBlockInputsReadyCB func(producedBlock state.Block) gpa.OutMessages,
	saveProducedBlockDoneCB func() gpa.OutMessages,
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

func (sub *syncSMImpl) ProposedBaseAliasOutputReceived(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	if sub.ProposedBaseAliasOutput != nil {
		return nil
	}
	sub.ProposedBaseAliasOutput = baseAliasOutput
	return sub.stateProposalQueryInputsReadyCB(sub.ProposedBaseAliasOutput)
}

func (sub *syncSMImpl) StateProposalConfirmedByStateMgr() gpa.OutMessages {
	if sub.stateProposalReceived {
		return nil
	}
	sub.stateProposalReceived = true
	return sub.stateProposalReceivedCB(sub.ProposedBaseAliasOutput)
}

func (sub *syncSMImpl) DecidedVirtualStateNeeded(decidedBaseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	if sub.decidedBaseAliasOutput != nil {
		return nil
	}
	sub.decidedBaseAliasOutput = decidedBaseAliasOutput
	return sub.decidedStateQueryInputsReadyCB(decidedBaseAliasOutput)
}

func (sub *syncSMImpl) DecidedVirtualStateReceived(
	stateBaseline coreutil.StateBaseline,
	virtualStateAccess state.VirtualStateAccess,
) gpa.OutMessages {
	if sub.decidedStateReceived {
		return nil
	}
	sub.decidedStateReceived = true
	return sub.decidedStateReceivedCB(stateBaseline, virtualStateAccess)
}

func (sub *syncSMImpl) BlockProduced(block state.Block) gpa.OutMessages {
	if sub.producedBlock != nil {
		return nil
	}
	sub.producedBlock = block
	return sub.saveProducedBlockInputsReadyCB(sub.producedBlock)
}

func (sub *syncSMImpl) BlockSaved() gpa.OutMessages {
	if sub.saveProducedBlockDone {
		return nil
	}
	sub.saveProducedBlockDone = true
	return sub.saveProducedBlockDoneCB()
}

// Try to provide useful human-readable compact status.
func (sub *syncSMImpl) String() string {
	str := "SM"
	if sub.stateProposalReceived && sub.decidedStateReceived {
		return str + statusStrOK
	}
	if sub.stateProposalReceived {
		str += "/proposal=OK"
	} else if sub.ProposedBaseAliasOutput == nil {
		str += "/proposal=WAIT[params: baseAliasOutput]"
	} else {
		str += "/proposal=WAIT[RespFromStateMgr]"
	}
	if sub.decidedStateReceived {
		str += "/state=OK"
	} else if sub.decidedBaseAliasOutput == nil {
		str += "/state=WAIT[acs decision]"
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
