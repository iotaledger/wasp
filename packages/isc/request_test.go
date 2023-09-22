package isc

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestRequestDataSerialization(t *testing.T) {
	var req Request
	var err error
	t.Run("off ledger", func(t *testing.T) {
		req = NewOffLedgerRequest(RandomChainID(), 3, 14, dict.New(), 1337, 100).Sign(cryptolib.NewKeyPair())
		rwutil.ReadWriteTest(t, req.(*OffLedgerRequestData), new(OffLedgerRequestData))
		rwutil.BytesTest(t, req, RequestFromBytes)
	})

	t.Run("on ledger", func(t *testing.T) {
		sender := tpkg.RandAliasAddress()
		requestMetadata := &RequestMetadata{
			SenderContract: ContractIdentityFromHname(Hn("sender_contract")),
			TargetContract: Hn("target_contract"),
			EntryPoint:     Hn("entrypoint"),
			Allowance:      NewAssetsBaseTokens(1),
			GasBudget:      1000,
		}
		basicOutput := &iotago.BasicOutput{
			Amount: 123,
			NativeTokens: iotago.NativeTokens{
				&iotago.NativeToken{
					ID:     [iotago.NativeTokenIDLength]byte{1},
					Amount: big.NewInt(100),
				},
			},
			Features: iotago.Features{
				&iotago.SenderFeature{Address: sender},
				&iotago.MetadataFeature{Data: requestMetadata.Bytes()},
			},
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: sender},
			},
		}
		req, err = OnLedgerFromUTXO(basicOutput, iotago.OutputID{})
		require.NoError(t, err)
		rwutil.ReadWriteTest(t, req.(*onLedgerRequestData), new(onLedgerRequestData))
	})
}

func TestRequestIDSerialization(t *testing.T) {
	req := NewOffLedgerRequest(RandomChainID(), 3, 14, dict.New(), 1337, 200).Sign(cryptolib.NewKeyPair())
	requestID := req.ID()
	rwutil.ReadWriteTest(t, &requestID, new(RequestID))
	rwutil.BytesTest(t, requestID, RequestIDFromBytes)
	rwutil.StringTest(t, requestID, RequestIDFromString)
}

func TestRequestRefSerialization(t *testing.T) {
	req := NewOffLedgerRequest(RandomChainID(), 3, 14, dict.New(), 1337, 200).Sign(cryptolib.NewKeyPair())
	reqRef0 := &RequestRef{
		ID:   req.ID(),
		Hash: hashing.PseudoRandomHash(nil),
	}

	b := reqRef0.Bytes()
	reqRef1, err := RequestRefFromBytes(b)
	require.NoError(t, err)
	require.Equal(t, reqRef0, reqRef1)

	rwutil.ReadWriteTest(t, reqRef0, new(RequestRef))
}
