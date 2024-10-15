package chainmanager

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/samber/lo"
)

// This message is used to inform access nodes on new blocks
// produced so that they can update their active state faster.
type msgBlockProduced struct {
	gpa.BasicMessage
	tx    *iotasigner.SignedTransaction
	block state.Block
}

var _ gpa.Message = new(msgBlockProduced)

func NewMsgBlockProduced(recipient gpa.NodeID, tx *iotasigner.SignedTransaction, block state.Block) gpa.Message {
	return &msgBlockProduced{
		BasicMessage: gpa.NewBasicMessage(recipient),
		tx:           tx,
		block:        block,
	}
}

func (msg *msgBlockProduced) MsgType() gpa.MessageType {
	return msgTypeBlockProduced
}

func (msg *msgBlockProduced) String() string {
	return fmt.Sprintf(
		"{chainMgr.msgBlockProduced, stateIndex=%v, l1Commitment=%v, txHash=%v}",
		msg.block.StateIndex(), msg.block.L1Commitment(), lo.Must(msg.tx.Hash()).Hex(),
	)
}

func (msg *msgBlockProduced) Read(r io.Reader) error {
	panic("implement msgBlockProduced.Read") // TODO: ..
	// rr := rwutil.NewReader(r)
	// msgTypeBlockProduced.ReadAndVerify(rr)
	// msg.tx = new(iotasigner.SignedTransaction)
	// rr.ReadSerialized(msg.tx)
	// msg.block = state.NewBlock()
	// rr.Read(msg.block)
	// return rr.Err
}

func (msg *msgBlockProduced) Write(w io.Writer) error {
	panic("implement msgBlockProduced.Write") // TODO: ..
	// ww := rwutil.NewWriter(w)
	// msgTypeBlockProduced.Write(ww)
	// ww.WriteSerialized(msg.tx)
	// ww.Write(msg.block)
	// return ww.Err
}
