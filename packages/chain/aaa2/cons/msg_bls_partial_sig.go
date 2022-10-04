// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
)

type msgBLSPartialSig struct {
	blsSuite   suites.Suite
	sender     gpa.NodeID
	recipient  gpa.NodeID
	partialSig []byte
}

var _ gpa.Message = &msgBLSPartialSig{}

func newMsgBLSPartialSig(blsSuite suites.Suite, recipient gpa.NodeID, partialSig []byte) *msgBLSPartialSig {
	return &msgBLSPartialSig{blsSuite: blsSuite, recipient: recipient, partialSig: partialSig}
}

func (m *msgBLSPartialSig) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgBLSPartialSig) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgBLSPartialSig) MarshalBinary() ([]byte, error) {
	panic("to be implemented")
}

func (m *msgBLSPartialSig) UnmarshalBinary(data []byte) error {
	panic("to be implemented")
}
