// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"go.dedis.ch/kyber/v3"
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
	recipient gpa.NodeID
	kind      msgImplicateKind
	i         int
	sk        kyber.Scalar
}

var _ gpa.Message = &msgVote{}

func (m *msgImplicateRecover) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgImplicateRecover) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implemnet.
}

func (m *msgImplicateRecover) UnmarshalBinary(data []byte) error {
	return nil // TODO: Implemnet.
}
