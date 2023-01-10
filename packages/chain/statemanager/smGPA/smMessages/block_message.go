package smMessages

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type BlockMessage struct {
	gpa.BasicMessage
	block state.Block
}

var _ gpa.Message = &BlockMessage{}

func NewBlockMessage(block state.Block, to gpa.NodeID) *BlockMessage {
	return &BlockMessage{
		BasicMessage: gpa.NewBasicMessage(to),
		block:        block,
	}
}

func NewEmptyBlockMessage() *BlockMessage { // `UnmarshalBinary` must be called afterwards
	return NewBlockMessage(nil, gpa.NodeID{})
}

func (bmT *BlockMessage) MarshalBinary() (data []byte, err error) {
	return append([]byte{MsgTypeBlockMessage}, bmT.block.Bytes()...), nil
}

func (bmT *BlockMessage) UnmarshalBinary(data []byte) error {
	if data[0] != MsgTypeBlockMessage {
		return fmt.Errorf("error creating block message from bytes: wrong message type %v", data[0])
	}
	var err error
	bmT.block, err = state.BlockFromBytes(data[1:])
	return err
}

func (bmT *BlockMessage) GetBlock() state.Block {
	return bmT.block
}
