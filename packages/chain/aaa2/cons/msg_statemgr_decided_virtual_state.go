// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/state"
)

type msgStateMgrDecidedVirtualState struct {
	node               gpa.NodeID
	stateBaseline      coreutil.StateBaseline
	virtualStateAccess state.VirtualStateAccess
}

var _ gpa.Message = &msgStateMgrDecidedVirtualState{}

func NewMsgStateMgrDecidedVirtualState(
	node gpa.NodeID,
	stateBaseline coreutil.StateBaseline,
	virtualStateAccess state.VirtualStateAccess,
) gpa.Message {
	return &msgStateMgrDecidedVirtualState{node, stateBaseline, virtualStateAccess}
}

func (m *msgStateMgrDecidedVirtualState) Recipient() gpa.NodeID {
	return m.node
}

func (m *msgStateMgrDecidedVirtualState) SetSender(sender gpa.NodeID) {
	if sender != m.node {
		panic("wrong sender/receiver for a local message")
	}
}

func (m *msgStateMgrDecidedVirtualState) MarshalBinary() ([]byte, error) {
	panic("trying to marshal a local message")
}
