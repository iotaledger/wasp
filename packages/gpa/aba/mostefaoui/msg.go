// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeVote gpa.MessageType = iota
	msgTypeDone
	msgTypeWrapped
)

func (a *abaImpl) MarshalMessage(msg gpa.Message) ([]byte, error) {
	switch msg := msg.(type) {
	case *msgVote:
		return gpa.MarshalMessage(msgTypeVote, msg)
	case *msgDone:
		return gpa.MarshalMessage(msgTypeDone, msg)
	default:
		return gpa.MarshalWrappedMessage(msgTypeWrapped, msg, a.msgWrapper)
	}
}

// Implements the gpa.GPA interface.
func (a *abaImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeVote: func() gpa.Message { return new(msgVote) },
		msgTypeDone: func() gpa.Message { return new(msgDone) },
	}, gpa.Fallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalMessage,
	})
}
