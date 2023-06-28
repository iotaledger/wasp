package sm_messages

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type GetBlockMessage struct {
	gpa.BasicMessage
	commitment *state.L1Commitment
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

func (msg *GetBlockMessage) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	MsgTypeGetBlockMessage.ReadAndVerify(rr)
	msg.commitment = new(state.L1Commitment)
	rr.Read(msg.commitment)
	return rr.Err
}

func (msg *GetBlockMessage) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	MsgTypeGetBlockMessage.Write(ww)
	ww.Write(msg.commitment)
	return ww.Err
}
