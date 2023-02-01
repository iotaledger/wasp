package isc

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func TestSerializeRequestData(t *testing.T) {
	var req Request
	var err error
	t.Run("off ledger", func(t *testing.T) {
		req = NewOffLedgerRequest(RandomChainID(), 3, 14, dict.New(), 1337).WithGasBudget(100).Sign(cryptolib.NewKeyPair())

		serialized := req.Bytes()
		req2, err := NewRequestFromMarshalUtil(marshalutil.New(serialized))
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
				&iotago.MetadataFeature{Data: requestMetadata.Bytes()},
				&iotago.SenderFeature{Address: sender},
			},
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: sender},
			},
		}
		req, err = OnLedgerFromUTXO(basicOutput, iotago.OutputID{})
		require.NoError(t, err)

		serialized := req.Bytes()
		req2, err := NewRequestFromMarshalUtil(marshalutil.New(serialized))
		require.NoError(t, err)
		chainID := ChainIDFromAddress(sender)
		require.True(t, req2.SenderAccount().Equals(NewContractAgentID(chainID, requestMetadata.SenderContract)))
		require.True(t, req2.CallTarget().Equals(NewCallTarget(requestMetadata.TargetContract, requestMetadata.EntryPoint)))
		require.EqualValues(t, req.ID(), req2.ID())
		require.True(t, req.SenderAccount().Equals(req2.SenderAccount()))

		serialized2 := req2.Bytes()
		require.True(t, bytes.Equal(serialized, serialized2))
	})
}

func TestRequestIDToFromString(t *testing.T) {
	req := NewOffLedgerRequest(RandomChainID(), 3, 14, dict.New(), 1337).WithGasBudget(200).Sign(cryptolib.NewKeyPair())
	oritinalID := req.ID()
	s := oritinalID.String()
	require.NotEmpty(t, s)
	parsedID, err := RequestIDFromString(s)
	require.NoError(t, err)
	require.EqualValues(t, oritinalID, parsedID)
}
