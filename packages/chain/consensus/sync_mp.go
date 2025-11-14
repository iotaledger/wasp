// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type SyncMempool struct {
	c                  *Consensus
	baseAnchor         *isc.StateAnchor
	baseAnchorReceived bool
	proposalReceived   bool
	requestsNeeded     bool
	requestsReceived   bool
}

func NewSyncMempool(
	c *Consensus,
) *SyncMempool {
	return &SyncMempool{
		c: c,
	}
}

func (s *SyncMempool) BaseAnchorReceived(baseAnchor *isc.StateAnchor) []gpa.PayloadOut {
	if s.baseAnchorReceived {
		return nil
	}
	s.baseAnchor = baseAnchor
	s.baseAnchorReceived = true
	return s.c.uponMempoolProposalInputsReady(s.baseAnchor)
}

func (s *SyncMempool) ProposalReceived(requestRefs []*isc.RequestRef) []gpa.PayloadOut {
	if s.proposalReceived {
		return nil
	}
	s.proposalReceived = true
	return s.c.uponMempoolProposalReceived(requestRefs)
}

func (s *SyncMempool) RequestsNeeded(requestRefs []*isc.RequestRef) []gpa.PayloadOut {
	if s.requestsNeeded {
		return nil
	}
	s.requestsNeeded = true
	return s.c.uponMempoolRequestsNeeded(requestRefs)
}

func (s *SyncMempool) RequestsReceived(requests []isc.Request) []gpa.PayloadOut {
	if s.requestsReceived {
		return nil
	}
	s.requestsReceived = true
	return s.c.uponMempoolRequestsReceived(requests)
}

// String tries to provide useful human-readable compact status.
func (s *SyncMempool) String() string {
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
