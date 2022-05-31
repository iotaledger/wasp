// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgImplicateKind byte

const (
	msgImplicateRecoverKindIMPLICATE msgImplicateKind = iota
	msgImplicateRecoverKindRECOVER
)

//
// The <IMPLICATE, i, skᵢ> and <RECOVER, i, skᵢ> messages.
//
type msgImplicateRecover struct {
	sender    gpa.NodeID
	recipient gpa.NodeID
	kind      msgImplicateKind
	i         int
	data      []byte // Either implication or the recovered secret.
}

var _ gpa.Message = &msgImplicateRecover{}

func (m *msgImplicateRecover) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgImplicateRecover) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgImplicateRecover) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implemnet.
}

func (m *msgImplicateRecover) UnmarshalBinary(data []byte) error {
	return nil // TODO: Implemnet.
}
