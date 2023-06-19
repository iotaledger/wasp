package chainmanager

import (
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// This message is used to inform access nodes on new blocks
// produced so that they can update their active state faster.
type msgBlockProduced struct {
	gpa.BasicMessage
	tx    *iotago.Transaction
	block state.Block
}

var _ gpa.Message = new(msgBlockProduced)

func NewMsgBlockProduced(recipient gpa.NodeID, tx *iotago.Transaction, block state.Block) gpa.Message {
	return &msgBlockProduced{
		BasicMessage: gpa.NewBasicMessage(recipient),
		tx:           tx,
		block:        block,
	}
}

func (msg *msgBlockProduced) String() string {
	txID, err := msg.tx.ID()
	if err != nil {
		panic(fmt.Errorf("cannot extract TX ID: %w", err))
	}
	return fmt.Sprintf(
		"{chainMgr.msgBlockProduced, stateIndex=%v, l1Commitment=%v, tx.ID=%v}",
		msg.block.StateIndex(), msg.block.L1Commitment(), txID.ToHex(),
	)
}

func (msg *msgBlockProduced) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeBlockProduced.ReadAndVerify(rr)
	msg.tx = new(iotago.Transaction)
	rr.ReadSerialized(msg.tx)
	msg.block = state.NewBlock()
	rr.Read(msg.block)
	return rr.Err
}

func (msg *msgBlockProduced) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeBlockProduced.Write(ww)
	ww.WriteSerialized(msg.tx)
	ww.Write(msg.block)
	return ww.Err
}
