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
	decidedStateQueryInputsReadyCB func(decidedBaseAliasOutputID *iotago.OutputID) gpa.OutMessages
	decidedBaseAliasOutput         *isc.AliasOutputWithID
	decidedStateReceivedCB         func(aliasOutput *isc.AliasOutputWithID, stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages
}

func New(
	stateProposalQueryInputsReadyCB func(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	stateProposalReceivedCB func(proposedAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	decidedStateQueryInputsReadyCB func(decidedBaseAliasOutputID *iotago.OutputID) gpa.OutMessages,
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

func (sub *SubsystemSM) StateProposalConfirmedByStateMgr(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	if sub.stateProposalReceived {
		return nil
	}
	sub.stateProposalReceived = true
	return sub.stateProposalReceivedCB(baseAliasOutput)
}

func (sub *SubsystemSM) DecidedVirtualStateNeeded(decidedAliasOutputID *iotago.OutputID) gpa.OutMessages {
	if sub.decidedBaseAliasOutputID != nil {
		return nil
	}
	sub.decidedBaseAliasOutputID = decidedAliasOutputID
	return sub.decidedStateQueryInputsReadyCB(decidedAliasOutputID)
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
