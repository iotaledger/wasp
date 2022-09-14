// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"

	"github.com/iotaledger/wasp/packages/gpa"
)

// An event to self.
type msgACSSOutput struct {
	me       gpa.NodeID
	index    int
	priShare *share.PriShare
	commits  []kyber.Point
}

var _ gpa.Message = &msgACSSOutput{}

func (m *msgACSSOutput) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgACSSOutput) SetSender(sender gpa.NodeID) {
	// Don't care the sender.
}

func (m *msgACSSOutput) MarshalBinary() ([]byte, error) {
	panic("msgACSSOutput is a local message, should not be marshaled")
}
