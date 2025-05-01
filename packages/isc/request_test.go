package isc_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestImpersonatedOffLedgerRequest(t *testing.T) {
	requestFrom := cryptolib.NewRandomAddress()

	req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(isc.HnameNil, isc.HnameNil), 0, 0).
		WithAllowance(isc.NewAssets(1074)).
		WithNonce(1074).
		WithGasBudget(1074)

	impRequest := isc.NewImpersonatedOffLedgerRequest(req.(*isc.OffLedgerRequestDataEssence)).
		WithSenderAddress(requestFrom)

	require.NotNil(t, impRequest)
	require.NotNil(t, impRequest.SenderAccount())
	require.Equal(t, impRequest.SenderAccount().String(), requestFrom.String())
}

func TestRequestToJSONObject(t *testing.T) {
	pk, addr := solo.NewEthereumAccount()
	req, err := isc.NewEVMOffLedgerTxRequest(isc.EmptyChainID(), types.MustSignNewTx(pk, types.NewEIP155Signer(big.NewInt(int64(1074))),
		&types.LegacyTx{
			Nonce:    0,
			To:       &addr,
			Value:    big.NewInt(123),
			Gas:      100000,
			GasPrice: big.NewInt(10000000000),
			Data:     []byte{1, 0, 7, 4},
		}))
	require.NoError(t, err)

	obj := isc.RequestToJSONObject(req)
	require.NotNil(t, obj)
	require.NotNil(t, obj.SenderAccount)

	keyPair := cryptolib.NewKeyPair()
	tx2 := isc.NewOffLedgerRequest(
		isc.EmptyChainID(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		0,
		1*isc.Million,
	).WithAllowance(isc.NewAssets(5000)).
		WithSender(keyPair.GetPublicKey())

	obj2 := isc.RequestToJSONObject(tx2)
	require.NotNil(t, obj2)
	require.NotNil(t, obj2.SenderAccount)

	requestFrom := cryptolib.NewRandomAddress()
	req3 := isc.NewOffLedgerRequest(isc.EmptyChainID(), isc.NewMessage(isc.HnameNil, isc.HnameNil), 0, 0)

	impRequest := isc.NewImpersonatedOffLedgerRequest(req3.(*isc.OffLedgerRequestDataEssence)).
		WithSenderAddress(requestFrom)

	require.NotNil(t, impRequest)
	require.NotNil(t, impRequest.SenderAccount())
	require.Equal(t, impRequest.SenderAccount().String(), requestFrom.String())

	obj3 := isc.RequestToJSONObject(impRequest)
	require.NotNil(t, obj3)
	require.NotNil(t, obj3.SenderAccount)
}

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
					Assets: *iscmove.NewAssets(200),
				},
				Message: iscmove.Message{
					Contract: uint32(isc.Hn("target_contract")),
					Function: uint32(isc.Hn("entrypoint")),
					Args:     [][]byte{},
				},
				AllowanceBCS: bcs.MustMarshal(iscmove.NewAssets(100)),
				GasBudget:    1000,
			},
		}
		req, err := isc.OnLedgerFromMoveRequest(&onledgerReq, cryptolib.NewRandomAddress())
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
