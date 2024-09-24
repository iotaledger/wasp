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

func (a *acssImpl) MarshalMessage(msg gpa.Message) ([]byte, error) {
	switch msg := msg.(type) {
	case *msgImplicateRecover:
		return gpa.MarshalMessage(msgTypeImplicateRecover, msg)
	case *msgVote:
		return gpa.MarshalMessage(msgTypeVote, msg)
	default:
		return gpa.MarshalWrappedMessage(msgTypeWrapped, msg, a.msgWrapper)
	}
}

func (a *acssImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeImplicateRecover: func() gpa.Message { return new(msgImplicateRecover) },
		msgTypeVote:             func() gpa.Message { return new(msgVote) },
	}, gpa.Fallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalMessage,
	})
}
