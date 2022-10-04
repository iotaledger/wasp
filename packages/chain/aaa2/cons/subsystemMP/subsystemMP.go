// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package subsystemMP

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type SubsystemMP struct {
	BaseAliasOutput       *isc.AliasOutputWithID
	DecidedRequestIDs     []isc.RequestID
	proposalInputsReadyCB func(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages
	proposalReceived      bool
	proposalReceivedCB    func(requestRefs []*isc.RequestRef) gpa.OutMessages
	requestsNeeded        bool
	requestsNeededCB      func(requestIDs []*isc.RequestRef) gpa.OutMessages
	requestsReceived      bool
	requestsReceivedCB    func(requests []isc.Request) gpa.OutMessages
}

func New(
	proposalInputsReadyCB func(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages,
	proposalReceivedCB func(requestRefs []*isc.RequestRef) gpa.OutMessages,
	requestsNeededCB func(requestIDs []*isc.RequestRef) gpa.OutMessages,
	requestsReceivedCB func(requests []isc.Request) gpa.OutMessages,
) *SubsystemMP {
	return &SubsystemMP{
		proposalInputsReadyCB: proposalInputsReadyCB,
		proposalReceivedCB:    proposalReceivedCB,
		requestsNeededCB:      requestsNeededCB,
		requestsReceivedCB:    requestsReceivedCB,
	}
}

func (sub *SubsystemMP) BaseAliasOutputReceived(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	if sub.BaseAliasOutput != nil {
		return nil
	}
	sub.BaseAliasOutput = baseAliasOutput
	return sub.proposalInputsReadyCB(sub.BaseAliasOutput)
}

func (sub *SubsystemMP) ProposalReceived(requestRefs []*isc.RequestRef) gpa.OutMessages {
	if sub.proposalReceived {
		return nil
	}
	sub.proposalReceived = true
	return sub.proposalReceivedCB(requestRefs)
}

func (sub *SubsystemMP) RequestsNeeded(requestRefs []*isc.RequestRef) gpa.OutMessages {
	if sub.requestsNeeded {
		return nil
	}
	sub.requestsNeeded = true
	return sub.requestsNeededCB(requestRefs)
}

func (sub *SubsystemMP) RequestsReceived(requests []isc.Request) gpa.OutMessages {
	if sub.requestsReceived {
		return nil
	}
	sub.requestsReceived = true
	return sub.requestsReceivedCB(requests)
}
