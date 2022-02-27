package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProofs(t *testing.T) {
	t.Run("chain ID", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")

		proof := ch.GetMerkleProofRaw(nil)
		stateCommitment := ch.GetStateCommitment()
		err := proof.Validate(stateCommitment, ch.ChainID[:])
		require.NoError(t, err)
	})
}
