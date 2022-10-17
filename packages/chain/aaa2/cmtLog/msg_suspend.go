// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

// That's a local message.
// Sent by the Chain, when another committee has received control over this chain.
type msgSuspend struct {
	gpa.BasicMessage
}

var _ gpa.Message = &msgSuspend{}

func NewMsgSuspend(recipient gpa.NodeID) gpa.Message {
	return &msgSuspend{
		BasicMessage: gpa.NewBasicMessage(recipient),
	}
}

func (m *msgSuspend) MarshalBinary() ([]byte, error) {
	panic("trying to marshal local message")
}
