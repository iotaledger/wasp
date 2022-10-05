// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type msgMempoolRequests struct {
	node     gpa.NodeID
	requests []isc.Request
}

var _ gpa.Message = &msgMempoolRequests{}

func NewMsgMempoolRequests(node gpa.NodeID, requests []isc.Request) gpa.Message {
	return &msgMempoolRequests{node: node, requests: requests}
}

func (m *msgMempoolRequests) Recipient() gpa.NodeID {
	return m.node
}

func (m *msgMempoolRequests) SetSender(sender gpa.NodeID) {
	if sender != m.node {
		panic("wrong sender/receiver for a local message")
	}
}

func (m *msgMempoolRequests) MarshalBinary() ([]byte, error) {
	panic("trying to marshal a local message")
}
