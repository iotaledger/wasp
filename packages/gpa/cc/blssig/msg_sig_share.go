// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig

import "github.com/iotaledger/wasp/packages/gpa"

type msgSigShare struct {
	recipient gpa.NodeID
	sender    gpa.NodeID
	sigShare  []byte
}

var _ gpa.Message = &msgSigShare{}

func (m *msgSigShare) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgSigShare) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgSigShare) MarshalBinary() ([]byte, error) {
	panic("not implemented yet") // TODO: ...
}
