package vmtxbuilder

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testiotago"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

var dummyStateMetadata = []byte("foobar")

type mockAccountContractRead struct {
	assets             *isc.Assets
	nativeTokenOutputs map[iotago.NativeTokenID]*iotago.BasicOutput
}

func (m *mockAccountContractRead) Read() AccountsContractRead {
	return AccountsContractRead{
		NativeTokenOutput: func(id iotago.FoundryID) (*iotago.BasicOutput, iotago.OutputID) {
			return m.nativeTokenOutputs[id], iotago.OutputID{}
		},
		FoundryOutput: func(uint32) (*iotago.FoundryOutput, iotago.OutputID) {
			return nil, iotago.OutputID{}
		},
		NFTOutput: func(id iotago.NFTID) (*iotago.NFTOutput, iotago.OutputID) {
			return nil, iotago.OutputID{}
		},
		TotalFungibleTokens: func() *isc.Assets {
			return m.assets
		},
	}
}

func newMockAccountsContractRead(anchor *iotago.AliasOutput) *mockAccountContractRead {
	anchorMinSD := parameters.L1().Protocol.RentStructure.MinRent(anchor)
	assets := isc.NewAssetsBaseTokens(anchor.Deposit() - anchorMinSD)
	return &mockAccountContractRead{
		assets:             assets,
		nativeTokenOutputs: make(map[iotago.FoundryID]*iotago.BasicOutput),
	}
}

func TestTxBuilderBasic(t *testing.T) {
	const initialTotalBaseTokens = 10 * isc.Million
	addr := tpkg.RandEd25519Address()
	aliasID := testiotago.RandAliasID()
	anchor := &iotago.AliasOutput{
		Amount:       initialTotalBaseTokens,
		NativeTokens: nil,
		AliasID:      aliasID,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: addr},
			&iotago.GovernorAddressUnlockCondition{Address: addr},
		},
		StateIndex:     0,
		StateMetadata:  dummyStateMetadata,
		FoundryCounter: 0,
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: aliasID.ToAddress(),
			},
		},
	}
	anchorID := tpkg.RandOutputID(0)
	t.Run("deposits", func(t *testing.T) {
		mockedAccounts := newMockAccountsContractRead(anchor)
		txb := NewAnchorTransactionBuilder(
			anchor,
			anchorID,
			parameters.L1().Protocol.RentStructure.MinRent(anchor),
			mockedAccounts.Read(),
		)
		essence, _ := txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()
		require.EqualValues(t, 1, txb.numInputs())
		require.EqualValues(t, 1, txb.numOutputs())
		require.False(t, txb.InputsAreFull())
		require.False(t, txb.outputsAreFull())

		require.EqualValues(t, 1, len(essence.Inputs))
		require.EqualValues(t, 1, len(essence.Outputs))

		_, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)

		// consume a request that sends 1Mi funds
		req1, err := isc.OnLedgerFromUTXO(&iotago.BasicOutput{
			Amount: 1 * isc.Million,
		}, iotago.OutputID{})
		require.NoError(t, err)
		txb.Consume(req1)
		mockedAccounts.assets.AddBaseTokens(req1.Output().Deposit())

		essence, _ = txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()
		require.Len(t, essence.Outputs, 1)
		require.EqualValues(t, essence.Outputs[0].Deposit(), anchor.Deposit()+req1.Output().Deposit())

		// consume a request that sends 1Mi, 1 NFT, and 4 native tokens
		nftID := tpkg.RandNFTAddress().NFTID()
		nativeTokenID1 := testiotago.RandNativeTokenID()
		nativeTokenID2 := testiotago.RandNativeTokenID()
		nativeTokenID3 := testiotago.RandNativeTokenID()
		nativeTokenID4 := testiotago.RandNativeTokenID()

		req2, err := isc.OnLedgerFromUTXO(&iotago.NFTOutput{
			Amount: 1 * isc.Million,
			NFTID:  nftID,
			NativeTokens: []*iotago.NativeToken{
				{ID: nativeTokenID1, Amount: big.NewInt(1)},
				{ID: nativeTokenID2, Amount: big.NewInt(2)},
				{ID: nativeTokenID3, Amount: big.NewInt(3)},
				{ID: nativeTokenID4, Amount: big.NewInt(4)},
			},
		}, iotago.OutputID{})
		require.NoError(t, err)
		totalSDBaseTokensUsedToSplitAssets := txb.Consume(req2)

		// deduct SD costs of creating the internal accounting outputs
		mockedAccounts.assets.Add(req2.Assets())
		mockedAccounts.assets.Spend(isc.NewAssetsBaseTokens(totalSDBaseTokensUsedToSplitAssets))

		essence, _ = txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()
		require.Len(t, essence.Outputs, 6) // 1 anchor AO, 1 NFT internal Output, 4 NativeTokens internal outputs
		require.EqualValues(t, essence.Outputs[0].Deposit(), anchor.Deposit()+req1.Output().Deposit()+req2.Output().Deposit()-totalSDBaseTokensUsedToSplitAssets)
	})
}

