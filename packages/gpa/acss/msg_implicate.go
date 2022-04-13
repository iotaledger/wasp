// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"go.dedis.ch/kyber/v3"
)

//
// The <IMPLICATE, i, skᵢ> message.
//
type msgImplicate struct {
	recipient gpa.NodeID
	i         int
	sk        kyber.Scalar
}

var _ gpa.Message = &msgVote{}

func (m *msgImplicate) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgImplicate) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implemnet.
}

func (m *msgImplicate) UnmarshalBinary(data []byte) error {
	return nil // TODO: Implemnet.
}
