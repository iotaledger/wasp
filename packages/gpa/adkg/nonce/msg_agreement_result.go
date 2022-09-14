// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
)

// An event to self.
type msgAgreementResult struct {
	me        gpa.NodeID
	proposals map[gpa.NodeID][]int
}

var _ gpa.Message = &msgAgreementResult{}

func NewMsgAgreementResult(me gpa.NodeID, proposals map[gpa.NodeID][]int) gpa.Message {
	return &msgAgreementResult{me: me, proposals: proposals}
}

func (m *msgAgreementResult) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgAgreementResult) SetSender(sender gpa.NodeID) {
	if sender != "" && sender != m.me {
		panic(xerrors.Errorf("local events cannot be sent from other nodes"))
	}
}

func (m *msgAgreementResult) MarshalBinary() ([]byte, error) {
	panic("msgAgreementResult is a local message, should not be marshaled")
}
