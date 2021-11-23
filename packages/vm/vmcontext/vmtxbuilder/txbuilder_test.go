package vmtxbuilder

import (
	"math/big"
	"testing"

	"github.com/iotaledger/wasp/packages/testutil/testiotago"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/wasp/packages/iscp"

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

func TestNewTxBuilder(t *testing.T) {
	addr := tpkg.RandEd25519Address()
	stateMetadata := hashing.HashStrings("test")
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
		iotasTotal, assetsTotal, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
		require.EqualValues(t, 1000, iotasTotal)
		require.EqualValues(t, 0, len(assetsTotal))

		require.EqualValues(t, 1, txb.numInputs())
		require.EqualValues(t, 1, txb.numOutputs())
		require.False(t, txb.InputsAreFull())
		require.False(t, txb.outputsAreFull())

		essence := txb.BuildTransactionEssence(&iscp.StateData{})
		require.EqualValues(t, 1, len(essence.Inputs))
		require.EqualValues(t, 1, len(essence.Outputs))

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
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
		_, _, isBalanced := txb.TotalAssets()
		require.False(t, isBalanced)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))

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
	t.Run("3", func(t *testing.T) {
		nativeTokenAmount := testiotago.RandNativeTokenAmount(84)
		balanceLoader := func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
			if id == nativeTokenAmount.ID {
				return nativeTokenAmount.Amount, iotago.UTXOInput{}
			}
			panic("too bad")
		}
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		out := &iotago.ExtendedOutput{
			Address:      nil,
			Amount:       42,
			NativeTokens: iotago.NativeTokens{nativeTokenAmount},
			Blocks:       nil,
		}
		reqData, err := iscp.OnLedgerFromUTXO(&iscp.UTXOMetaData{}, out)
		require.NoError(t, err)
		txb.ConsumeOutput(reqData)
		txb.AddDeltaIotas(42)
		txb.AddDeltaNativeToken(nativeTokenAmount.ID, nativeTokenAmount.Amount)

		totalIotas, totalTokens, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
		require.EqualValues(t, 1042, totalIotas)
		require.EqualValues(t, 1, len(totalTokens))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}
