package sm_messages

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMarshalUnmarshalBlockMessage(t *testing.T) {
	blocks := sm_gpa_utils.NewBlockFactory(t).GetBlocks(4, 1)
	for i := range blocks {
		t.Logf("Checking block %v: %v", i, blocks[i].L1Commitment())
		msg := NewBlockMessage(blocks[i], gpa.RandomTestNodeID())
		marshaled := rwutil.WriteToBytes(msg)
		unmarshaled, err := rwutil.ReadFromBytes(marshaled, NewEmptyBlockMessage())
		require.NoError(t, err)
		sm_gpa_utils.CheckBlocksEqual(t, blocks[i], unmarshaled.GetBlock())
	}
}

func TestSerializationBlockMessage(t *testing.T) {
	msg := &BlockMessage{
		gpa.BasicMessage{},
		state.RandomBlock(),
	}

	rwutil.ReadWriteTest(t, msg, new(BlockMessage))
}
