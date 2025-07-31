package messages

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa/utils"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
)

func TestBlockMessageSerialization(t *testing.T) {
	blocks := utils.NewBlockFactory(t).GetBlocks(4, 1)
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
		statetest.RandomBlock(),
	}

	bcs.TestCodec(t, msg)

	bcs.TestCodecAndHash(t, &BlockMessage{
		gpa.BasicMessage{},
		statetest.TestBlock,
	}, "453dabc9e5e2")
}
