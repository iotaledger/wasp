package messages

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/chain/statemanager/gpa/utils"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
)

func TestMarshalUnmarshalGetBlockMessage(t *testing.T) {
	blocks := utils.NewBlockFactory(t).GetBlocks(4, 1)
	for i := range blocks {
		// note that sender/receiver node IDs are transient
		// so don't use a random non-null node id here
		commitment := blocks[i].L1Commitment()
		msg := NewGetBlockMessage(commitment)
		bcs.TestCodec(t, msg)
	}
}

func TestGetBlockMessageSerialization(t *testing.T) {
	msg := &GetBlockMessage{
		statetest.NewRandL1Commitment(),
	}

	bcs.TestCodec(t, msg)

	bcs.TestCodecAndHash(t, &GetBlockMessage{
		statetest.TestL1Commitment,
	}, "30dd892c3980")
}