func TestTxBuilderConsistency(t *testing.T) {
	const initialTotalBaseTokens = 10000 * isc.Million
	addr := tpkg.RandEd25519Address()
	aliasID := testiotago.RandAliasID()
	anchor := &iotago.AliasOutput{
		Amount:       initialTotalBaseTokens,
		NativeTokens: nil,
		AliasID:      aliasID,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: addr},
			&iotago.GovernorAddressUnlockCondition{Address: addr},
		},
		StateIndex:     0,
		StateMetadata:  dummyStateMetadata,
		FoundryCounter: 0,
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: aliasID.ToAddress(),
			},
		},
	}
	anchorID := tpkg.RandOutputID(0)

	initTest := func(numTokenIDs int) (*AnchorTransactionBuilder, *mockAccountContractRead, []iotago.NativeTokenID) {
		mockedAccounts := newMockAccountsContractRead(anchor)
		txb := NewAnchorTransactionBuilder(
			anchor,
			anchorID,
			parameters.L1().Protocol.RentStructure.MinRent(anchor),
			mockedAccounts.Read(),
		)

		nativeTokenIDs := make([]iotago.NativeTokenID, 0)
		for i := 0; i < numTokenIDs; i++ {
			nativeTokenIDs = append(nativeTokenIDs, testiotago.RandNativeTokenID())
		}
		return txb, mockedAccounts, nativeTokenIDs
	}

	// return deposit in BaseToken
	consumeUTXO := func(t *testing.T, txb *AnchorTransactionBuilder, id iotago.NativeTokenID, amountNative uint64, mockedAccounts *mockAccountContractRead) {
		out := transaction.MakeBasicOutput(
			txb.anchorOutput.AliasID.ToAddress(),
			nil,
			&isc.Assets{
				NativeTokens: iotago.NativeTokens{{ID: id, Amount: big.NewInt(int64(amountNative))}},
			},
			nil,
			isc.SendOptions{},
		)
		req, err := isc.OnLedgerFromUTXO(
			transaction.AdjustToMinimumStorageDeposit(out), iotago.OutputID{})
		require.NoError(t, err)
		sdCost := txb.Consume(req)
		mockedAccounts.assets.Add(req.Assets())
		mockedAccounts.assets.Spend(isc.NewAssetsBaseTokens(sdCost))
		txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()
	}

	addOutput := func(txb *AnchorTransactionBuilder, amount uint64, nativeTokenID iotago.NativeTokenID, mockedAccounts *mockAccountContractRead) {
		outAssets := &isc.Assets{
			BaseTokens: 1 * isc.Million,
			NativeTokens: iotago.NativeTokens{{
				ID:     nativeTokenID,
				Amount: new(big.Int).SetUint64(amount),
			}},
		}
		out := transaction.BasicOutputFromPostData(
			txb.anchorOutput.AliasID.ToAddress(),
			isc.ContractIdentityFromHname(isc.Hn("test")),
			isc.RequestParameters{
				TargetAddress:                 tpkg.RandEd25519Address(),
				Assets:                        outAssets,
				Metadata:                      &isc.SendMetadata{},
				Options:                       isc.SendOptions{},
				AdjustToMinimumStorageDeposit: true,
			},
		)
		sdAdjust := txb.AddOutput(out)
		if !mockedAccounts.assets.Spend(outAssets) {
			panic("out of balance in chain output")
		}
		if sdAdjust < 0 {
			mockedAccounts.assets.Spend(isc.NewAssetsBaseTokens(uint64(-sdAdjust)))
		} else {
			mockedAccounts.assets.AddBaseTokens(uint64(sdAdjust))
		}
		txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()
	}

	t.Run("consistency check", func(t *testing.T) {
		const runTimes = 100
		const testAmount = 10
		const numTokenIDs = 4

		txb, mockedAccounts, nativeTokenIDs := initTest(numTokenIDs)
		for i := 0; i < runTimes; i++ {
			idx := rand.Intn(numTokenIDs)
			consumeUTXO(t, txb, nativeTokenIDs[idx], testAmount, mockedAccounts)
		}

		essence, _ := txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})

	runConsume := func(txb *AnchorTransactionBuilder, nativeTokenIDs []iotago.NativeTokenID, numRun int, amountNative uint64, mockedAccounts *mockAccountContractRead) {
		for i := 0; i < numRun; i++ {
			idx := i % len(nativeTokenIDs)
			consumeUTXO(t, txb, nativeTokenIDs[idx], amountNative, mockedAccounts)
			txb.BuildTransactionEssence(dummyStateMetadata)
			txb.MustBalanced()
		}
	}

	t.Run("exceed inputs", func(t *testing.T) {
		const runTimes = 150
		const testAmount = 10
		const numTokenIDs = 4

		txb, mockedAccounts, nativeTokenIDs := initTest(numTokenIDs)
		err := panicutil.CatchPanicReturnError(func() {
			runConsume(txb, nativeTokenIDs, runTimes, testAmount, mockedAccounts)
		}, vmexceptions.ErrInputLimitExceeded)
		require.Error(t, err, vmexceptions.ErrInputLimitExceeded)
	})
	t.Run("exceeded outputs", func(t *testing.T) {
		const runTimesInputs = 120
		const runTimesOutputs = 130
		const numTokenIDs = 5

		txb, mockedAccounts, nativeTokenIDs := initTest(numTokenIDs)
		runConsume(txb, nativeTokenIDs, runTimesInputs, 10, mockedAccounts)
		txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()

		err := panicutil.CatchPanicReturnError(func() {
			for i := 0; i < runTimesOutputs; i++ {
				idx := rand.Intn(numTokenIDs)
				addOutput(txb, 1, nativeTokenIDs[idx], mockedAccounts)
			}
		}, vmexceptions.ErrOutputLimitExceeded)
		require.Error(t, err, vmexceptions.ErrOutputLimitExceeded)
	})
	t.Run("randomize", func(t *testing.T) {
		const runTimes = 30
		const numTokenIDs = 5

		txb, mockedAccounts, nativeTokenIDs := initTest(numTokenIDs)
		for _, id := range nativeTokenIDs {
			consumeUTXO(t, txb, id, 10, mockedAccounts)
		}

		for i := 0; i < runTimes; i++ {
			idx1 := rand.Intn(numTokenIDs)
			consumeUTXO(t, txb, nativeTokenIDs[idx1], 1, mockedAccounts)
			idx2 := rand.Intn(numTokenIDs)
			if mockedAccounts.assets.AmountNativeToken(nativeTokenIDs[idx2]).Uint64() > 0 {
				addOutput(txb, 1, nativeTokenIDs[idx2], mockedAccounts)
			}
		}
		essence, _ := txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
	t.Run("clone", func(t *testing.T) {
		const runTimes = 7
		const numTokenIDs = 5

		txb, mockedAccounts, nativeTokenIDs := initTest(numTokenIDs)
		for _, id := range nativeTokenIDs {
			consumeUTXO(t, txb, id, 100, mockedAccounts)
		}
		txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()

		txbClone := txb.Clone()
		txbClone.BuildTransactionEssence(dummyStateMetadata)

		for i := 0; i < runTimes; i++ {
			idx1 := rand.Intn(numTokenIDs)
			consumeUTXO(t, txb, nativeTokenIDs[idx1], 1, mockedAccounts)
			idx2 := rand.Intn(numTokenIDs)
			addOutput(txb, 1, nativeTokenIDs[idx2], mockedAccounts)
		}

		txbClone.BuildTransactionEssence(dummyStateMetadata)
	})
	t.Run("send some of the tokens in balance", func(t *testing.T) {
		txb, mockedAccounts, nativeTokenIDs := initTest(5)
		setNativeTokenAccountsBalance := func(id iotago.NativeTokenID, amount int64) {
			mockedAccounts.assets.AddNativeTokens(id, amount)
			// create internal accounting outputs with 0 base tokens (they must be updated in the output side)
			out := txb.newInternalTokenOutput(aliasID, id)
			out.NativeTokens[0].Amount = new(big.Int).SetInt64(amount)
			mockedAccounts.nativeTokenOutputs[id] = out
		}

		// send 90 < 100 which is on-chain. 10 must be left and storage deposit should not disappear
		for i := range nativeTokenIDs {
			setNativeTokenAccountsBalance(nativeTokenIDs[i], 100)
			addOutput(txb, 90, nativeTokenIDs[i], mockedAccounts)
		}
		essence, _ := txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()

		require.EqualValues(t, 6, len(essence.Inputs))
		require.EqualValues(t, 11, len(essence.Outputs)) // 6 + 5 internal outputs with the 10 remaining tokens

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})

	t.Run("test consistency - consume send out, consume again", func(t *testing.T) {
		txb, mockedAccounts, nativeTokenIDs := initTest(1)
		tokenID := nativeTokenIDs[0]
		consumeUTXO(t, txb, tokenID, 1, mockedAccounts)
		addOutput(txb, 1, tokenID, mockedAccounts)
		consumeUTXO(t, txb, tokenID, 1, mockedAccounts)

		essence, _ := txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()

		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}

func TestFoundries(t *testing.T) {
	const initialTotalBaseTokens = 10*isc.Million + governance.DefaultMinBaseTokensOnCommonAccount
	addr := tpkg.RandEd25519Address()
	aliasID := testiotago.RandAliasID()
	anchor := &iotago.AliasOutput{
		Amount:       initialTotalBaseTokens,
		NativeTokens: nil,
		AliasID:      aliasID,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: addr},
			&iotago.GovernorAddressUnlockCondition{Address: addr},
		},
		StateIndex:     0,
		StateMetadata:  dummyStateMetadata,
		FoundryCounter: 0,
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: aliasID.ToAddress(),
			},
		},
	}
	anchorID := tpkg.RandOutputID(0)

	var nativeTokenIDs []iotago.NativeTokenID
	var txb *AnchorTransactionBuilder
	var numTokenIDs int

	var mockedAccounts *mockAccountContractRead
	initTest := func() {
		mockedAccounts = newMockAccountsContractRead(anchor)
		txb = NewAnchorTransactionBuilder(
			anchor,
			anchorID,
			parameters.L1().Protocol.RentStructure.MinRent(anchor),
			mockedAccounts.Read(),
		)

		nativeTokenIDs = make([]iotago.NativeTokenID, 0)

		for i := 0; i < numTokenIDs; i++ {
			nativeTokenIDs = append(nativeTokenIDs, testiotago.RandNativeTokenID())
		}
	}
	createNFoundries := func(n int) {
		for i := 0; i < n; i++ {
			sn, storageDeposit := txb.CreateNewFoundry(
				&iotago.SimpleTokenScheme{MaximumSupply: big.NewInt(10_000_000), MeltedTokens: util.Big0, MintedTokens: util.Big0},
				nil,
			)
			require.EqualValues(t, i+1, int(sn))

			mockedAccounts.assets.BaseTokens -= storageDeposit
			txb.BuildTransactionEssence(dummyStateMetadata)
			txb.MustBalanced()
		}
	}
	t.Run("create foundry ok", func(t *testing.T) {
		initTest()
		createNFoundries(3)
		essence, _ := txb.BuildTransactionEssence(dummyStateMetadata)
		txb.MustBalanced()
		essenceBytes, err := essence.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		t.Logf("essence bytes len = %d", len(essenceBytes))
	})
}

