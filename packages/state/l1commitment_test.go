package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestL1Commitment(t *testing.T) {
	sc := RandL1Commitment()

	data := sc.Bytes()
	scBack, err := L1CommitmentFromBytes(data)
	require.NoError(t, err)
	require.True(t, EqualCommitments(sc.GetTrieRoot(), scBack.GetTrieRoot()))
	require.True(t, sc.GetBlockHash().Equals(scBack.GetBlockHash()))
}
