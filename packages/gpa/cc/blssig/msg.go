package blssig

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeSigShare gpa.MessageType = iota
)

func (cc *ccImpl) MarshalMessage(msg gpa.Message) ([]byte, error) {
	switch msg := msg.(type) {
	case *msgSigShare:
		return gpa.MarshalMessage(msgTypeSigShare, msg)
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
}

func (cc *ccImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeSigShare: func() gpa.Message { return new(msgSigShare) },
	})
}
