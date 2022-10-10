// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package subsystemSM

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/state"
)

type SubsystemSM struct {
	//
	// Query for a proposal.
	ProposedBaseAliasOutput         *isc.AliasOutputWithID
	stateProposalQueryInputsReadyCB func(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	stateProposalReceived           bool
	stateProposalReceivedCB         func(proposedAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	//
	// Query for a decided Virtual State.
	decidedBaseAliasOutputID       *iotago.OutputID
	decidedBaseStateCommitment     *state.L1Commitment
	decidedStateQueryInputsReadyCB func(decidedBaseAliasOutputID *iotago.OutputID, decidedBaseStateCommitment *state.L1Commitment) gpa.OutMessages
	decidedBaseAliasOutput         *isc.AliasOutputWithID
	decidedStateReceivedCB         func(aliasOutput *isc.AliasOutputWithID, stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages
}

func New(
	stateProposalQueryInputsReadyCB func(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	stateProposalReceivedCB func(proposedAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	decidedStateQueryInputsReadyCB func(decidedBaseAliasOutputID *iotago.OutputID, decidedBaseStateCommitment *state.L1Commitment) gpa.OutMessages,
	decidedStateReceivedCB func(aliasOutput *isc.AliasOutputWithID, stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages,
) *SubsystemSM {
	return &SubsystemSM{
		stateProposalQueryInputsReadyCB: stateProposalQueryInputsReadyCB,
		stateProposalReceivedCB:         stateProposalReceivedCB,
		decidedStateQueryInputsReadyCB:  decidedStateQueryInputsReadyCB,
		decidedStateReceivedCB:          decidedStateReceivedCB,
	}
}

func (sub *SubsystemSM) ProposedBaseAliasOutputReceived(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	if sub.ProposedBaseAliasOutput != nil {
		return nil
	}
	sub.ProposedBaseAliasOutput = baseAliasOutput
	return sub.stateProposalQueryInputsReadyCB(sub.ProposedBaseAliasOutput)
}

func (sub *SubsystemSM) StateProposalConfirmedByStateMgr() gpa.OutMessages {
	if sub.stateProposalReceived {
		return nil
	}
	sub.stateProposalReceived = true
	return sub.stateProposalReceivedCB(sub.ProposedBaseAliasOutput)
}

func (sub *SubsystemSM) DecidedVirtualStateNeeded(decidedBaseAliasOutputID *iotago.OutputID, decidedBaseStateCommitment *state.L1Commitment) gpa.OutMessages {
	if sub.decidedBaseAliasOutputID != nil {
		return nil
	}
	sub.decidedBaseAliasOutputID = decidedBaseAliasOutputID
	sub.decidedBaseStateCommitment = decidedBaseStateCommitment
	return sub.decidedStateQueryInputsReadyCB(decidedBaseAliasOutputID, decidedBaseStateCommitment)
}

func (sub *SubsystemSM) DecidedVirtualStateReceived(
	aliasOutput *isc.AliasOutputWithID,
	stateBaseline coreutil.StateBaseline,
	virtualStateAccess state.VirtualStateAccess,
) gpa.OutMessages {
	if sub.decidedBaseAliasOutput != nil {
		return nil
	}
	sub.decidedBaseAliasOutput = aliasOutput
	return sub.decidedStateReceivedCB(aliasOutput, stateBaseline, virtualStateAccess)
}

// Try to provide useful human-readable compact status.
func (sub *SubsystemSM) String() string {
	str := "SM"
	if sub.stateProposalReceived && sub.decidedBaseAliasOutput != nil {
		return str + "/OK"
	}
	if sub.stateProposalReceived {
		str += "/proposal=OK"
	} else if sub.ProposedBaseAliasOutput == nil {
		str += "/proposal=WAIT[params: baseAliasOutput]"
	} else {
		str += "/proposal=WAIT[RespFromStateMgr]"
	}
	if sub.decidedBaseAliasOutput != nil {
		str += "/state=OK"
	} else if sub.decidedBaseAliasOutputID == nil {
		str += "/state=WAIT[acs decision]"
	} else {
		str += "/state=WAIT[RespFromStateMgr]"
	}
	return str
}
