// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgVoteOK byte = iota
	msgVoteREADY
)

//
// This message is used a vote for the "Bracha-style totality" agreement.
//
type msgVote struct {
	sender    gpa.NodeID
	recipient gpa.NodeID
	kind      byte
}

var _ gpa.Message = &msgVote{}

func (m *msgVote) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgVote) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implemnet.
}

func (m *msgVote) UnmarshalBinary(data []byte) error {
	return nil // TODO: Implemnet.
}