func TestSerDe(t *testing.T) {
	t.Run("serde BasicOutput", func(t *testing.T) {
		reqMetadata := isc.RequestMetadata{
			SenderContract: isc.EmptyContractIdentity(),
			TargetContract: 0,
			EntryPoint:     0,
			Params:         dict.New(),
			Allowance:      isc.NewEmptyAssets(),
			GasBudget:      0,
		}
		assets := isc.NewEmptyAssets()
		out := transaction.AdjustToMinimumStorageDeposit(transaction.MakeBasicOutput(
			&iotago.Ed25519Address{},
			&iotago.Ed25519Address{1, 2, 3},
			assets,
			&reqMetadata,
			isc.SendOptions{},
		))
		data, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		outBack := &iotago.BasicOutput{}
		_, err = outBack.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		condSet := out.Conditions.MustSet()
		condSetBack := outBack.Conditions.MustSet()
		require.True(t, condSet[iotago.UnlockConditionAddress].Equal(condSetBack[iotago.UnlockConditionAddress]))
		require.EqualValues(t, out.Deposit(), outBack.Amount)
		require.EqualValues(t, 0, len(outBack.NativeTokens))
		require.True(t, outBack.Features.Equal(out.Features))
	})
	t.Run("serde FoundryOutput", func(t *testing.T) {
		out := &iotago.FoundryOutput{
			Conditions: iotago.UnlockConditions{
				&iotago.ImmutableAliasUnlockCondition{Address: tpkg.RandAliasAddress()},
			},
			Amount:       1337,
			NativeTokens: nil,
			SerialNumber: 5,
			TokenScheme: &iotago.SimpleTokenScheme{
				MintedTokens:  big.NewInt(200),
				MeltedTokens:  big.NewInt(0),
				MaximumSupply: big.NewInt(2000),
			},
			Features: nil,
		}
		data, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		outBack := &iotago.FoundryOutput{}
		_, err = outBack.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
		require.NoError(t, err)
		require.True(t, identicalFoundries(out, outBack))
	})
}
