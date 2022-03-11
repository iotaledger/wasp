package testcore

import (
	"os"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/stretchr/testify/require"
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

		data, err := os.ReadFile(randomFile)
		require.NoError(t, err)

		key := blob.FieldValueKey(h, "file")
		proof := ch.GetMerkleProof(blob.Contract.Hname(), key)
		err = proof.Validate(ch.GetStateCommitment(), data)
		require.NoError(t, err)
		t.Logf("key size = %d", len(key))
		t.Logf("proof size = %d", len(proof.Bytes()))
	})
	t.Run("check PoI receipt", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		err := ch.DepositIotasToL2(100_000, nil)
		require.NoError(t, err)

		_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		rec := ch.LastReceipt()

		recKey := blocklog.RequestReceiptKey(rec.LookupKey())
		proof := ch.GetMerkleProof(blocklog.Contract.Hname(), recKey)
		err = proof.Validate(ch.GetStateCommitment(), rec.Bytes())

		require.NoError(t, err)
		t.Logf("proof size = %d", len(proof.Bytes()))
	})
	t.Run("check PoI past state", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		err := ch.DepositIotasToL2(100_000, nil)
		require.NoError(t, err)

		pastStateCommitment := ch.GetStateCommitment()
		pastBlockIndex := ch.State.BlockIndex()

		_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		_, err = ch.UploadWasm(nil, []byte("1234567890"))
		require.NoError(t, err)

		bi, err := ch.GetBlockInfo(pastBlockIndex)
		require.NoError(t, err)

		proof := ch.GetMerkleProof(blocklog.Contract.Hname(), blocklog.BlockInfoKey(pastBlockIndex))
		err = proof.Validate(ch.GetStateCommitment(), bi.Bytes())

		require.NoError(t, err)
		t.Logf("proof size = %d", len(proof.Bytes()))

		require.True(t, trie.EqualCommitments(pastStateCommitment, bi.StateCommitment))
	})
}
