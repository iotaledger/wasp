// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type SyncMP interface {
	BaseAnchorReceived(baseAnchor *isc.StateAnchor) gpa.OutMessages
	ProposalReceived(requestRefs []*isc.RequestRef) gpa.OutMessages
	RequestsNeeded(requestRefs []*isc.RequestRef) gpa.OutMessages
	RequestsReceived(requests []isc.Request) gpa.OutMessages
	String() string
}

type syncMPImpl struct {
	baseAnchor            *isc.StateAnchor
	baseAnchorReceived    bool
	proposalInputsReadyCB func(baseAnchor *isc.StateAnchor) gpa.OutMessages
	proposalReceived      bool
	proposalReceivedCB    func(requestRefs []*isc.RequestRef) gpa.OutMessages
	requestsNeeded        bool
	requestsNeededCB      func(requestIDs []*isc.RequestRef) gpa.OutMessages
	requestsReceived      bool
	requestsReceivedCB    func(requests []isc.Request) gpa.OutMessages
}

func NewSyncMP(
	proposalInputsReadyCB func(baseAnchor *isc.StateAnchor) gpa.OutMessages,
	proposalReceivedCB func(requestRefs []*isc.RequestRef) gpa.OutMessages,
	requestsNeededCB func(requestIDs []*isc.RequestRef) gpa.OutMessages,
	requestsReceivedCB func(requests []isc.Request) gpa.OutMessages,
) SyncMP {
	return &syncMPImpl{
		proposalInputsReadyCB: proposalInputsReadyCB,
		proposalReceivedCB:    proposalReceivedCB,
		requestsNeededCB:      requestsNeededCB,
		requestsReceivedCB:    requestsReceivedCB,
	}
}

func (sub *syncMPImpl) BaseAnchorReceived(baseAnchor *isc.StateAnchor) gpa.OutMessages {
	if sub.baseAnchorReceived {
		return nil
	}
	sub.baseAnchor = baseAnchor
	sub.baseAnchorReceived = true
	return sub.proposalInputsReadyCB(sub.baseAnchor)
}

func (sub *syncMPImpl) ProposalReceived(requestRefs []*isc.RequestRef) gpa.OutMessages {
	if sub.proposalReceived {
		return nil
	}
	sub.proposalReceived = true
	return sub.proposalReceivedCB(requestRefs)
}

func (sub *syncMPImpl) RequestsNeeded(requestRefs []*isc.RequestRef) gpa.OutMessages {
	if sub.requestsNeeded {
		return nil
	}
	sub.requestsNeeded = true
	return sub.requestsNeededCB(requestRefs)
}

func (sub *syncMPImpl) RequestsReceived(requests []isc.Request) gpa.OutMessages {
	if sub.requestsReceived {
		return nil
	}
	sub.requestsReceived = true
	return sub.requestsReceivedCB(requests)
}

// Try to provide useful human-readable compact status.
func (sub *syncMPImpl) String() string {
	str := "MP"
	if sub.proposalReceived && sub.requestsReceived {
		return str + statusStrOK
	}
	if sub.proposalReceived {
		str += "/proposal=OK"
	} else if !sub.baseAnchorReceived {
		str += "/proposal=WAIT[BaseAnchor]"
	} else {
		str += "/proposal=WAIT[RespFromMemPool]"
	}
	if sub.requestsReceived {
		str += "/requests=OK"
	} else if !sub.requestsNeeded {
		str += "/requests=WAIT[AcsDecision]"
	} else {
		str += "/requests=WAIT[RespFromMemPool]"
	}
	return str
}
