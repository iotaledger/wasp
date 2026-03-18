// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const msgType gpa.MessageType = iota

// The type for message kinds (only one of these in this case).
type msgBrachaType byte

const (
	msgBrachaTypePropose msgBrachaType = iota
	msgBrachaTypeEcho
	msgBrachaTypeReady
)

func (msgBrachaType msgBrachaType) String() string {
	switch msgBrachaType {
	case msgBrachaTypePropose:
		return "Propose"
	case msgBrachaTypeEcho:
		return "Echo"
	case msgBrachaTypeReady:
		return "Ready"
	default:
		return "UnknownBrachaMsgType"
	}
}

type msgBracha struct {
	brachaType msgBrachaType `bcs:"export"` // Type
	value      []byte        `bcs:"export"` // Value
}

var _ gpa.MessagePayload = new(msgBracha)

func (msg *msgBracha) MsgType() gpa.MessageType {
	return msgType
}

func (msg *msgBracha) String() string {
	return fmt.Sprintf("Bracha.%s(%q)", msg.brachaType.String(), msg.value)
}
