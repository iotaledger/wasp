package chainmanager

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/state"
)

// This message is used to inform access nodes on new blocks
// produced so that they can update their active state faster.
type msgBlockProduced struct {
	gpa.BasicMessage
	tx    *iotasigner.SignedTransaction `bcs:"export"`
	block state.Block                   `bcs:"export"`
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
		"{chainMgr.msgBlockProduced, stateIndex=%v, l1Commitment=%v, txDigest=%s}",
		msg.block.StateIndex(), msg.block.L1Commitment(), lo.Must(msg.tx.Digest()),
	)
}
