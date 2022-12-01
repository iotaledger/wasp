package smMessages

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
)

func TestMarshalUnmarshalGetBlockMessage(t *testing.T) {
	_, blocks, _ := smGPAUtils.GetBlocks(t, 4, 1)
	for i := range blocks {
		blockHash := blocks[i].GetHash()
		t.Logf("Checking block %v: %v", i, blockHash)
		marshaled, err := NewGetBlockMessage(uint32(i+1), blockHash, "SOMETHING").MarshalBinary()
		require.NoError(t, err)
		unmarshaled := NewEmptyGetBlockMessage()
		err = unmarshaled.UnmarshalBinary(marshaled)
		require.NoError(t, err)
		require.True(t, blockHash.Equals(unmarshaled.GetBlockHash()))
		require.Equal(t, uint32(i+1), unmarshaled.GetBlockIndex())
	}
}
