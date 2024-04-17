package testcore

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const (
	randomFile = "blob_test.go"
)

func TestUploadBlob(t *testing.T) {
	t.Run("from binary", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()

		ch.MustDepositBaseTokensToL2(100_000, nil)

		h, err := ch.UploadBlob(nil, "field", "dummy data")
		require.NoError(t, err)

		_, ok := ch.GetBlobInfo(h)
		require.True(t, ok)

		testdbhash.VerifyContractStateHash(env, blob.Contract, "", t.Name())
	})
	t.Run("huge", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()

		err := ch.DepositBaseTokensToL2(1_000_000, nil)
		require.NoError(t, err)

		limits := *gas.LimitsDefault
		limits.MaxGasPerRequest = 10 * limits.MaxGasPerRequest
		limits.MaxGasExternalViewCall = 10 * limits.MaxGasExternalViewCall
		ch.SetGasLimits(nil, &limits)
		ch.WaitForRequestsMark()

		size := int64(1 * 900 * 1024) // 900 KB
		randomData := make([]byte, size+1)
		_, err = rand.Read(randomData)
		require.NoError(t, err)
		h, err := ch.UploadBlob(nil, "field", randomData)
		require.NoError(t, err)

		_, ok := ch.GetBlobInfo(h)
		require.True(t, ok)
	})
	t.Run("from file", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()

		err := ch.DepositBaseTokensToL2(100_000, nil)
		require.NoError(t, err)

		h, err := ch.UploadBlobFromFile(nil, randomFile, "file")
		require.NoError(t, err)

		_, ok := ch.GetBlobInfo(h)
		require.True(t, ok)
	})
	t.Run("several", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()

		err := ch.DepositBaseTokensToL2(100_000, nil)
		require.NoError(t, err)

		const howMany = 5
		hashes := make([]hashing.HashValue, howMany)
		for i := 0; i < howMany; i++ {
			data := []byte(fmt.Sprintf("dummy data #%d", i))
			hashes[i], err = ch.UploadBlob(nil, "field", data)
			require.NoError(t, err)
			m, ok := ch.GetBlobInfo(hashes[i])
			require.True(t, ok)
			require.EqualValues(t, 1, len(m))
			require.EqualValues(t, len(data), m["field"])
		}
		ret, err := ch.CallView(blob.Contract.Name, blob.ViewListBlobs.Name)
		require.NoError(t, err)
		require.EqualValues(t, howMany, len(ret))
		for _, h := range hashes {
			sizeBin := ret.Get(kv.Key(h[:]))
			size, err := codec.DecodeUint32(sizeBin)
			require.NoError(t, err)
			require.EqualValues(t, len("dummy data #1"), int(size))

			ret, err := ch.CallView(blob.Contract.Name, blob.ViewGetBlobField.Name,
				blob.ParamHash, h,
				blob.ParamField, "field",
			)
			require.NoError(t, err)
			require.EqualValues(t, 1, len(ret))
			data := ret.Get(blob.ParamBytes)
			require.EqualValues(t, size, len(data))
		}
	})
}

func TestUploadContractBinary(t *testing.T) {
	vmType := "dummy"

	t.Run("upload once", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()
		ch.MustDepositBaseTokensToL2(100_000, nil)
		binary := []byte("supposed to be binary")
		hash, err := ch.UploadContractBinary(nil, vmType, binary)
		require.NoError(t, err)

		vmTypeBack, binBack, err := ch.GetContractBinary(hash)
		require.NoError(t, err)

		require.EqualValues(t, vmType, vmTypeBack)
		require.EqualValues(t, binary, binBack)
	})
	t.Run("upload twice", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()
		ch.MustDepositBaseTokensToL2(100_000, nil)
		binary := []byte("supposed to be binary")
		hash1, err := ch.UploadContractBinary(nil, vmType, binary)
		require.NoError(t, err)

		// we upload exactly the same, if it exists it silently returns no error
		hash2, err := ch.UploadContractBinary(nil, vmType, binary)
		require.NoError(t, err)

		require.EqualValues(t, hash1, hash2)

		vmTypeBack, binBack, err := ch.GetContractBinary(hash1)
		require.NoError(t, err)

		require.EqualValues(t, vmType, vmTypeBack)
		require.EqualValues(t, binary, binBack)
	})
	t.Run("list blobs", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()
		ch.MustDepositBaseTokensToL2(100_000, nil)
		binary := []byte("supposed to be binary")
		_, err := ch.UploadContractBinary(nil, vmType, binary)
		require.NoError(t, err)

		ret, err := ch.CallView(blob.Contract.Name, blob.ViewListBlobs.Name)
		require.NoError(t, err)
		require.EqualValues(t, 1, len(ret))
	})
}

func TestBigBlob(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()
	ch.MustDepositBaseTokensToL2(1*isc.Million, nil)

	vmType := "dummy"
	upload := func(n int) uint64 {
		blobBin := make([]byte, n)
		_, err := ch.UploadContractBinary(ch.OriginatorPrivateKey, vmType, blobBin)
		require.NoError(t, err)
		return ch.LastReceipt().GasBurned
	}

	gas1k := upload(100_000)
	gas2k := upload(200_000)

	t.Log(gas1k, gas2k)
	require.Greater(t, gas2k, gas1k)
}
