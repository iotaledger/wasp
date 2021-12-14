package iscp

import (
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestOnledgerRequest(t *testing.T) {
	sender, _ := iotago.ParseEd25519AddressFromHexString("0152fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c649")
	requestMetadata := &RequestMetadata{
		SenderContract: Hn("sender_contract"),
		TargetContract: Hn("target_contract"),
		EntryPoint:     Hn("entrypoint"),
		Transfer:       &Assets{Iotas: 1},
		GasBudget:      1000,
	}
	output := &iotago.ExtendedOutput{
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
	req, err := OnLedgerFromUTXO(output, &iotago.UTXOInput{})
	require.NoError(t, err)

	serialized := req.Bytes()
	req2, err := OnLedgerRequestFromMarshalUtil(marshalutil.New(serialized))
	require.NoError(t, err)
	require.True(t, req2.SenderAccount().Equals(NewAgentID(sender, requestMetadata.SenderContract)))
	require.True(t, req2.CallTarget().Equals(NewCallTarget(requestMetadata.TargetContract, requestMetadata.EntryPoint)))
}
