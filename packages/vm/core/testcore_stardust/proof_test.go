package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/stretchr/testify/require"
	"os"
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
	t.Run("check PoI blob", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		err := ch.DepositIotasToL2(100_000, nil)
		require.NoError(t, err)

		h, err := ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		cFromLedger := ch.GetStateCommitment()
		cFromState := ch.GetRootCommitmentFromState()
		require.True(t, trie.EqualCommitments(cFromState, cFromLedger))

		data, err := os.ReadFile(randomFile)
		require.NoError(t, err)

		proof := ch.GetMerkleProof(blob.Contract.Hname(), blob.FieldValueKey(h, "file"))
		err = proof.Validate(cFromLedger, data)
		require.NoError(t, err)
		t.Logf("proof size = %d", len(proof.Bytes()))
	})
}
