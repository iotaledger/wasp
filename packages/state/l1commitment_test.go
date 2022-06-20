package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestL1Commitment(t *testing.T) {
	t.Run("base1", func(t *testing.T) {
		sc := OriginL1Commitment()
		require.True(t, EqualCommitments(sc.StateCommitment, OriginStateCommitment()))
		require.EqualValues(t, sc.BlockHash, OriginBlockHash())

		data := sc.Bytes()
		require.EqualValues(t, l1CommitmentSize, len(data))

		scBack, err := L1CommitmentFromBytes(data)
		require.NoError(t, err)
		require.True(t, EqualCommitments(sc.StateCommitment, scBack.StateCommitment))
	})
	t.Run("base2", func(t *testing.T) {
		sc := RandL1Commitment()

		data := sc.Bytes()
		scBack, err := L1CommitmentFromBytes(data)
		require.NoError(t, err)
		require.True(t, EqualCommitments(sc.StateCommitment, scBack.StateCommitment))
		require.EqualValues(t, sc.BlockHash, scBack.BlockHash)
	})
}
