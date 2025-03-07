package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestRequestDataSerialization(t *testing.T) {
	t.Run("off ledger", func(t *testing.T) {
		req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(3, 14), 1337, 100).Sign(cryptolib.NewKeyPair())
		bcs.TestCodec(t, isc.Request(req))
		rwutil.BytesTest(t, isc.Request(req), func(data []byte) (isc.Request, error) {
			return bcs.Unmarshal[isc.Request](data)
		})
	})

	t.Run("on ledger", func(t *testing.T) {
		sender := cryptolib.NewRandomAddress()
		objectRef := iotatest.RandomObjectRef()
		onledgerReq := iscmove.RefWithObject[iscmove.Request]{
			ObjectRef: *objectRef,
			Object: &iscmove.Request{
				ID:     *objectRef.ObjectID,
				Sender: sender,
				AssetsBag: iscmove.AssetsBagWithBalances{
					AssetsBag: iscmove.AssetsBag{
						ID:   *iotatest.RandomAddress(),
						Size: 1,
					},
					Balances: iscmove.AssetsBagBalances{iotajsonrpc.IotaCoinType: 200},
				},
				Message: iscmove.Message{
					Contract: uint32(isc.Hn("target_contract")),
					Function: uint32(isc.Hn("entrypoint")),
					Args:     [][]byte{},
				},
				Allowance: iscmove.Assets{
					Coins: iscmove.CoinBalances{iotajsonrpc.IotaCoinType: 100},
				},
				GasBudget: 1000,
			},
		}
		req, err := isc.OnLedgerFromRequest(&onledgerReq, cryptolib.NewRandomAddress())
		require.NoError(t, err)
		bcs.TestCodec(t, isc.Request(req))
		rwutil.BytesTest(t, isc.Request(req), func(data []byte) (isc.Request, error) {
			return bcs.Unmarshal[isc.Request](data)
		})
	})
}

func TestRequestIDSerialization(t *testing.T) {
	req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(3, 14, isc.NewCallArguments()), 1337, 200).Sign(cryptolib.NewKeyPair())
	requestID := req.ID()
	bcs.TestCodec(t, requestID)
	rwutil.StringTest(t, requestID, isc.RequestIDFromString)
}

func TestRequestRefSerialization(t *testing.T) {
	req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(3, 14, isc.NewCallArguments()), 1337, 200).Sign(cryptolib.NewKeyPair())
	reqRef0 := &isc.RequestRef{
		ID:   req.ID(),
		Hash: hashing.PseudoRandomHash(nil),
	}

	b := reqRef0.Bytes()
	reqRef1, err := isc.RequestRefFromBytes(b)
	require.NoError(t, err)
	require.Equal(t, reqRef0, reqRef1)

	bcs.TestCodec(t, reqRef0)
}
