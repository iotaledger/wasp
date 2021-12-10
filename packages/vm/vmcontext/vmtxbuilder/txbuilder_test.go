package vmtxbuilder

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testiotago"
	"github.com/iotaledger/wasp/packages/util"
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
	reqData, err := iscp.OnLedgerFromUTXO(out, &iotago.UTXOInput{})
	require.NoError(t, err)
	txb.Consume(reqData)
	_, _, isBalanced := txb.Totals()
	require.True(t, isBalanced)
}

func addOutput(txb *AnchorTransactionBuilder, amount uint64, tokenID iotago.NativeTokenID) {
	assets := &iscp.Assets{
		Iotas: amount,
		Tokens: iotago.NativeTokens{
			&iotago.NativeToken{
				ID:     tokenID,
				Amount: new(big.Int).SetUint64(amount),
			},
		},
	}
	exout := ExtendedOutputFromPostData(
		txb.anchorOutput.AliasID.ToAddress(),
		iscp.Hn("test"),
		iscp.RequestParameters{
			TargetAddress: tpkg.RandEd25519Address(),
			Assets:        assets,
			Metadata:      &iscp.SendMetadata{},
			Options:       nil,
		},
	)
	txb.AddOutput(exout)
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
	anchorID := tpkg.RandUTXOInput()
	tokenID := testiotago.RandNativeTokenID()
	balanceLoader := func(id iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
		return nil, &iotago.UTXOInput{}
	}
	t.Run("1", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, anchorID, anchor.Amount, func(id iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
			return nil, nil
		})
		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, 1000-anchor.VByteCost(parameters.RentStructure(), nil), totals.TotalIotasOnChain)
		require.EqualValues(t, 0, len(totals.TokenBalances))

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
		txb := NewAnchorTransactionBuilder(anchor, anchorID, anchor.Amount, func(id iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
			return nil, nil
		})
		txb.addDeltaIotasToTotal(42)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
	})
	t.Run("3", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, anchorID, anchor.Amount, balanceLoader)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		consumeUTXO(t, txb, 10, tokenID, 10)

		t.Logf("vByteCost internal output = %d", txb.vByteCostOfNativeTokenBalance())

		totalsIn, totalsOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, 0, totalsIn.TotalIotasInDustDeposit)
		require.EqualValues(t, txb.vByteCostOfNativeTokenBalance(), totalsOut.TotalIotasInDustDeposit)

		require.EqualValues(t, initialTotalIotas+10-int(txb.vByteCostOfNativeTokenBalance()), int(totalsOut.TotalIotasOnChain))
		require.EqualValues(t, 1, len(totalsOut.TokenBalances))
		require.True(t, totalsOut.TokenBalances[tokenID].Cmp(new(big.Int).SetUint64(10)) == 0)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}

