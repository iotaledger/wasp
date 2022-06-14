// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"go.dedis.ch/kyber/v3/sign/dss"
)

type msgPartialSig struct {
	sender     gpa.NodeID
	recipient  gpa.NodeID
	partialSig *dss.PartialSig
}

var _ gpa.Message = &msgPartialSig{}

func (m *msgPartialSig) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgPartialSig) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgPartialSig) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}
