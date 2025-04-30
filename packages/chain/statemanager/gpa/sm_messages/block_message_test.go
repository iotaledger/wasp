package sm_messages

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

func TestBlockMessageSerialization(t *testing.T) {
	blocks := sm_gpa_utils.NewBlockFactory(t).GetBlocks(4, 1)
	for i := range blocks {
		// note that sender/receiver node IDs are transient
		// so don't use a random non-null node id here
		msg := NewBlockMessage(blocks[i], gpa.NodeID{})
		bcs.TestCodec(t, msg)
	}
}

func TestSerializationBlockMessage(t *testing.T) {
	msg := &BlockMessage{
		gpa.BasicMessage{},
		state.RandomBlock(),
	}

	bcs.TestCodec(t, msg)
}
