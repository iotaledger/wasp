package testcore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
)

const (
	randomFile = "blob_test.go"
	wasmFile   = "sbtests/sbtestsc/testcore_bg.wasm"
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
			sizeBin := ret.MustGet(kv.Key(h[:]))
			size, err := codec.DecodeUint32(sizeBin)
			require.NoError(t, err)
			require.EqualValues(t, len("dummy data #1"), int(size))

			ret, err := ch.CallView(blob.Contract.Name, blob.ViewGetBlobField.Name,
				blob.ParamHash, h,
				blob.ParamField, "field",
			)
			require.NoError(t, err)
			require.EqualValues(t, 1, len(ret))
			data := ret.MustGet(blob.ParamBytes)
			require.EqualValues(t, size, len(data))
		}
	})
}

func TestUploadWasm(t *testing.T) {
	t.Run("upload wasm", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()
		ch.MustDepositBaseTokensToL2(100_000, nil)
		binary := []byte("supposed to be wasm")
		hwasm, err := ch.UploadWasm(nil, binary)
		require.NoError(t, err)

		binBack, err := ch.GetWasmBinary(hwasm)
		require.NoError(t, err)

		require.EqualValues(t, binary, binBack)
	})
	t.Run("upload twice", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()
		ch.MustDepositBaseTokensToL2(100_000, nil)
		binary := []byte("supposed to be wasm")
		hwasm1, err := ch.UploadWasm(nil, binary)
		require.NoError(t, err)

		// we upload exactly the same, if it exists it silently returns no error
		hwasm2, err := ch.UploadWasm(nil, binary)
		require.NoError(t, err)

		require.EqualValues(t, hwasm1, hwasm2)

		binBack, err := ch.GetWasmBinary(hwasm1)
		require.NoError(t, err)

		require.EqualValues(t, binary, binBack)
	})
	t.Run("upload wasm from file", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		ch.MustDepositBaseTokensToL2(100_000, nil)
		progHash, err := ch.UploadWasmFromFile(nil, wasmFile)
		require.NoError(t, err)

		err = ch.DeployContract(nil, "testCore", progHash)
		require.NoError(t, err)
	})
	t.Run("list blobs", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain()
		ch.MustDepositBaseTokensToL2(100_000, nil)
		_, err := ch.UploadWasmFromFile(nil, wasmFile)
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

	upload := func(n int) uint64 {
		blobBin := make([]byte, n)
		_, err := ch.UploadWasm(ch.OriginatorPrivateKey, blobBin)
		require.NoError(t, err)
		return ch.LastReceipt().GasBurned
	}

	gas1k := upload(100_000)
	gas2k := upload(200_000)

	t.Log(gas1k, gas2k)
	require.Greater(t, gas2k, gas1k)
}
