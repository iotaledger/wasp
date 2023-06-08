// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgVoteKind byte

const (
	msgVoteOK msgVoteKind = iota
	msgVoteREADY
)

// This message is used a vote for the "Bracha-style totality" agreement.
type msgVote struct {
	sender    gpa.NodeID
	recipient gpa.NodeID
	kind      msgVoteKind
}

var _ gpa.Message = &msgVote{}

func (m *msgVote) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgVote) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgVote) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := rwutil.WriteByte(w, msgTypeVote); err != nil {
		return nil, err
	}
	if err := rwutil.WriteByte(w, byte(m.kind)); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgVote) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	t, err := rwutil.ReadByte(r)
	if err != nil {
		return err
	}
	if t != msgTypeVote {
		return fmt.Errorf("unexpected msgType: %v in acss.msgVote", t)
	}
	k, err := rwutil.ReadByte(r)
	if err != nil {
		return err
	}
	m.kind = msgVoteKind(k)
	return nil
}
