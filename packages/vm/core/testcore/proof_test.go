package testcore

import (
	"github.com/iotaledger/wasp/packages/state"
	"os"
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/stretchr/testify/require"
)

func TestProofs(t *testing.T) {
	t.Run("chain ID", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")

		proof := ch.GetMerkleProofRaw(nil)
		l1Commitment := ch.GetL1Commitment()
		err := state.ValidateMerkleProof(proof, l1Commitment.StateCommitment, ch.ChainID[:])
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
		err = state.ValidateMerkleProof(proof, ch.GetL1Commitment().StateCommitment, data)
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
		err = state.ValidateMerkleProof(proof, ch.GetL1Commitment().StateCommitment, rec.Bytes())

		require.NoError(t, err)
		t.Logf("proof size = %d", len(proof.Bytes()))
	})
	t.Run("check PoI past state", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		err := ch.DepositIotasToL2(100_000, nil)
		require.NoError(t, err)

		pastL1Commitment := ch.GetL1Commitment()
		pastBlockIndex := ch.State.BlockIndex()

		_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		_, err = ch.UploadWasm(nil, []byte("1234567890"))
		require.NoError(t, err)

		bi, err := ch.GetBlockInfo(pastBlockIndex)
		require.NoError(t, err)

		proof := ch.GetMerkleProof(blocklog.Contract.Hname(), blocklog.BlockInfoKey(pastBlockIndex))
		err = state.ValidateMerkleProof(proof, ch.GetL1Commitment().StateCommitment, bi.Bytes())

		require.NoError(t, err)
		t.Logf("proof size = %d", len(proof.Bytes()))

		require.True(t, state.EqualCommitments(pastL1Commitment.StateCommitment, bi.L1Commitment.StateCommitment))
	})
	t.Run("proof past block", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		err := ch.DepositIotasToL2(100_000, nil)
		require.NoError(t, err)

		pastBlockIndex := ch.State.BlockIndex()
		pastL1Commitment := ch.GetL1Commitment()

		_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		_, err = ch.UploadWasm(nil, []byte("1234567890"))
		require.NoError(t, err)

		pastBlockInfo, poi, err := ch.GetBlockProof(pastBlockIndex)
		require.NoError(t, err)

		require.True(t, state.EqualCommitments(pastL1Commitment.StateCommitment, pastBlockInfo.L1Commitment.StateCommitment))
		err = state.ValidateMerkleProof(poi, ch.GetL1Commitment().StateCommitment, pastBlockInfo.Bytes())

		require.NoError(t, err)
		t.Logf("proof size = %d", len(poi.Bytes()))
	})
}

func TestProofStateTerminals(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain(nil, "chain1")

	err := ch.DepositIotasToL2(100_000, nil)
	require.NoError(t, err)

	// core contracts must contain their hname at nil key in their state
	for _, ci := range corecontracts.AllSortedByName() {
		proof := ch.GetMerkleProof(ci.Hname(), nil)
		err = state.ValidateMerkleProof(proof, ch.GetL1Commitment().StateCommitment, ci.Hname().Bytes())
		if err != nil {
			t.Fatalf("core contract '%s' does not contain it's hname '%s' at its nil key",
				ci.Name, ci.Hname())
		}
		cS, err := ch.GetContractStateCommitment(ci.Hname())
		require.NoError(t, err)
		t.Logf("BEFORE: commitment to the state of the contract '%s': %s", ci.Name, cS)
	}

	_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
	require.NoError(t, err)

	_, err = ch.UploadWasm(nil, []byte("1234567890"))
	require.NoError(t, err)

	// core contracts must contain their hname at nil key in their state
	for _, ci := range corecontracts.AllSortedByName() {
		proof := ch.GetMerkleProof(ci.Hname(), nil)
		err = state.ValidateMerkleProof(proof, ch.GetL1Commitment().StateCommitment, ci.Hname().Bytes())
		if err != nil {
			t.Fatalf("core contract '%s' does not contain it's hname '%s' at its nil key",
				ci.Name, ci.Hname())
		}
		cS, err := ch.GetContractStateCommitment(ci.Hname())
		require.NoError(t, err)
		t.Logf("AFTER: commitment to the state of the contract '%s': %s", ci.Name, cS)
	}
}
