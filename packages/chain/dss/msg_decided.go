// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"golang.org/x/xerrors"
)

type msgDecided struct {
	me                    gpa.NodeID
	decidedIndexProposals map[gpa.NodeID][]int
	messageToSign         []byte
}

var _ gpa.Message = &msgDecided{}

func NewMsgDecided(me gpa.NodeID, decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) gpa.Message {
	return &msgDecided{me, decidedIndexProposals, messageToSign}
}

func (m *msgDecided) Recipient() gpa.NodeID {
	return m.me // Local message.
}

func (m *msgDecided) SetSender(sender gpa.NodeID) {
	// Local message, don't care the sender.
}

func (m *msgDecided) MarshalBinary() ([]byte, error) {
	panic(xerrors.Errorf("local message, cannot be marshaled"))
}
