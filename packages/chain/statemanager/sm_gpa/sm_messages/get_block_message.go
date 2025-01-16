package sm_messages

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type GetBlockMessage struct {
	gpa.BasicMessage
	commitment *state.L1Commitment `bcs:"export"`
}

var _ gpa.Message = new(GetBlockMessage)

func NewGetBlockMessage(commitment *state.L1Commitment, to gpa.NodeID) *GetBlockMessage {
	return &GetBlockMessage{
		BasicMessage: gpa.NewBasicMessage(to),
		commitment:   commitment,
	}
}

func NewEmptyGetBlockMessage() *GetBlockMessage {
	return NewGetBlockMessage(&state.L1Commitment{}, gpa.NodeID{})
}

func (msg *GetBlockMessage) GetL1Commitment() *state.L1Commitment {
	return msg.commitment
}

func (msg *GetBlockMessage) MsgType() gpa.MessageType {
	return MsgTypeGetBlockMessage
}
