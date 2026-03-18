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

// UnmarshalPayload implements the gpa.GPA interface.
func (a *ABA) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeVote: func() gpa.MessagePayload { return new(msgVote) },
		msgTypeDone: func() gpa.MessagePayload { return new(msgDone) },
	}, gpa.PayloadFallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalPayload,
	})
}
