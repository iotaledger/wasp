package sm_messages

import (
	"testing"

	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMarshalUnmarshalGetBlockMessage(t *testing.T) {
	blocks := sm_gpa_utils.NewBlockFactory(t).GetBlocks(4, 1)
	for i := range blocks {
		// note that sender/receiver node IDs are transient
		// so don't use a random non-null node id here
		commitment := blocks[i].L1Commitment()
		msg := NewGetBlockMessage(commitment, gpa.NodeID{})
		rwutil.ReadWriteTest(t, msg, NewEmptyGetBlockMessage())
	}
}

func TestGetBlockMessageSerialization(t *testing.T) {
	msg := &GetBlockMessage{
		gpa.BasicMessage{},
		state.PseudoRandL1Commitment(),
	}

	rwutil.ReadWriteTest(t, msg, new(GetBlockMessage))
}
