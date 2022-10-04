// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type msgMempoolProposal struct {
	node        gpa.NodeID
	requestRefs []*isc.RequestRef
}

var _ gpa.Message = &msgMempoolProposal{}

func (m *msgMempoolProposal) Recipient() gpa.NodeID {
	return m.node
}

func (m *msgMempoolProposal) SetSender(sender gpa.NodeID) {
	if sender != m.node {
		panic("wrong sender/receiver for a local message")
	}
}

func (m *msgMempoolProposal) MarshalBinary() ([]byte, error) {
	panic("trying to marshal a local message")
}
