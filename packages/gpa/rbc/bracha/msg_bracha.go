// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgBrachaType byte

const (
	// The type for message kinds (only one of these in this case).
	msgType gpa.MessageType = iota

	msgBrachaTypePropose msgBrachaType = iota
	msgBrachaTypeEcho
	msgBrachaTypeReady
)

type msgBracha struct {
	gpa.BasicMessage
	brachaType msgBrachaType `bcs:"export"` // Type
	value      []byte        `bcs:"export"` // Value
}

var _ gpa.Message = new(msgBracha)

func (msg *msgBracha) MsgType() gpa.MessageType {
	return msgType
}
