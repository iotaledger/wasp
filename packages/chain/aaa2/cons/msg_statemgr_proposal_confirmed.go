// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type msgStateMgrProposalConfirmed struct {
	node            gpa.NodeID
	baseAliasOutput *isc.AliasOutputWithID
}

var _ gpa.Message = &msgStateMgrProposalConfirmed{}

func (m *msgStateMgrProposalConfirmed) Recipient() gpa.NodeID {
	return m.node
}

func (m *msgStateMgrProposalConfirmed) SetSender(sender gpa.NodeID) {
	if sender != m.node {
		panic("wrong sender/receiver for a local message")
	}
}

func (m *msgStateMgrProposalConfirmed) MarshalBinary() ([]byte, error) {
	panic("trying to marshal a local message")
}
