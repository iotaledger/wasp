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
	reqData, err := iscp.OnLedgerFromUTXO(&iscp.UTXOMetaData{}, out)
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
		tpkg.RandEd25519Address(),
		txb.anchorOutput.AliasID.ToAddress(),
		iscp.Hn("test"),
		assets,
		&iscp.SendMetadata{},
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
		require.EqualValues(t, 1000, totals.TotalIotas)
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
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
			return nil, iotago.UTXOInput{}
		})
		txb.addDeltaIotasToAnchor(42)
		_, _, isBalanced := txb.Totals()
		require.False(t, isBalanced)
		//
		//essence := txb.BuildTransactionEssence(&iscp.StateData{})
		//
		//essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		//require.NoError(t, err)
		//t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("3", func(t *testing.T) {
		txb := NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		_, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		consumeUTXO(t, txb, 10, tokenID, 10)

		t.Logf("vByteCost internal output = %d", txb.vByteCostOfNativeTokenBalance())

		totalsIn, totalsOut, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, 0, totalsIn.InternalDustDeposit)
		require.EqualValues(t, txb.vByteCostOfNativeTokenBalance(), totalsOut.InternalDustDeposit)

		require.EqualValues(t, initialTotalIotas+10, int(totalsOut.TotalIotas))
		require.EqualValues(t, initialTotalIotas+10-int(txb.vByteCostOfNativeTokenBalance()), txb.currentBalanceIotasOnAnchor)
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
	anchorID, _ := tpkg.RandUTXOInput()

	var nativeTokenIDs []iotago.NativeTokenID
	var utxoInputsNativeTokens []iotago.UTXOInput
	// all token accounts initially are empty
	balanceLoader := func(_ iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
		return nil, iotago.UTXOInput{}
	}

	var txb *AnchorTransactionBuilder
	var amounts map[int]uint64

	var numTokenIDs int

	initTest := func() {
		txb = NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
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
			require.EqualValues(t, initialTotalIotas+uint64(i+1)*amount, totals.TotalIotas)
		}
	}
	runCreateBuilderAndConsumeRandomly := func(numRun int, amount uint64) {
		txb = NewAnchorTransactionBuilder(anchor, *anchorID, anchor.Amount, balanceLoader)
		amounts = make(map[int]uint64)

		for i := 0; i < numRun; i++ {
			idx := rand.Intn(numTokenIDs)
			s, _ := amounts[idx]
			amounts[idx] = s + 10

			consumeUTXO(t, txb, amount, nativeTokenIDs[idx], amount)

			totals, _, isBalanced := txb.Totals()
			require.True(t, isBalanced)
			require.EqualValues(t, initialTotalIotas+(i+1)*10, totals.TotalIotas)
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
		numTokenIDs = 5

		initTest()
		runConsume(runTimes, testAmount)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.TotalIotas))
		require.EqualValues(t, numTokenIDs, len(totals.TokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 2", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 10
		numTokenIDs = 5

		initTest()
		runConsume(runTimes, testAmount)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.TotalIotas))
		require.EqualValues(t, numTokenIDs, len(totals.TokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 3", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 10
		numTokenIDs = 5

		initTest()
		runCreateBuilderAndConsumeRandomly(runTimes, testAmount)

		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)
		require.EqualValues(t, initialTotalIotas+runTimes*10, int(totals.TotalIotas))
		require.EqualValues(t, numTokenIDs, len(totals.TokenBalances))

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("consistency check 4", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 10
		numTokenIDs = 10

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
		numTokenIDs = 5

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
		runConsume(runTimesInputs, testAmount+1)

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
		runConsume(runTimesInputs, testAmount+1)

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
	t.Run("overflow 2", func(t *testing.T) {
		const runTimesInputs = 120
		const runTimesOutputs = 130
		const testAmount = 10
		numTokenIDs = 5

		initTest()
		runConsume(runTimesInputs, testAmount+1)

		err := util.CatchPanicReturnError(func() {
			runPostRequestRandomly(runTimesOutputs, testAmount)
		}, ErrOverflow)

		require.Error(t, err, ErrOverflow)

		_, _, isBalanced := txb.Totals()
		require.False(t, isBalanced)
	})
	t.Run("randomize", func(t *testing.T) {
		const runTimes = 50
		numTokenIDs = 5

		initTest()
		for _, id := range nativeTokenIDs {
			consumeUTXO(t, txb, 200, id, 10)
		}
		expectedIotas := initialTotalIotas + numTokenIDs*(200-int(txb.vByteCostOfNativeTokenBalance()))
		require.EqualValues(t, int(expectedIotas), int(txb.currentBalanceIotasOnAnchor))

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
		require.EqualValues(t, 0, int(totalsIN.InternalDustDeposit))
		require.EqualValues(t, numTokenIDs*int(txb.vByteCostOfNativeTokenBalance()), int(totalsOUT.InternalDustDeposit))
		require.EqualValues(t, int(expectedIotas), int(txb.currentBalanceIotasOnAnchor))

		t.Logf(">>>>>>>>>> \n%s", txb.String())

		essence := txb.BuildTransactionEssence(&iscp.StateData{})

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("clone", func(t *testing.T) {
		const runTimes = 100
		numTokenIDs = 5

		initTest()
		for _, id := range nativeTokenIDs {
			consumeUTXO(t, txb, runTimes, id, runTimes)
		}
		totals, _, isBalanced := txb.Totals()
		require.True(t, isBalanced)

		txbClone := txb.Clone()
		totalsClone, _, isBalanced := txbClone.Totals()
		require.True(t, isBalanced)
		require.True(t, totals.Equal(totalsClone))

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
		require.True(t, totals.Equal(totalsClone))
	})
}