func TestTxBuilderConsistency(t *testing.T) {
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
	anchorID := tpkg.RandUTXOInput()

	var nativeTokenIDs []iotago.NativeTokenID
	var utxoInputsNativeTokens []iotago.UTXOInput
	// all token accounts initially are empty
	balanceLoader := func(_ iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
		return nil, &iotago.UTXOInput{}
	}

	initialBalance := new(big.Int)
	balanceLoaderWithInitialBalance := func(id iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
		for _, id1 := range nativeTokenIDs {
			if id == id1 {
				return new(big.Int).Set(initialBalance), &iotago.UTXOInput{}
			}
		}
		return nil, &iotago.UTXOInput{}
	}

	var txb *AnchorTransactionBuilder
	var amounts map[int]uint64

	var numTokenIDs int

	initTest := func() {
		txb = NewAnchorTransactionBuilder(anchor, anchorID, anchor.Amount, balanceLoader)
		amounts = make(map[int]uint64)

		nativeTokenIDs = make([]iotago.NativeTokenID, 0)
		utxoInputsNativeTokens = make([]iotago.UTXOInput, 0)

		for i := 0; i < numTokenIDs; i++ {
			nativeTokenIDs = append(nativeTokenIDs, testiotago.RandNativeTokenID())
			utxoInputsNativeTokens = append(utxoInputsNativeTokens, testiotago.RandUTXOInput())
		}
	}
	runConsume := func(numRun int, amount uint64) {
		for i := 0; i < numRun; i++ {
			idx := i % numTokenIDs
			s, _ := amounts[idx]
			amounts[idx] = s + amount

			consumeUTXO(t, txb, amount, nativeTokenIDs[idx], amount)

			totals, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
			require.EqualValues(t, initialTotalIotas+uint64(i+1)*amount, totals.TotalIotasOnChain)
		}
	}
	runCreateBuilderAndConsumeRandomly := func(numRun int, amount uint64) {
		txb = NewAnchorTransactionBuilder(anchor, anchorID, anchor.Amount, balanceLoader)
		amounts = make(map[int]uint64)

		for i := 0; i < numRun; i++ {
			idx := rand.Intn(numTokenIDs)
			s, _ := amounts[idx]
			amounts[idx] = s + 10

			consumeUTXO(t, txb, amount, nativeTokenIDs[idx], amount)

			totals, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
			require.EqualValues(t, initialTotalIotas+(i+1)*10, totals.TotalIotasOnChain)
		}
	}
	runPostRequest := func(n int, amount uint64) {
		for i := 0; i < n; i++ {
			idx := i % numTokenIDs
			addOutput(txb, amount, nativeTokenIDs[idx])
		}
	}
	runPostRequestRandomly := func(n int, amount uint64) {
		for i := 0; i < n; i++ {
			idx := rand.Intn(numTokenIDs)
			addOutput(txb, amount, nativeTokenIDs[idx])
		}
	}

	t.Run("consistency check 1", func(t *testing.T) {
		const runTimes = 5
		const testAmount = 10
		numTokenIDs = 4

		initTest()
		runConsume(runTimes, testAmount)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.TotalIotasOnChain))
		require.EqualValues(t, numTokenIDs, len(totals.TokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 2", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 10
		numTokenIDs = 4

		initTest()
		runConsume(runTimes, testAmount)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.TotalIotasOnChain))
		require.EqualValues(t, numTokenIDs, len(totals.TokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 3", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 10
		numTokenIDs = 4

		initTest()
		runCreateBuilderAndConsumeRandomly(runTimes, testAmount)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.TotalIotasOnChain))
		require.EqualValues(t, numTokenIDs, len(totals.TokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 4", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 10
		numTokenIDs = 6

		initTest()
		err := util.CatchPanicReturnError(func() {
			runCreateBuilderAndConsumeRandomly(runTimes, testAmount)
		}, ErrNotEnoughFundsForInternalDustDeposit)
		require.Error(t, err, ErrNotEnoughFundsForInternalDustDeposit)

		// the txb state left inconsistent
		_, _, isBalanced := txb.Totals()
		require.False(t, isBalanced)
	})
	t.Run("exceed inputs", func(t *testing.T) {
		const runTimes = 150
		const testAmount = 10
		numTokenIDs = 4

		initTest()
		err := util.CatchPanicReturnError(func() {
			runConsume(runTimes, testAmount)
		}, ErrInputLimitExceeded)
		require.Error(t, err, ErrInputLimitExceeded)

		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("exceeded outputs 1", func(t *testing.T) {
		const runTimesInputs = 120
		const runTimesOutputs = 130
		const testAmount = 1
		numTokenIDs = 5

		initTest()
		runConsume(runTimesInputs, testAmount+1000)

		err := util.CatchPanicReturnError(func() {
			runPostRequest(runTimesOutputs, testAmount)
		}, ErrOutputLimitExceeded)

		require.Error(t, err, ErrOutputLimitExceeded)

		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("exceeded outputs 2", func(t *testing.T) {
		const runTimesInputs = 120
		const runTimesOutputs = 130
		const testAmount = 1
		numTokenIDs = 5

		initTest()
		runConsume(runTimesInputs, testAmount+1000)

		err := util.CatchPanicReturnError(func() {
			runPostRequestRandomly(runTimesOutputs, testAmount)
		}, ErrOutputLimitExceeded)

		require.Error(t, err, ErrOutputLimitExceeded)

		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("not enough for dust", func(t *testing.T) {
		const runTimesInputs = 10
		const testAmount = 10
		numTokenIDs = 5

		initTest()
		err := util.CatchPanicReturnError(func() {
			runConsume(runTimesInputs, testAmount)
		}, ErrNotEnoughFundsForInternalDustDeposit)

		require.Error(t, err, ErrNotEnoughFundsForInternalDustDeposit)

		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
	})
	t.Run("randomize", func(t *testing.T) {
		const runTimes = 30
		numTokenIDs = 5

		initTest()
		for _, id := range nativeTokenIDs {
			consumeUTXO(t, txb, 5000, id, 10)
		}
		dustDeposit := txb.vByteCostOfNativeTokenBalance()
		expectedIotas := initialTotalIotas + numTokenIDs*(5000-int(dustDeposit))
		require.EqualValues(t, expectedIotas, int(txb.currentBalanceIotasOnAnchor))

		for i := 0; i < runTimes; i++ {
			idx1 := rand.Intn(numTokenIDs)
			consumeUTXO(t, txb, 1, nativeTokenIDs[idx1], 1)
			idx2 := rand.Intn(numTokenIDs)
			addOutput(txb, 1, nativeTokenIDs[idx2])
			_, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
		}
		totalsIN, totalsOUT, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, 0, int(totalsIN.TotalIotasInDustDeposit))
		require.EqualValues(t, numTokenIDs*int(dustDeposit), int(totalsOUT.TotalIotasInDustDeposit))

		t.Logf(">>>>>>>>>> \n%s", txb.String())

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("clone", func(t *testing.T) {
		const runTimes = 7
		numTokenIDs = 5

		initTest()
		for _, id := range nativeTokenIDs {
			consumeUTXO(t, txb, 1000, id, 100)
		}
		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)

		txbClone := txb.Clone()
		totalsClone, _, isBalanced := txbClone.Totals()
		require.True(t, isBalanced)
		require.True(t, totals.BalancedWith(totalsClone))

		for i := 0; i < runTimes; i++ {
			idx1 := rand.Intn(numTokenIDs)
			consumeUTXO(t, txb, 1, nativeTokenIDs[idx1], 1)
			idx2 := rand.Intn(numTokenIDs)
			addOutput(txb, 1, nativeTokenIDs[idx2])
			_, _, isBalanced = txb.Totals()
			require.True(t, isBalanced)
		}

		totalsClone, _, isBalanced = txbClone.Totals()
		require.True(t, isBalanced)
		require.True(t, totals.BalancedWith(totalsClone))
	})
	t.Run("initial balance 1", func(t *testing.T) {
		numTokenIDs = 5

		initialBalance.SetUint64(100)
		balanceLoader = balanceLoaderWithInitialBalance
		initTest()

		// send 90 < 100 which is on-chain. 10 must be left and dust deposit should not disappear
		addOutput(txb, 90, nativeTokenIDs[0])

		totalIn, totalOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas, int(totalOut.TotalIotasOnChain))
		require.EqualValues(t, int(txb.vByteCostOfNativeTokenBalance()), int(totalIn.TotalIotasInDustDeposit))
		require.EqualValues(t, int(txb.vByteCostOfNativeTokenBalance()), int(totalOut.TotalIotasInDustDeposit))
		beforeTokens, afterTokens := txb.InternalNativeTokenBalances()

		require.True(t, beforeTokens[nativeTokenIDs[0]].Cmp(new(big.Int).SetInt64(100)) == 0)
		require.True(t, afterTokens[nativeTokenIDs[0]].Cmp(new(big.Int).SetInt64(10)) == 0)
		t.Logf(">>>>>>>>>> \n%s", txb.String())

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("initial balance 2", func(t *testing.T) {
		numTokenIDs = 5

		initialBalance.SetUint64(100)
		balanceLoader = balanceLoaderWithInitialBalance
		initTest()

		// send 90 < 100 which is on-chain. 10 must be left and dust deposit should not disappear
		addOutput(txb, 100, nativeTokenIDs[0])

		totalIn, totalOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+txb.vByteCostOfNativeTokenBalance(), int(totalOut.TotalIotasOnChain))
		require.EqualValues(t, int(txb.vByteCostOfNativeTokenBalance()), int(totalIn.TotalIotasInDustDeposit))
		require.EqualValues(t, 0, int(totalOut.TotalIotasInDustDeposit))
		beforeTokens, afterTokens := txb.InternalNativeTokenBalances()

		require.True(t, beforeTokens[nativeTokenIDs[0]].Cmp(new(big.Int).SetInt64(100)) == 0)
		_, ok := afterTokens[nativeTokenIDs[0]]
		require.False(t, ok)

		t.Logf(">>>>>>>>>> \n%s", txb.String())

		essence := txb.BuildTransactionEssence(&iscp.StateData{})
		require.EqualValues(t, 2, len(essence.Inputs))
		require.EqualValues(t, 2, len(essence.Outputs))

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("initial balance 3", func(t *testing.T) {
		numTokenIDs = 5

		initialBalance.SetUint64(100)
		balanceLoader = balanceLoaderWithInitialBalance
		initTest()

		// send 90 < 100 which is on-chain. 10 must be left and dust deposit should not disappear
		for i := range nativeTokenIDs {
			addOutput(txb, 100, nativeTokenIDs[i])
		}

		totalIn, totalOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		expectedIotas := initialTotalIotas + int(txb.vByteCostOfNativeTokenBalance())*len(nativeTokenIDs)
		require.EqualValues(t, expectedIotas, int(totalOut.TotalIotasOnChain))
		require.EqualValues(t, int(txb.vByteCostOfNativeTokenBalance())*len(nativeTokenIDs), int(totalIn.TotalIotasInDustDeposit))
		require.EqualValues(t, 0, int(totalOut.TotalIotasInDustDeposit))
		beforeTokens, afterTokens := txb.InternalNativeTokenBalances()

		for i := range nativeTokenIDs {
			require.True(t, beforeTokens[nativeTokenIDs[i]].Cmp(new(big.Int).SetInt64(100)) == 0)
			_, ok := afterTokens[nativeTokenIDs[i]]
			require.False(t, ok)
		}

		t.Logf(">>>>>>>>>> \n%s", txb.String())

		essence := txb.BuildTransactionEssence(&iscp.StateData{})
		require.EqualValues(t, 6, len(essence.Inputs))
		require.EqualValues(t, 6, len(essence.Outputs))

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}

func TestDustDeposit(t *testing.T) {
	reqMetadata := iscp.RequestMetadata{
		SenderContract: 0,
		TargetContract: 0,
		EntryPoint:     0,
		Params:         dict.New(),
		Transfer:       iscp.NewEmptyAssets(),
		GasBudget:      0,
	}

	t.Run("adjusts the output amount to the correct bytecost when needed", func(t *testing.T) {
		assets := iscp.NewEmptyAssets()
		out, wasAdjusted := NewExtendedOutput(
			&iotago.Ed25519Address{},
			assets,
			&iotago.Ed25519Address{1, 2, 3},
			&reqMetadata,
			nil,
		)
		require.True(t, wasAdjusted)
		require.Equal(t, out.Amount, out.VByteCost(parameters.RentStructure(), nil))
	})
	t.Run("keeps the same amount of iotas when enough for dust cost", func(t *testing.T) {
		assets := iscp.NewAssets(10000, nil)
		out, wasAdjusted := NewExtendedOutput(
			&iotago.Ed25519Address{},
			assets,
			&iotago.Ed25519Address{1, 2, 3},
			&reqMetadata,
			nil,
		)
		require.False(t, wasAdjusted)
		require.GreaterOrEqual(t, out.Amount, out.VByteCost(parameters.RentStructure(), nil))
	})
}
