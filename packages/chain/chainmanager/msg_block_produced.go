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
	tx    *iotasigner.SignedTransaction `bcs:"export"`
	block state.Block                   `bcs:"export"`
}

func NewMsgBlockProduced(recipient gpa.NodeID, tx *iotasigner.SignedTransaction, block state.Block) gpa.MessageOut {
	return gpa.NewMessageOut(recipient, msgBlockProduced{
		tx:    tx,
		block: block,
	})
}

func (msg *msgBlockProduced) String() string {
	return fmt.Sprintf(
		"{chainMgr.msgBlockProduced, stateIndex=%v, l1Commitment=%v, txDigest=%s}",
		msg.block.StateIndex(), msg.block.L1Commitment(), lo.Must(msg.tx.Digest()),
	)
}
