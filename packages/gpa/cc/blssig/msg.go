package blssig

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

const (
	msgTypeSigShare gpa.MessageType = iota
)

func (cc *ccImpl) MarshalMessage(msg gpa.Message) ([]byte, error) {
	var b bytes.Buffer
	e := bcs.NewEncoder(&b)

	switch msg := msg.(type) {
	case *msgSigShare:
		e.Encode(msgTypeSigShare)
		e.Encode(msg)
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}

	return b.Bytes(), e.Err()
}

func (cc *ccImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeSigShare: func() gpa.Message { return new(msgSigShare) },
	})
}
