// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type SyncMempool interface {
	BaseAnchorReceived(baseAnchor *isc.StateAnchor) gpa.OutMessages
	ProposalReceived(requestRefs []*isc.RequestRef) gpa.OutMessages
	RequestsNeeded(requestRefs []*isc.RequestRef) gpa.OutMessages
	RequestsReceived(requests []isc.Request) gpa.OutMessages
	String() string
}

type syncMempoolImpl struct {
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

func NewSyncMempool(
	proposalInputsReadyCB func(baseAnchor *isc.StateAnchor) gpa.OutMessages,
	proposalReceivedCB func(requestRefs []*isc.RequestRef) gpa.OutMessages,
	requestsNeededCB func(requestIDs []*isc.RequestRef) gpa.OutMessages,
	requestsReceivedCB func(requests []isc.Request) gpa.OutMessages,
) SyncMempool {
	return &syncMempoolImpl{
		proposalInputsReadyCB: proposalInputsReadyCB,
		proposalReceivedCB:    proposalReceivedCB,
		requestsNeededCB:      requestsNeededCB,
		requestsReceivedCB:    requestsReceivedCB,
	}
}

func (s *syncMempoolImpl) BaseAnchorReceived(baseAnchor *isc.StateAnchor) gpa.OutMessages {
	if s.baseAnchorReceived {
		return nil
	}
	s.baseAnchor = baseAnchor
	s.baseAnchorReceived = true
	return s.proposalInputsReadyCB(s.baseAnchor)
}

func (s *syncMempoolImpl) ProposalReceived(requestRefs []*isc.RequestRef) gpa.OutMessages {
	if s.proposalReceived {
		return nil
	}
	s.proposalReceived = true
	return s.proposalReceivedCB(requestRefs)
}

func (s *syncMempoolImpl) RequestsNeeded(requestRefs []*isc.RequestRef) gpa.OutMessages {
	if s.requestsNeeded {
		return nil
	}
	s.requestsNeeded = true
	return s.requestsNeededCB(requestRefs)
}

func (s *syncMempoolImpl) RequestsReceived(requests []isc.Request) gpa.OutMessages {
	if s.requestsReceived {
		return nil
	}
	s.requestsReceived = true
	return s.requestsReceivedCB(requests)
}

// Try to provide useful human-readable compact status.
func (s *syncMempoolImpl) String() string {
	str := "MP"
	if s.proposalReceived && s.requestsReceived {
		return str + statusStrOK
	}
	if s.proposalReceived {
		str += "/proposal=OK"
	} else if !s.baseAnchorReceived {
		str += "/proposal=WAIT[BaseAnchor]"
	} else {
		str += "/proposal=WAIT[RespFromMemPool]"
	}
	if s.requestsReceived {
		str += "/requests=OK"
	} else if !s.requestsNeeded {
		str += "/requests=WAIT[AcsDecision]"
	} else {
		str += "/requests=WAIT[RespFromMemPool]"
	}
	return str
}
