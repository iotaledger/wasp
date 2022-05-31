package iscp

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func TestSerializeRequestData(t *testing.T) {
	var req Request
	var err error
	t.Run("off ledger", func(t *testing.T) {
		req = NewOffLedgerRequest(RandomChainID(), 3, 14, dict.New(), 1337).WithGasBudget(100).Sign(cryptolib.NewKeyPair())

		serialized := req.Bytes()
		req2, err := RequestDataFromMarshalUtil(marshalutil.New(serialized))
		require.NoError(t, err)

		reqBack := req2.(*offLedgerRequestData)
		require.EqualValues(t, req.ID(), reqBack.ID())
		require.True(t, req.SenderAccount().Equals(reqBack.SenderAccount()))

		serialized2 := req2.Bytes()
		require.True(t, bytes.Equal(serialized, serialized2))
	})

	t.Run("on ledger", func(t *testing.T) {
		sender := tpkg.RandAliasAddress()
		requestMetadata := &RequestMetadata{
			SenderContract: Hn("sender_contract"),
			TargetContract: Hn("target_contract"),
			EntryPoint:     Hn("entrypoint"),
			Allowance:      NewAllowanceIotas(1),
			GasBudget:      1000,
		}
		outputOn := &iotago.BasicOutput{
			Amount: 123,
			NativeTokens: iotago.NativeTokens{
				&iotago.NativeToken{
					ID:     [iotago.NativeTokenIDLength]byte{1},
					Amount: big.NewInt(100),
				},
			},
			Features: iotago.Features{
				&iotago.MetadataFeature{Data: requestMetadata.Bytes()},
				&iotago.SenderFeature{Address: sender},
			},
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: sender},
			},
		}
		req, err = OnLedgerFromUTXO(outputOn, &iotago.UTXOInput{})
		require.NoError(t, err)

		serialized := req.Bytes()
		req2, err := RequestDataFromMarshalUtil(marshalutil.New(serialized))
		require.NoError(t, err)
		chainID := ChainIDFromAddress(sender)
		require.True(t, req2.SenderAccount().Equals(NewContractAgentID(&chainID, requestMetadata.SenderContract)))
		require.True(t, req2.CallTarget().Equals(NewCallTarget(requestMetadata.TargetContract, requestMetadata.EntryPoint)))
		require.EqualValues(t, req.ID(), req2.ID())
		require.True(t, req.SenderAccount().Equals(req2.SenderAccount()))

		serialized2 := req2.Bytes()
		require.True(t, bytes.Equal(serialized, serialized2))
	})
}

func TestRequestIDToFromString(t *testing.T) {
	req := NewOffLedgerRequest(RandomChainID(), 3, 14, dict.New(), 1337).WithGasBudget(200)
	oritinalID := req.ID()
	s := oritinalID.String()
	require.NotEmpty(t, s)
	parsedID, err := RequestIDFromString(s)
	require.NoError(t, err)
	require.Equal(t, oritinalID.TransactionID, parsedID.TransactionID)
	require.Equal(t, oritinalID.TransactionOutputIndex, parsedID.TransactionOutputIndex)
}
