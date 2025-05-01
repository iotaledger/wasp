// Package messages defines message types used in the state manager's communication protocol.
package messages

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type BlockMessage struct {
	gpa.BasicMessage
	block state.Block
}

var _ gpa.Message = new(BlockMessage)

func NewBlockMessage(block state.Block, to gpa.NodeID) *BlockMessage {
	return &BlockMessage{
		BasicMessage: gpa.NewBasicMessage(to),
		block:        block,
	}
}

func NewEmptyBlockMessage() *BlockMessage {
	return NewBlockMessage(nil, gpa.NodeID{})
}

func (msg *BlockMessage) GetBlock() state.Block {
	return msg.block
}

func (msg *BlockMessage) UnmarshalBCS(d *bcs.Decoder) error {
	msg.block = state.NewBlock()
	d.Decode(msg.block)

	return nil
}

func (msg *BlockMessage) MarshalBCS(e *bcs.Encoder) error {
	e.Encode(msg.block)
	return nil
}

func (msg *BlockMessage) MsgType() gpa.MessageType {
	return MsgTypeBlockMessage
}
