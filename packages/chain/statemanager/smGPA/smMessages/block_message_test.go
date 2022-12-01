package smMessages

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
)

func TestMarshalUnmarshalBlockMessage(t *testing.T) {
	_, blocks, _ := smGPAUtils.GetBlocks(t, 4, 1)
	for i := range blocks {
		t.Logf("Checking block %v: %v", i, blocks[i].GetHash())
		marshaled, err := NewBlockMessage(blocks[i], "SOMETHING").MarshalBinary()
		require.NoError(t, err)
		unmarshaled := NewEmptyBlockMessage()
		err = unmarshaled.UnmarshalBinary(marshaled)
		require.NoError(t, err)
		require.True(t, blocks[i].Equals(unmarshaled.GetBlock()))
	}
}
