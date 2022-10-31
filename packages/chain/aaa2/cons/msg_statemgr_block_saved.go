// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgStateMgrBlockSaved struct {
	node gpa.NodeID
}

var _ gpa.Message = &msgStateMgrBlockSaved{}

func NewMsgStateMgrBlockSaved(node gpa.NodeID) gpa.Message {
	return &msgStateMgrBlockSaved{node: node}
}

func (m *msgStateMgrBlockSaved) Recipient() gpa.NodeID {
	return m.node
}

func (m *msgStateMgrBlockSaved) SetSender(sender gpa.NodeID) {
	if sender != m.node {
		panic("wrong sender/receiver for a local message")
	}
}

func (m *msgStateMgrBlockSaved) MarshalBinary() ([]byte, error) {
	panic("trying to marshal a local message")
}
