package testcore

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/stretchr/testify/require"
)

var fileName = "blob_test.go"

func TestUploadBlob(t *testing.T) {
	t.Run("from binary", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		h, err := ch.UploadBlob(nil, "field", "dummy data")
		require.NoError(t, err)

		_, ok := ch.GetBlobInfo(h)
		require.True(t, ok)
	})
	t.Run("from file", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		// get more iotas for originator
		originatorBalance := env.L1AddressBalances(ch.OriginatorAddress).Iotas
		_, err := env.L1Ledger().GetFundsFromFaucet(ch.OriginatorAddress)
		require.NoError(t, err)
		env.AssertL1AddressIotas(ch.OriginatorAddress, originatorBalance+utxodb.FundsFromFaucetAmount)

		h, err := ch.UploadBlobFromFile(nil, fileName, "file")
		require.NoError(t, err)

		_, ok := ch.GetBlobInfo(h)
		require.True(t, ok)
	})
	t.Run("several", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")

		const howMany = 5
		hashes := make([]hashing.HashValue, howMany)
		var err error
		for i := 0; i < howMany; i++ {
			data := []byte(fmt.Sprintf("dummy data #%d", i))
			hashes[i], err = ch.UploadBlob(nil, "field", data)
			require.NoError(t, err)
			m, ok := ch.GetBlobInfo(hashes[i])
			require.True(t, ok)
			require.EqualValues(t, 1, len(m))
			require.EqualValues(t, len(data), m["field"])
		}
		ret, err := ch.CallView(blob.Contract.Name, blob.FuncListBlobs.Name)
		require.NoError(t, err)
		require.EqualValues(t, howMany, len(ret))
		for _, h := range hashes {
			sizeBin := ret.MustGet(kv.Key(h[:]))
			size, err := codec.DecodeUint32(sizeBin)
			require.NoError(t, err)
			require.EqualValues(t, len("dummy data #1"), int(size))

			ret, err := ch.CallView(blob.Contract.Name, blob.FuncGetBlobField.Name,
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
