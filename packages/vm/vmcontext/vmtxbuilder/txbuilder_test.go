package vmtxbuilder

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/testutil/testdeserparams"

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

// return deposit in iotas
func consumeUTXO(t *testing.T, txb *AnchorTransactionBuilder, id iotago.NativeTokenID, amountNative uint64, addIotasToDustMinimum ...uint64) uint64 {
	var assets *iscp.Assets
	if amountNative > 0 {
		assets = &iscp.Assets{
			Iotas:  0,
			Tokens: iotago.NativeTokens{{id, big.NewInt(int64(amountNative))}},
		}
	}
	out, _ := MakeExtendedOutput(
		txb.anchorOutput.AliasID.ToAddress(),
		nil,
		assets,
		nil,
		nil,
		testdeserparams.RentStructure(),
	)
	if len(addIotasToDustMinimum) > 0 {
		out.Amount += addIotasToDustMinimum[0]
	}
	reqData, err := iscp.OnLedgerFromUTXO(out, &iotago.UTXOInput{})
	require.NoError(t, err)
	txb.Consume(reqData)
	_, _, isBalanced := txb.Totals()
	require.True(t, isBalanced)
	return out.Deposit()
}

func addOutput(txb *AnchorTransactionBuilder, amount uint64, tokenID iotago.NativeTokenID) uint64 {
	assets := &iscp.Assets{
		Iotas: 0,
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
		testdeserparams.DeSerializationParameters().RentStructure,
	)
	txb.AddOutput(exout)
	_, _, isBalanced := txb.Totals()
	if !isBalanced {
		panic("not balanced txbuilder")
	}
	return exout.Deposit()
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
	balanceLoader := func(_ *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
		return nil, &iotago.UTXOInput{}
	}
	t.Run("1", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, anchorID, func(id *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
			return nil, nil
		},
			nil,
			*NewDepositEstimate(testdeserparams.RentStructure()),
			testdeserparams.RentStructure(),
		)
		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, 1000-txb.dustDepositAssumption.AnchorOutput, totals.TotalIotasOnChain)
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
		txb := NewAnchorTransactionBuilder(anchor, anchorID, func(id *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
			return nil, nil
		},
			nil,
			*NewDepositEstimate(testdeserparams.RentStructure()),
			testdeserparams.RentStructure(),
		)
		txb.addDeltaIotasToTotal(42)
		require.EqualValues(t, int(initialTotalIotas-txb.dustDepositAssumption.AnchorOutput+42), int(txb.totalIotasInContracts))
		_, _, isBalanced := txb.Totals()
		require.False(t, isBalanced)
	})
	t.Run("3", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(
			anchor, anchorID, balanceLoader, nil,
			*NewDepositEstimate(testdeserparams.RentStructure()),
			testdeserparams.RentStructure(),
		)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		deposit := consumeUTXO(t, txb, tokenID, 0)

		t.Logf("vByteCost anchor: %d, internal output: %d, 'empty' output deposit: %d",
			txb.dustDepositAssumption.AnchorOutput, txb.dustDepositAssumption.NativeTokenOutput, deposit)

		totalsIn, totalsOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, txb.dustDepositAssumption.AnchorOutput, totalsIn.TotalIotasInDustDeposit)
		require.EqualValues(t, txb.dustDepositAssumption.AnchorOutput, totalsOut.TotalIotasInDustDeposit)

		expectedIotas := initialTotalIotas - txb.dustDepositAssumption.AnchorOutput + deposit
		require.EqualValues(t, expectedIotas, int(totalsOut.TotalIotasOnChain))
		require.EqualValues(t, 0, len(totalsOut.TokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("4", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, anchorID, balanceLoader, nil,
			*NewDepositEstimate(testdeserparams.RentStructure()),
			testdeserparams.RentStructure(),
		)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		deposit := consumeUTXO(t, txb, tokenID, 10)

		t.Logf("vByteCost anchor: %d, internal output: %d",
			txb.dustDepositAssumption.AnchorOutput, txb.dustDepositAssumption.NativeTokenOutput)

		totalsIn, totalsOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, int(txb.dustDepositAssumption.AnchorOutput), int(totalsIn.TotalIotasInDustDeposit))
		require.EqualValues(t, int(txb.dustDepositAssumption.AnchorOutput+txb.dustDepositAssumption.NativeTokenOutput), int(totalsOut.TotalIotasInDustDeposit))

		expectedIotas := initialTotalIotas + deposit - txb.dustDepositAssumption.AnchorOutput - txb.dustDepositAssumption.NativeTokenOutput
		require.EqualValues(t, int(expectedIotas), int(totalsOut.TotalIotasOnChain))
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
	balanceLoader := func(_ *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
		return nil, &iotago.UTXOInput{}
	}

	var txb *AnchorTransactionBuilder
	var amounts map[int]uint64

	initialBalance := new(big.Int)
	balanceLoaderWithInitialBalance := func(id *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
		for _, id1 := range nativeTokenIDs {
			if *id == id1 {
				ret := txb.newInternalTokenOutput(aliasID, *id)
				ret.NativeTokens[0].Amount = new(big.Int).Set(initialBalance)
				return ret, &iotago.UTXOInput{}
			}
		}
		return nil, &iotago.UTXOInput{}
	}

	var numTokenIDs int

	initTest := func() {
		txb = NewAnchorTransactionBuilder(anchor, anchorID, balanceLoader, nil,
			*NewDepositEstimate(testdeserparams.RentStructure()),
			testdeserparams.RentStructure(),
		)
		amounts = make(map[int]uint64)

		nativeTokenIDs = make([]iotago.NativeTokenID, 0)
		utxoInputsNativeTokens = make([]iotago.UTXOInput, 0)

		for i := 0; i < numTokenIDs; i++ {
			nativeTokenIDs = append(nativeTokenIDs, testiotago.RandNativeTokenID())
			utxoInputsNativeTokens = append(utxoInputsNativeTokens, testiotago.RandUTXOInput())
		}
	}
	runConsume := func(numRun int, amountNative uint64, addIotasToDustMinimum ...uint64) {
		deposit := uint64(0)
		for i := 0; i < numRun; i++ {
			idx := i % numTokenIDs
			s := amounts[idx]
			amounts[idx] = s + amountNative

			deposit += consumeUTXO(t, txb, nativeTokenIDs[idx], amountNative, addIotasToDustMinimum...)

			_, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
		}
		sumIN, sumOUT, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		expectedDust := txb.dustDepositAssumption.AnchorOutput
		if numRun < numTokenIDs {
			expectedDust += uint64(numRun) * txb.dustDepositAssumption.NativeTokenOutput
		} else {
			expectedDust += uint64(numTokenIDs) * txb.dustDepositAssumption.NativeTokenOutput
		}
		require.EqualValues(t, int(txb.dustDepositAssumption.AnchorOutput), sumIN.TotalIotasInDustDeposit)
		require.EqualValues(t, int(expectedDust), sumOUT.TotalIotasInDustDeposit)
	}
	runCreateBuilderAndConsumeRandomly := func(numRun int, amount uint64) {
		txb = NewAnchorTransactionBuilder(anchor, anchorID, balanceLoader, nil,
			*NewDepositEstimate(testdeserparams.RentStructure()),
			testdeserparams.RentStructure(),
		)
		amounts = make(map[int]uint64)

		deposit := uint64(0)
		for i := 0; i < numRun; i++ {
			idx := rand.Intn(numTokenIDs)
			amounts[idx] += amount
			deposit += consumeUTXO(t, txb, nativeTokenIDs[idx], amount)

			_, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
		}
		sumIN, sumOUT, isBalanced := txb.Totals()
		require.True(t, isBalanced)

		expectedIotas := initialTotalIotas - int(txb.dustDepositAssumption.AnchorOutput) + int(deposit)
		require.EqualValues(t, expectedIotas, int(sumIN.TotalIotasOnChain))
		expectedIotas -= len(amounts) * int(txb.dustDepositAssumption.NativeTokenOutput)
		require.EqualValues(t, expectedIotas, int(sumOUT.TotalIotasOnChain))
	}

	runPostRequest := func(n int, amount uint64) uint64 {
		ret := uint64(0)
		for i := 0; i < n; i++ {
			idx := i % numTokenIDs
			ret += addOutput(txb, amount, nativeTokenIDs[idx])
			_, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
		}
		return ret
	}

	runPostRequestRandomly := func(n int, amount uint64) uint64 {
		ret := uint64(0)
		for i := 0; i < n; i++ {
			idx := rand.Intn(numTokenIDs)
			ret += addOutput(txb, amount, nativeTokenIDs[idx])
			_, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
		}
		return ret
	}

	t.Run("consistency check 0", func(t *testing.T) {
		const runTimes = 3
		const testAmount = 10
		numTokenIDs = 4

		initTest()
		runConsume(runTimes, testAmount)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 1", func(t *testing.T) {
		const runTimes = 7
		const testAmount = 10
		numTokenIDs = 4

		initTest()
		runConsume(runTimes, testAmount)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 2", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 100
		numTokenIDs = 5

		initTest()
		runConsume(runTimes, testAmount)
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

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
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
		const runTimesInputs = 100
		const runTimesOutputs = 130
		numTokenIDs = 5

		initTest()
		runConsume(runTimesInputs, 10, 1000)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)

		err := util.CatchPanicReturnError(func() {
			runPostRequest(runTimesOutputs, 1)
		}, ErrOutputLimitExceeded)

		require.Error(t, err, ErrOutputLimitExceeded)

		_, _, isBalanced = txb.Totals()
		require.True(t, isBalanced)
		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("exceeded outputs 2", func(t *testing.T) {
		const runTimesInputs = 120
		const runTimesOutputs = 130
		numTokenIDs = 5

		initTest()
		runConsume(runTimesInputs, 10, 1000)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)

		err := util.CatchPanicReturnError(func() {
			runPostRequestRandomly(runTimesOutputs, 1)
		}, ErrOutputLimitExceeded)

		require.Error(t, err, ErrOutputLimitExceeded)

		_, _, isBalanced = txb.Totals()
		require.True(t, isBalanced)
		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("randomize", func(t *testing.T) {
		const runTimes = 30
		numTokenIDs = 5

		initTest()
		for _, id := range nativeTokenIDs {
			consumeUTXO(t, txb, id, 10)
		}

		for i := 0; i < runTimes; i++ {
			idx1 := rand.Intn(numTokenIDs)
			consumeUTXO(t, txb, nativeTokenIDs[idx1], 1, 1000)
			idx2 := rand.Intn(numTokenIDs)
			addOutput(txb, 1, nativeTokenIDs[idx2])
			_, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
		}
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)

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
			consumeUTXO(t, txb, id, 100)
		}
		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)

		txbClone := txb.Clone()
		totalsClone, _, isBalanced := txbClone.Totals()
		require.True(t, isBalanced)
		require.True(t, totals.BalancedWith(totalsClone))

		for i := 0; i < runTimes; i++ {
			idx1 := rand.Intn(numTokenIDs)
			consumeUTXO(t, txb, nativeTokenIDs[idx1], 1, 100)
			idx2 := rand.Intn(numTokenIDs)
			addOutput(txb, 1, nativeTokenIDs[idx2])
			_, _, isBalanced = txb.Totals()
			require.True(t, isBalanced)
		}

		totalsClone, _, isBalanced = txbClone.Totals()
		require.True(t, isBalanced)
		require.True(t, totals.BalancedWith(totalsClone))
	})
	t.Run("in balance 1", func(t *testing.T) {
		numTokenIDs = 5

		initialBalance.SetUint64(100)
		balanceLoader = balanceLoaderWithInitialBalance
		initTest()

		// send 90 < 100 which is on-chain. 10 must be left and dust deposit should not disappear
		addOutput(txb, 90, nativeTokenIDs[0])

		totalIn, totalOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas-txb.dustDepositAssumption.AnchorOutput, int(totalOut.TotalIotasOnChain))
		require.EqualValues(t, int(txb.dustDepositAssumption.NativeTokenOutput+txb.dustDepositAssumption.AnchorOutput), int(totalIn.TotalIotasInDustDeposit))
		require.EqualValues(t, int(txb.dustDepositAssumption.NativeTokenOutput+txb.dustDepositAssumption.AnchorOutput), int(totalOut.TotalIotasInDustDeposit))
		beforeTokens, afterTokens := txb.InternalNativeTokenBalances()

		require.True(t, beforeTokens[nativeTokenIDs[0]].Cmp(new(big.Int).SetInt64(100)) == 0)
		require.True(t, afterTokens[nativeTokenIDs[0]].Cmp(new(big.Int).SetInt64(10)) == 0)
		t.Logf(">>>>>>>>>> \n%s", txb.String())

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("in balance 2", func(t *testing.T) {
		numTokenIDs = 5

		initialBalance.SetUint64(100)
		balanceLoader = balanceLoaderWithInitialBalance
		initTest()

		// output will close internal account
		sentOut := addOutput(txb, 100, nativeTokenIDs[0])

		totalIn, totalOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, int(txb.dustDepositAssumption.NativeTokenOutput+txb.dustDepositAssumption.AnchorOutput), int(totalIn.TotalIotasInDustDeposit))
		require.EqualValues(t, int(sentOut), totalOut.SentOutIotas)
		require.EqualValues(t, int(initialTotalIotas-txb.dustDepositAssumption.AnchorOutput-sentOut+txb.dustDepositAssumption.NativeTokenOutput), int(txb.totalIotasInContracts))
		require.EqualValues(t, txb.dustDepositAssumption.AnchorOutput, int(totalOut.TotalIotasInDustDeposit))
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
	t.Run("in balance 3", func(t *testing.T) {
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
		expectedIotas := initialTotalIotas - int(txb.dustDepositAssumption.AnchorOutput) + int(txb.dustDepositAssumption.NativeTokenOutput)*len(nativeTokenIDs)
		require.EqualValues(t, expectedIotas, int(totalOut.TotalIotasOnChain))
		require.EqualValues(t, int(txb.dustDepositAssumption.NativeTokenOutput)*len(nativeTokenIDs)+int(txb.dustDepositAssumption.AnchorOutput), int(totalIn.TotalIotasInDustDeposit))
		require.EqualValues(t, txb.dustDepositAssumption.AnchorOutput, int(totalOut.TotalIotasInDustDeposit))
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
	t.Run("calc dust assumptions", func(t *testing.T) {
		d := NewDepositEstimate(testdeserparams.RentStructure())
		t.Logf("dust deposit assumptions:\n%s", d.String())

		d1, err := InternalDustDepositAssumptionFromBytes(d.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, d.AnchorOutput, d1.AnchorOutput)
		require.EqualValues(t, d.NativeTokenOutput, d1.NativeTokenOutput)
	})
	t.Run("adjusts the output amount to the correct bytecost when needed", func(t *testing.T) {
		assets := iscp.NewEmptyAssets()
		out, wasAdjusted := MakeExtendedOutput(
			&iotago.Ed25519Address{},
			&iotago.Ed25519Address{1, 2, 3},
			assets,
			&reqMetadata,
			nil,
			testdeserparams.RentStructure(),
		)
		require.True(t, wasAdjusted)
		require.Equal(t, out.Amount, out.VByteCost(parameters.RentStructure(), nil))
	})
	t.Run("keeps the same amount of iotas when enough for dust cost", func(t *testing.T) {
		assets := iscp.NewAssets(10000, nil)
		out, wasAdjusted := MakeExtendedOutput(
			&iotago.Ed25519Address{},
			&iotago.Ed25519Address{1, 2, 3},
			assets,
			&reqMetadata,
			nil,
			testdeserparams.RentStructure(),
		)
		require.False(t, wasAdjusted)
		require.GreaterOrEqual(t, out.Amount, out.VByteCost(parameters.RentStructure(), nil))
	})
}

func TestFoundries(t *testing.T) {
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
	balanceLoader := func(_ *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
		return nil, &iotago.UTXOInput{}
	}
	var txb *AnchorTransactionBuilder

	var numTokenIDs int

	initTest := func() {
		txb = NewAnchorTransactionBuilder(anchor, anchorID, balanceLoader, nil,
			*NewDepositEstimate(testdeserparams.RentStructure()),
			testdeserparams.RentStructure(),
		)

		nativeTokenIDs = make([]iotago.NativeTokenID, 0)
		utxoInputsNativeTokens = make([]iotago.UTXOInput, 0)

		for i := 0; i < numTokenIDs; i++ {
			nativeTokenIDs = append(nativeTokenIDs, testiotago.RandNativeTokenID())
			utxoInputsNativeTokens = append(utxoInputsNativeTokens, testiotago.RandUTXOInput())
		}
	}
	createNFoundries := func(n int) {
		for i := 0; i < n; i++ {
			sn, _ := txb.CreateNewFoundry(&iotago.SimpleTokenScheme{}, iotago.TokenTag{}, big.NewInt(10_000_000))
			require.EqualValues(t, i+1, int(sn))

			tin, tout, balanced := txb.Totals()
			require.True(t, balanced)
			t.Logf("%d. total iotas IN: %d, total iotas OUT: %d", i, tin.TotalIotasOnChain, tout.TotalIotasOnChain)
			t.Logf("%d. dust deposit IN: %d, dust deposit OUT: %d", i, tin.TotalIotasInDustDeposit, tout.TotalIotasInDustDeposit)
			t.Logf("%d. num foundries: %d", i, txb.nextFoundrySerialNumber())
		}

	}
	t.Run("create foundry ok", func(t *testing.T) {
		initTest()
		createNFoundries(3)
		essence := txb.BuildTransactionEssence(&iscp.StateData{})
		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("create foundry not enough", func(t *testing.T) {
		initTest()
		err := util.CatchPanicReturnError(func() {
			createNFoundries(5)
		}, ErrNotEnoughFundsForInternalDustDeposit)
		require.Error(t, err, ErrNotEnoughFundsForInternalDustDeposit)

		essence := txb.BuildTransactionEssence(&iscp.StateData{})
		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}

func TestSerDe(t *testing.T) {
	t.Run("serde ExtendedOutput", func(t *testing.T) {
		reqMetadata := iscp.RequestMetadata{
			SenderContract: 0,
			TargetContract: 0,
			EntryPoint:     0,
			Params:         dict.New(),
			Transfer:       iscp.NewEmptyAssets(),
			GasBudget:      0,
		}
		assets := iscp.NewEmptyAssets()
		out, _ := MakeExtendedOutput(
			&iotago.Ed25519Address{},
			&iotago.Ed25519Address{1, 2, 3},
			assets,
			&reqMetadata,
			nil,
			testdeserparams.DeSerializationParameters().RentStructure,
		)
		data, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		outBack := &iotago.ExtendedOutput{}
		_, err = outBack.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		require.True(t, out.Address.Equal(outBack.Address))
		require.EqualValues(t, out.Amount, outBack.Amount)
		require.EqualValues(t, 0, len(outBack.NativeTokens))
		require.True(t, outBack.Blocks.Equal(out.Blocks))
	})
	t.Run("serde FoundryOutput", func(t *testing.T) {
		out := &iotago.FoundryOutput{
			Address:           tpkg.RandAliasAddress(),
			Amount:            1337,
			NativeTokens:      nil,
			SerialNumber:      5,
			TokenTag:          iotago.TokenTag{},
			CirculatingSupply: big.NewInt(200),
			MaximumSupply:     big.NewInt(2000),
			TokenScheme:       &iotago.SimpleTokenScheme{},
			Blocks:            nil,
		}
		data, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		outBack := &iotago.FoundryOutput{}
		_, err = outBack.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		require.True(t, identicalFoundries(out, outBack))
	})
}
