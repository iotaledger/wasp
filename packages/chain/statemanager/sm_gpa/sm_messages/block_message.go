package sm_messages

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
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

func NewEmptyBlockMessage() *BlockMessage { // `UnmarshalBinary` must be called afterwards
	return NewBlockMessage(nil, gpa.NodeID{})
}

func (msg *BlockMessage) GetBlock() state.Block {
	return msg.block
}

func (msg *BlockMessage) MarshalBinary() (data []byte, err error) {
	return rwutil.MarshalBinary(msg)
}

func (msg *BlockMessage) UnmarshalBinary(data []byte) error {
	return rwutil.UnmarshalBinary(data, msg)
}

func (msg *BlockMessage) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	MsgTypeBlockMessage.ReadAndVerify(rr)
	data := rr.ReadBytes()
	if rr.Err == nil {
		msg.block, rr.Err = state.BlockFromBytes(data)
	}
	return rr.Err
}

func (msg *BlockMessage) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	MsgTypeBlockMessage.Write(ww)
	ww.WriteBytes(msg.block.Bytes())
	return ww.Err
}
