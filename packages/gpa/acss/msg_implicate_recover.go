// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgImplicateKind byte

const (
	msgImplicateRecoverKindIMPLICATE msgImplicateKind = iota
	msgImplicateRecoverKindRECOVER
)

// The <IMPLICATE, i, skᵢ> and <RECOVER, i, skᵢ> messages.
type msgImplicateRecover struct {
	kind msgImplicateKind `bcs:"export"`
	i    int              `bcs:"export,type=u16"`
	data []byte           `bcs:"export"` // Either implication or the recovered secret.
}

var _ gpa.MessagePayload = new(msgImplicateRecover)

func (msg *msgImplicateRecover) MsgType() gpa.MessageType {
	return msgTypeImplicateRecover
}
