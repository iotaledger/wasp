// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

//
// An event to self.
type msgAgreementResult struct {
	me      gpa.NodeID
	indexes []int
}

var _ gpa.Message = &msgAgreementResult{}

func NewMsgAgreementResult(me gpa.NodeID, indexes []int) gpa.Message {
	return &msgAgreementResult{me: me, indexes: indexes}
}

func (m *msgAgreementResult) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgAgreementResult) SetSender(sender gpa.NodeID) {
	// Don't care the sender.
}

func (m *msgAgreementResult) MarshalBinary() ([]byte, error) {
	panic("should be not needed")
}
