package messages

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type GetBlockMessage struct {
	commitment *state.L1Commitment `bcs:"export"`
}

var _ gpa.MessagePayload = new(GetBlockMessage)

func NewGetBlockMessage(commitment *state.L1Commitment) *GetBlockMessage {
	return &GetBlockMessage{
		commitment: commitment,
	}
}

func NewEmptyGetBlockMessage() *GetBlockMessage {
	return NewGetBlockMessage(&state.L1Commitment{})
}

func (msg *GetBlockMessage) GetL1Commitment() *state.L1Commitment {
	return msg.commitment
}

func (msg *GetBlockMessage) MsgType() gpa.MessageType {
	return MsgTypeGetBlockMessage
}
