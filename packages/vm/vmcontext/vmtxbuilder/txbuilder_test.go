package vmtxbuilder

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/testutil/testiotago"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func rndAliasID() (ret iotago.AliasID) {
	a := tpkg.RandAliasAddress()
	copy(ret[:], a[:])
	return
}

func consumeUTXO(t *testing.T, txb *AnchorTransactionBuilder, iotas uint64, id iotago.NativeTokenID, amount uint64) {
	depositNativeToken := testiotago.NewNativeTokenAmount(id, amount)
	out := &iotago.ExtendedOutput{
		Address:      nil,
		Amount:       iotas,
		NativeTokens: iotago.NativeTokens{depositNativeToken},
		Blocks:       nil,
	}
	reqData, err := iscp.OnLedgerFromUTXO(&iscp.UTXOMetaData{}, out)
	require.NoError(t, err)
	txb.Consume(reqData)
}

func TestTxBuilderBasic(t *testing.T) {
	addr := tpkg.RandEd25519Address()
	stateMetadata := hashing.HashStrings("test")
	aliasID := rndAliasID()
	const initialTotalIotas = 1000
	anchor := &iotago.AliasOutput{
		Amount:               initialTotalIotas,
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

	const numInitBalances = 5

	nativeTokensOnChain := make(iotago.NativeTokens, 0)
	utxoInputsOnChain := make([]iotago.UTXOInput, 0)

	for i := uint64(0); i < numInitBalances; i++ {
		nativeTokensOnChain = append(nativeTokensOnChain, testiotago.RandNativeTokenAmount(2000+i))
		utxoInputsOnChain = append(utxoInputsOnChain, testiotago.RandUTXOInput())
	}

	balanceLoader := func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
		for i, nt := range nativeTokensOnChain {
			if id == nt.ID {
				return nt.Amount, utxoInputsOnChain[i]
			}
		}
		panic("too bad")
	}

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
	})
	t.Run("2", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
			return nil, iotago.UTXOInput{}
		})
		txb.addDeltaIotas(42)
		_, _, isBalanced := txb.TotalAssets()
		require.False(t, isBalanced)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("3", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		consumeUTXO(t, txb, 42, nativeTokensOnChain[0].ID, 42)

		expectedBalance := new(big.Int).Set(nativeTokensOnChain[0].Amount)
		expectedBalance.Add(expectedBalance, new(big.Int).SetUint64(42))

		totalIotas, totalTokens, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+42, totalIotas)
		require.EqualValues(t, 1, len(totalTokens))
		require.True(t, totalTokens[nativeTokensOnChain[0].ID].Cmp(expectedBalance) == 0)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check", func(t *testing.T) {
		const runTimes = 100
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		amounts := make(map[int]uint64)
		for i := 0; i < 100; i++ {
			amount := uint64(rand.Intn(runTimes))
			idx := rand.Intn(numInitBalances)
			s, _ := amounts[idx]
			amounts[idx] = s + amount

			consumeUTXO(t, txb, 1, nativeTokensOnChain[idx].ID, amount)

			totalIotas, _, isBalanced := txb.TotalAssets()
			require.True(t, isBalanced)
			require.EqualValues(t, initialTotalIotas+i+1, totalIotas)
		}
		totalIotas, totalTokens, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes, totalIotas)
		require.EqualValues(t, numInitBalances, len(totalTokens))

		for idx, b := range amounts {
			expectedBalance := new(big.Int).Set(nativeTokensOnChain[idx].Amount)
			expectedBalance.Add(expectedBalance, new(big.Int).SetUint64(b))

			require.True(t, totalTokens[nativeTokensOnChain[idx].ID].Cmp(expectedBalance) == 0)
		}

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("exceed inputs", func(t *testing.T) {
		const runTimes = 150
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		amounts := make(map[int]uint64)
		err := util.CatchPanicReturnError(func() {
			for i := 0; i < runTimes; i++ {
				amount := uint64(rand.Intn(runTimes))
				idx := rand.Intn(numInitBalances)
				s, _ := amounts[idx]
				amounts[idx] = s + amount

				consumeUTXO(t, txb, 1, nativeTokensOnChain[idx].ID, amount)

				totalIotas, _, isBalanced := txb.TotalAssets()
				require.True(t, isBalanced)
				require.EqualValues(t, initialTotalIotas+i+1, totalIotas)
			}
		}, ErrInputLimitExceeded)
		require.True(t, xerrors.Is(err, ErrInputLimitExceeded))

		_, _, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
	})
}
