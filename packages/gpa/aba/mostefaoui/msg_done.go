// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"encoding"

	"github.com/iotaledger/wasp/packages/gpa"
)

type msgDone struct {
	sender    gpa.NodeID
	recipient gpa.NodeID
	round     int
}

var (
	_ gpa.Message                = &msgDone{}
	_ encoding.BinaryUnmarshaler = &msgDone{}
)

func multicastMsgDone(nodeIDs []gpa.NodeID, me gpa.NodeID, round int) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, n := range nodeIDs {
		if n != me {
			msgs.Add(&msgDone{recipient: n, round: round})
		}
	}
	return msgs
}

func (m *msgDone) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgDone) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgDone) MarshalBinary() ([]byte, error) {
	panic("to be implemented") // TODO: Impl MarshalBinary
}

func (m *msgDone) UnmarshalBinary(data []byte) error {
	panic("to be implemented") // TODO: Impl UnmarshalBinary
}
