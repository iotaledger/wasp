// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"fmt"
)

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

type MsgBracha struct {
	brachaType msgBrachaType `bcs:"export"` // Type
	value      []byte        `bcs:"export"` // Value
}

func (msg *MsgBracha) String() string {
	return fmt.Sprintf("Bracha.%s(%q)", msg.brachaType.String(), msg.value)
}
