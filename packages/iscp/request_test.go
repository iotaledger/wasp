package iscp_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/dict"

	"github.com/iotaledger/wasp/packages/iscp"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestSerializeRequestData(t *testing.T) {
	var req iscp.RequestData
	var err error
	t.Run("off ledger", func(t *testing.T) {
		req = iscp.NewOffLedgerRequest(iscp.RandomChainID(), 3, 14, dict.New(), 1337)

		serialized := req.Bytes()
		req2, err := iscp.RequestDataFromMarshalUtil(marshalutil.New(serialized))
		require.NoError(t, err)

		reqBack := req2.(*iscp.OffLedgerRequestData)
		require.EqualValues(t, req.ID(), reqBack.ID())
		require.True(t, req.SenderAddress().Equal(reqBack.SenderAddress()))

		serialized2 := req2.Bytes()
		require.True(t, bytes.Equal(serialized, serialized2))
	})

	t.Run("on ledger", func(t *testing.T) {
		sender, _ := iotago.ParseEd25519AddressFromHexString("0152fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c649")
		requestMetadata := &iscp.RequestMetadata{
			SenderContract: iscp.Hn("sender_contract"),
			TargetContract: iscp.Hn("target_contract"),
			EntryPoint:     iscp.Hn("entrypoint"),
			Allowance:      &iscp.Assets{Iotas: 1},
			GasBudget:      1000,
		}
		outputOn := &iotago.ExtendedOutput{
			Address: sender,
			Amount:  123,
			NativeTokens: iotago.NativeTokens{
				&iotago.NativeToken{
					ID:     [iotago.NativeTokenIDLength]byte{1},
					Amount: big.NewInt(100),
				},
			},
			Blocks: iotago.FeatureBlocks{
				&iotago.MetadataFeatureBlock{Data: requestMetadata.Bytes()},
				&iotago.SenderFeatureBlock{Address: sender},
			},
		}
		req, err = iscp.OnLedgerFromUTXO(outputOn, &iotago.UTXOInput{})
		require.NoError(t, err)

		serialized := req.Bytes()
		req2, err := iscp.RequestDataFromMarshalUtil(marshalutil.New(serialized))
		require.NoError(t, err)
		require.True(t, req2.SenderAccount().Equals(iscp.NewAgentID(sender, requestMetadata.SenderContract)))
		require.True(t, req2.CallTarget().Equals(iscp.NewCallTarget(requestMetadata.TargetContract, requestMetadata.EntryPoint)))
		require.EqualValues(t, req.ID(), req2.ID())
		require.True(t, req.SenderAddress().Equal(req2.SenderAddress()))

		serialized2 := req2.Bytes()
		require.True(t, bytes.Equal(serialized, serialized2))
	})
}
