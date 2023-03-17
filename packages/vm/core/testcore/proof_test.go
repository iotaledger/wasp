package testcore

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func TestProofs(t *testing.T) {
	t.Run("check PoI blob", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()

		err := ch.DepositBaseTokensToL2(100_000, nil)
		require.NoError(t, err)

		h, err := ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		data, err := os.ReadFile(randomFile)
		require.NoError(t, err)

		key := blob.FieldValueKey(h, "file")
		proof := ch.GetMerkleProof(blob.Contract.Hname(), key)
		err = proof.ValidateValue(ch.GetL1Commitment().TrieRoot(), data)
		require.NoError(t, err)
		t.Logf("key size = %d", len(key))
	})
	t.Run("check PoI receipt", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()

		err := ch.DepositBaseTokensToL2(100_000, nil)
		require.NoError(t, err)

		_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		lastBlockReceipts := ch.GetRequestReceiptsForBlock()
		rec := lastBlockReceipts[len(lastBlockReceipts)-1]

		recKey := blocklog.RequestReceiptKey(rec.LookupKey())
		proof := ch.GetMerkleProof(blocklog.Contract.Hname(), recKey)
		err = proof.ValidateValue(ch.GetL1Commitment().TrieRoot(), rec.Bytes())

		require.NoError(t, err)
	})
	t.Run("check PoI past state", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()

		err := ch.DepositBaseTokensToL2(100_000, nil)
		require.NoError(t, err)

		pastL1Commitment := ch.GetL1Commitment()
		pastBlockIndex, err := ch.Store().LatestBlockIndex()
		require.NoError(t, err)

		_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		_, err = ch.UploadWasm(nil, []byte("1234567890"))
		require.NoError(t, err)

		bi, err := ch.GetBlockInfo(pastBlockIndex)
		require.NoError(t, err)

		proof := ch.GetMerkleProof(blocklog.Contract.Hname(), blocklog.BlockInfoKey(pastBlockIndex))
		err = proof.ValidateValue(ch.GetL1Commitment().TrieRoot(), bi.Bytes())

		require.NoError(t, err)

		require.Equal(t, pastL1Commitment.TrieRoot(), bi.L1Commitment.TrieRoot())
	})
	t.Run("proof past block", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()

		err := ch.DepositBaseTokensToL2(100_000, nil)
		require.NoError(t, err)

		pastBlockIndex, err := ch.Store().LatestBlockIndex()
		require.NoError(t, err)
		pastL1Commitment := ch.GetL1Commitment()

		_, err = ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		_, err = ch.UploadWasm(nil, []byte("1234567890"))
		require.NoError(t, err)

		pastBlockInfo, poi, err := ch.GetBlockProof(pastBlockIndex)
		require.NoError(t, err)

		require.Equal(t, pastL1Commitment.TrieRoot(), pastBlockInfo.L1Commitment.TrieRoot())
		err = poi.ValidateValue(ch.GetL1Commitment().TrieRoot(), pastBlockInfo.Bytes())

		require.NoError(t, err)
	})
}
