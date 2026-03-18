// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgSigShare struct {
	sigShare []byte `bcs:"export"`
}

var _ gpa.MessagePayload = new(msgSigShare)

func (msg *msgSigShare) MsgType() gpa.MessageType {
	return msgTypeSigShare
}
