package blssig

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeSigShare gpa.MessageType = iota
)

func (cc *ccImpl) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeSigShare: func() gpa.MessagePayload { return new(msgSigShare) },
	})
}
