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

func NewEmptyGetBlockMessage() *GetBlockMessage { // `UnmarshalBinary` must be called afterwards
	return NewGetBlockMessage(&state.L1Commitment{}, gpa.NodeID{})
}

func (msg *GetBlockMessage) GetL1Commitment() *state.L1Commitment {
	return msg.commitment
}

func (msg *GetBlockMessage) MarshalBinary() (data []byte, err error) {
	return rwutil.MarshalBinary(msg)
}

func (msg *GetBlockMessage) UnmarshalBinary(data []byte) error {
	return rwutil.UnmarshalBinary(data, msg)
}

func (msg *GetBlockMessage) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	MsgTypeGetBlockMessage.ReadAndVerify(rr)
	data := rr.ReadBytes()
	if rr.Err == nil {
		msg.commitment, rr.Err = state.L1CommitmentFromBytes(data)
	}
	return rr.Err
}

func (msg *GetBlockMessage) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	MsgTypeGetBlockMessage.Write(ww)
	ww.WriteBytes(msg.commitment.Bytes())
	return ww.Err
}
