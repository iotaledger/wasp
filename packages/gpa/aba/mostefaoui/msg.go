// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeVote gpa.MessageType = iota
	msgTypeDone
	msgTypeWrapped
)

// Implements the gpa.GPA interface.
func (a *abaImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeVote: func() gpa.Message { return new(msgVote) },
		msgTypeDone: func() gpa.Message { return new(msgDone) },
	}, gpa.Fallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalMessage,
	})
}
