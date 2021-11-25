package vmtxbuilder

import (
	"errors"
	"math/big"
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/testutil/testiotago"
	"github.com/stretchr/testify/require"
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
	_, _, isBalanced := txb.Totals()
	require.True(t, isBalanced)
}

func TestTxBuilderBasic(t *testing.T) {
	const initialTotalIotas = 1000
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
	balanceLoader := func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
		return nil, iotago.UTXOInput{}
	}
	t.Run("1", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
			return nil, iotago.UTXOInput{}
		})
		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, 1000, totals.totalIotas)
		require.EqualValues(t, 0, len(totals.tokenBalances))

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
		txb.addDeltaIotasToAnchor(42)
		_, _, isBalanced := txb.Totals()
		require.False(t, isBalanced)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("3", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		consumeUTXO(t, txb, 10, tokenID, 10)

		t.Logf("vByteCost internal output = %d", txb.vByteCostOfNativeTokenBalance())

		totalsIn, totalsOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, 0, totalsIn.internalDustDeposit)
		require.EqualValues(t, txb.vByteCostOfNativeTokenBalance(), totalsOut.internalDustDeposit)

		require.EqualValues(t, initialTotalIotas+10, int(totalsOut.totalIotas))
		require.EqualValues(t, initialTotalIotas+10-int(txb.vByteCostOfNativeTokenBalance()), txb.currentBalanceIotasOnAnchor)
		require.EqualValues(t, 1, len(totalsOut.tokenBalances))
		require.True(t, totalsOut.tokenBalances[tokenID].Cmp(new(big.Int).SetUint64(10)) == 0)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}

func TestTxBuilderConsistency(t *testing.T) {
	const initialTotalIotas = 1000
	const testAmount = 10

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

	// generate some random ids
	nativeTokenIDs := make([]iotago.NativeTokenID, 0)
	utxoInputsNativeTokens := make([]iotago.UTXOInput, 0)
	genNativeTokenIDs := func(n int) {
		for i := 0; i < n; i++ {
			nativeTokenIDs = append(nativeTokenIDs, testiotago.RandNativeTokenID())
			utxoInputsNativeTokens = append(utxoInputsNativeTokens, testiotago.RandUTXOInput())
		}
	}
	// all token accounts initially are empty
	balanceLoader := func(_ iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
		return nil, iotago.UTXOInput{}
	}

	var txb *AnchorTransactionBuilder
	var amounts map[int]uint64

	runCreateBuilderAndConsume := func(n int, numNativeTokens int) {
		txb = NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		amounts = make(map[int]uint64)

		for i := 0; i < n; i++ {
			idx := i % numNativeTokens
			s, _ := amounts[idx]
			amounts[idx] = s + 10

			consumeUTXO(t, txb, testAmount, nativeTokenIDs[idx], testAmount)

			totals, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
			require.EqualValues(t, initialTotalIotas+(i+1)*10, totals.totalIotas)
		}
	}
	runCreateBuilderAndConsumeRandomly := func(n int, numNativeTokens int) {
		txb = NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		amounts = make(map[int]uint64)

		for i := 0; i < n; i++ {
			idx := rand.Intn(numNativeTokens)
			s, _ := amounts[idx]
			amounts[idx] = s + 10

			consumeUTXO(t, txb, testAmount, nativeTokenIDs[idx], testAmount)

			totals, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
			require.EqualValues(t, initialTotalIotas+(i+1)*10, totals.totalIotas)
		}
	}
	runPostRequest := func(n int, numNativeTokens int) {
		for i := 0; i < n; i++ {
			idx := i % numNativeTokens
			assets := &iscp.Assets{
				Iotas: testAmount,
				Tokens: iotago.NativeTokens{
					&iotago.NativeToken{
						ID:     nativeTokenIDs[idx],
						Amount: new(big.Int).SetUint64(testAmount),
					},
				},
			}
			exout := ExtendedOutputFromPostData(
				tpkg.RandEd25519Address(),
				anchor.AliasID.ToAddress(),
				iscp.Hn("test"),
				assets,
				&iscp.SendMetadata{},
			)
			txb.AddOutput(exout)
		}
	}
	runPostRequest = runPostRequest
	t.Run("consistency check 1", func(t *testing.T) {
		const runTimes = 5
		const numNativeTokens = 5

		genNativeTokenIDs(numNativeTokens)
		runCreateBuilderAndConsume(runTimes, numNativeTokens)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.totalIotas))
		require.EqualValues(t, numNativeTokens, len(totals.tokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 2", func(t *testing.T) {
		const runTimes = 100
		const numNativeTokens = 5

		genNativeTokenIDs(numNativeTokens)
		runCreateBuilderAndConsume(runTimes, numNativeTokens)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.totalIotas))
		require.EqualValues(t, numNativeTokens, len(totals.tokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 3", func(t *testing.T) {
		const runTimes = 100
		const numNativeTokens = 5

		genNativeTokenIDs(numNativeTokens)
		runCreateBuilderAndConsumeRandomly(runTimes, numNativeTokens)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.totalIotas))
		require.EqualValues(t, numNativeTokens, len(totals.tokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 4", func(t *testing.T) {
		const runTimes = 100
		const numNativeTokens = 10

		genNativeTokenIDs(numNativeTokens)
		err := util.CatchPanicReturnError(func() {
			runCreateBuilderAndConsumeRandomly(runTimes, numNativeTokens)
		}, ErrNotEnoughFundsForInternalDustDeposit)
		require.True(t, errors.Is(err, ErrNotEnoughFundsForInternalDustDeposit))

		// the txb state left inconsistent
		_, _, isBalanced := txb.Totals()
		require.False(t, isBalanced)
	})
	//t.Run("exceed inputs", func(t *testing.T) {
	//	const runTimes = 150
	//	err := util.CatchPanicReturnError(func() {
	//		runCreateBuilderAndConsume(runTimes)
	//	}, ErrInputLimitExceeded)
	//	require.True(t, xerrors.Is(err, ErrInputLimitExceeded))
	//
	//	_, _, isBalanced := txb.Totals()
	//	require.True(t, isBalanced)
	//})
	//t.Run("consistency check 2", func(t *testing.T) {
	//	const runTimes = 100
	//	runCreateBuilderAndConsume(runTimes)
	//
	//	totalsBefore, _, ok := txb.Totals()
	//	require.True(t, ok)
	//
	//	runPostRequest(runTimes)
	//
	//	totalsAfter, _, ok := txb.Totals()
	//	require.True(t, ok)
	//	sumBefore := new(big.Int)
	//	sumAfter := new(big.Int)
	//	for id := range totalsAfter.tokenBalances {
	//		require.True(t, ok)
	//		sumBefore.Add(sumBefore, totalsAfter.tokenBalances[id])
	//		sumAfter.Add(sumAfter, totalsBefore.tokenBalances[id])
	//	}
	//	require.True(t, sumBefore.Cmp(sumAfter) == 0)
	//})
	//t.Run("exceeded outputs", func(t *testing.T) {
	//	const runTimesInputs = 100
	//	const runTimesOutputs = 150
	//	runCreateBuilderAndConsume(runTimesInputs)
	//
	//	err := util.CatchPanicReturnError(func() {
	//		runPostRequest(runTimesOutputs)
	//	}, ErrOutputLimitExceeded)
	//
	//	require.True(t, xerrors.Is(err, ErrOutputLimitExceeded))
	//
	//	_, _, isBalanced := txb.Totals()
	//	require.True(t, isBalanced)
	//})
}
