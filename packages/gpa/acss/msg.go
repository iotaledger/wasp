package acss

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeImplicateRecover gpa.MessageType = iota
	msgTypeVote
	msgTypeWrapped
	msgTypeRBCCEPayload
)

func (a *acssImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeImplicateRecover: func() gpa.Message { return new(msgImplicateRecover) },
		msgTypeVote:             func() gpa.Message { return new(msgVote) },
	}, gpa.Fallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalMessage,
	})
}
