package sm_messages

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMarshalUnmarshalGetBlockMessage(t *testing.T) {
	blocks := sm_gpa_utils.NewBlockFactory(t).GetBlocks(4, 1)
	for i := range blocks {
		commitment := blocks[i].L1Commitment()
		t.Logf("Checking block %v: %v", i, commitment)
		msg := NewGetBlockMessage(commitment, gpa.RandomTestNodeID())
		marshaled := rwutil.WriteToBytes(msg)
		unmarshaled, err := rwutil.ReadFromBytes(marshaled, NewEmptyGetBlockMessage())
		require.NoError(t, err)
		require.True(t, commitment.Equals(unmarshaled.GetL1Commitment()))
	}
}
