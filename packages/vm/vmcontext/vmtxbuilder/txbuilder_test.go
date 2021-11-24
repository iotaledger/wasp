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

const initialTotalIotas = 1000

func TestTxBuilderBasic(t *testing.T) {
	addr := tpkg.RandEd25519Address()
	stateMetadata := hashing.HashStrings("test")
	aliasID := rndAliasID()
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
	tokenID := testiotago.RandNativeTokenID()
	const initialTokenBalance = 100
	balanceLoader := func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
		if id == tokenID {
			return new(big.Int).SetUint64(initialTokenBalance), testiotago.RandUTXOInput()
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
		consumeUTXO(t, txb, 42, tokenID, 42)

		expectedBalance := new(big.Int).SetUint64(initialTokenBalance)
		expectedBalance.Add(expectedBalance, new(big.Int).SetUint64(42))

		totalIotas, totalTokens, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+42, totalIotas)
		require.EqualValues(t, 1, len(totalTokens))
		require.True(t, totalTokens[tokenID].Cmp(expectedBalance) == 0)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}

func TestTxBuilderConsistency(t *testing.T) {
	addr := tpkg.RandEd25519Address()
	stateMetadata := hashing.HashStrings("test")
	aliasID := rndAliasID()
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
		nativeTokensOnChain = append(nativeTokensOnChain, testiotago.RandNativeTokenAmount(2000+i*10))
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

	var txb *AnchorTransactionBuilder
	var amounts map[int]uint64

	runCreateBuilderAndConsume := func(n int) {
		txb = NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		amounts = make(map[int]uint64)
		for i := 0; i < n; i++ {
			idx := rand.Intn(numInitBalances)
			s, _ := amounts[idx]
			amounts[idx] = s + 10

			consumeUTXO(t, txb, 10, nativeTokensOnChain[idx].ID, 10)

			totalIotas, _, isBalanced := txb.TotalAssets()
			require.True(t, isBalanced)
			require.EqualValues(t, initialTotalIotas+(i+1)*10, totalIotas)
		}
	}
	runPostRequest := func(n int) {
		for i := 0; i < n; i++ {
			rndIndex := rand.Intn(numInitBalances)
			assets := &iscp.Assets{
				Iotas: 10,
				Tokens: iotago.NativeTokens{
					&iotago.NativeToken{
						ID:     nativeTokensOnChain[rndIndex].ID,
						Amount: new(big.Int).SetUint64(10),
					},
				},
			}
			txb.PostRequest(iscp.PostRequestData{
				TargetAddress:  tpkg.RandEd25519Address(),
				SenderContract: iscp.Hn("test"),
				Assets:         assets,
				Metadata:       &iscp.SendMetadata{},
				SendOptions:    nil,
				GasBudget:      0,
			})
		}
	}
	t.Run("consistency check 1", func(t *testing.T) {
		const runTimes = 100
		runCreateBuilderAndConsume(runTimes)

		totalIotas, totalTokens, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totalIotas))
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
		err := util.CatchPanicReturnError(func() {
			runCreateBuilderAndConsume(runTimes)
		}, ErrInputLimitExceeded)
		require.True(t, xerrors.Is(err, ErrInputLimitExceeded))

		_, _, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
	})
	t.Run("consistency check 2", func(t *testing.T) {
		const runTimes = 100
		runCreateBuilderAndConsume(runTimes)

		totalIotasBefore, totalAssetsBefore, ok := txb.TotalAssets()
		require.True(t, ok)

		runPostRequest(runTimes)

		totalIotasAfter, totalAssetsAfter, ok := txb.TotalAssets()
		require.True(t, ok)
		require.EqualValues(t, totalIotasBefore, totalIotasAfter)
		require.EqualValues(t, len(totalAssetsBefore), len(totalAssetsAfter))
		sumBefore := new(big.Int)
		sumAfter := new(big.Int)
		for id := range totalAssetsAfter {
			require.True(t, ok)
			sumBefore.Add(sumBefore, totalAssetsBefore[id])
			sumAfter.Add(sumAfter, totalAssetsAfter[id])
		}
		require.True(t, sumBefore.Cmp(sumAfter) == 0)
	})
	t.Run("exceeded outputs", func(t *testing.T) {
		const runTimesInputs = 100
		const runTimesOutputs = 150
		runCreateBuilderAndConsume(runTimesInputs)

		err := util.CatchPanicReturnError(func() {
			runPostRequest(runTimesOutputs)
		}, ErrOutputLimitExceeded)

		require.True(t, xerrors.Is(err, ErrOutputLimitExceeded))

		_, _, isBalanced := txb.TotalAssets()
		require.True(t, isBalanced)
	})
}
