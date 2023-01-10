package smMessages

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/gpa"
)

func TestMarshalUnmarshalBlockMessage(t *testing.T) {
	blocks := smGPAUtils.NewBlockFactory(t).GetBlocks(4, 1)
	for i := range blocks {
		t.Logf("Checking block %v: %v", i, blocks[i].L1Commitment())
		marshaled, err := NewBlockMessage(blocks[i], gpa.RandomTestNodeID()).MarshalBinary()
		require.NoError(t, err)
		unmarshaled := NewEmptyBlockMessage()
		err = unmarshaled.UnmarshalBinary(marshaled)
		require.NoError(t, err)
		require.True(t, blocks[i].Hash().Equals(unmarshaled.GetBlock().Hash())) // Should be Equals instead of Hash().Equals()
	}
}
