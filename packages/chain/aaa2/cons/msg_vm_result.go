// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/vm"
)

type msgVMResult struct {
	node gpa.NodeID // This node.
	task *vm.VMTask // With results set.
}

var _ gpa.Message = &msgVMResult{}

func NewMsgVMResult(node gpa.NodeID, task *vm.VMTask) gpa.Message {
	return &msgVMResult{node: node, task: task}
}

func (m *msgVMResult) Recipient() gpa.NodeID {
	return m.node
}

func (m *msgVMResult) SetSender(sender gpa.NodeID) {
	if sender != m.node {
		panic("wrong sender/receiver for a local message")
	}
}

func (m *msgVMResult) MarshalBinary() ([]byte, error) {
	panic("trying to marshal a local message")
}
