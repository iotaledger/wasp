package vmtxbuilder

import (
	"math/big"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/serializer"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/hashing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func rndAliasID() (ret iotago.AliasID) {
	a := tpkg.RandAliasAddress()
	copy(ret[:], a[:])
	return
}

func sumDeposits(outs iotago.Outputs) uint64 {
	var ret uint64
	for _, o := range outs {
		ret += o.Deposit()
	}
	return ret
}

func TestNewTxBuilder(t *testing.T) {
	addr := tpkg.RandEd25519Address()
	stateMetadata := hashing.HashStrings("test")
	nextStateMetadata := hashing.HashStrings("test1")
	aliasID := rndAliasID()
	const totalIotas = 1000
	anchor := &iotago.AliasOutput{
		Amount:               totalIotas,
		NativeTokens:         nil,
		AliasID:              aliasID,
		StateController:      addr,
		GovernanceController: addr,
		StateIndex:           0,
		StateMetadata:        stateMetadata[:],
		FoundryCounter:       0,
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: aliasID.ToAddress(),
			},
		},
	}
	anchorID, _ := tpkg.RandUTXOInput()

	t.Run("1", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
			return nil, iotago.UTXOInput{}
		})
		require.EqualValues(t, 1, txb.numInputs())
		require.EqualValues(t, 1, txb.numOutputs())
		require.False(t, txb.InputsAreFull())
		require.False(t, txb.outputsAreFull())

		essence := txb.BuildTransactionEssence(nextStateMetadata, time.Now())
		require.EqualValues(t, 1, len(essence.Inputs))
		require.EqualValues(t, 1, len(essence.Outputs))

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))

		total := sumDeposits(essence.Outputs)
		require.EqualValues(t, totalIotas, total)
		//essenceBack := iotago.TransactionEssence{}
		//consumed, err := essenceBack.Deserialize(essenceBytes, serializer.DeSeriModeNoValidation, nil)
		//require.NoError(t, err)
		//require.EqualValues(t, len(essenceBytes), consumed)

	})

	t.Run("2", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
			return nil, iotago.UTXOInput{}
		})
		txb.AddDeltaIotas(42)
		essence := txb.BuildTransactionEssence(nextStateMetadata, time.Now())

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))

		total := sumDeposits(essence.Outputs)
		require.EqualValues(t, totalIotas+42, int(total))

		//essenceBack := iotago.TransactionEssence{}
		//consumed, err := essenceBack.Deserialize(essenceBytes, serializer.DeSeriModeNoValidation, nil)
		//require.NoError(t, err)
		//require.EqualValues(t, len(essenceBytes), consumed)

		//var buf bytes.Buffer
		//tpkg.Must(binary.Write(&buf, binary.LittleEndian, iotago.PayloadTransaction))
		//
		//sigTxPayload := &iotago.Transaction{}
		//_, err := buf.Write(unTxData)
		//tpkg.Must(err)
		//sigTxPayload.Essence = unTx
	})
}
