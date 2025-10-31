package acss

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeImplicateRecover gpa.MessageType = iota
	msgTypeVote
	msgTypeWrapped
	msgTypeRBCCEPayload
)

func (a *acssImpl) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeImplicateRecover: func() gpa.MessagePayload { return new(msgImplicateRecover) },
		msgTypeVote:             func() gpa.MessagePayload { return new(msgVote) },
	}, gpa.PayloadFallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalPayload,
	})
}
