package testcore

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func GetStorageDeposit(tx *iotago.Transaction) []uint64 {
	ret := make([]uint64, len(tx.Essence.Outputs))
	for i, out := range tx.Essence.Outputs {
		ret[i] = parameters.L1().Protocol.RentStructure.MinRent(out)
	}
	return ret
}

func TestInitLoad(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	user, userAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(12))
	env.AssertL1BaseTokens(userAddr, utxodb.FundsFromFaucetAmount)
	ch, _, _ := env.NewChainExt(user, 10_000, "chain1")
	_ = ch.Log().Sync()

	storageDepositCosts := transaction.NewStorageDepositEstimate()
	cassets := ch.L2CommonAccountAssets()
	require.EqualValues(t, 10_000-storageDepositCosts.AnchorOutput, cassets.BaseTokens)
	require.EqualValues(t, 0, len(cassets.NativeTokens))

	t.Logf("common base tokens: %d", ch.L2CommonAccountBaseTokens())
	require.True(t, cassets.BaseTokens >= accounts.MinimumBaseTokensOnCommonAccount)
}

// TestLedgerBaseConsistency deploys chain and check consistency of L1 and L2 ledgers
func TestLedgerBaseConsistency(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	genesisAddr := env.L1Ledger().GenesisAddress()
	assets := env.L1Assets(genesisAddr)
	require.EqualValues(t, env.L1Ledger().Supply(), assets.BaseTokens)

	// create chain
	ch, _, initTx := env.NewChainExt(nil, 0, "chain1")
	defer func() {
		_ = ch.Log().Sync()
	}()
	ch.AssertControlAddresses()
	t.Logf("originator address base tokens: %d (spent %d)",
		env.L1BaseTokens(ch.OriginatorAddress), utxodb.FundsFromFaucetAmount-env.L1BaseTokens(ch.OriginatorAddress))

	// get all native tokens. Must be empty
	nativeTokenIDs := ch.GetOnChainTokenIDs()
	require.EqualValues(t, 0, len(nativeTokenIDs))

	// query storage deposit parameters of the latest block
	totalBaseTokensInfo := ch.GetTotalBaseTokensInfo()
	totalBaseTokensOnChain := ch.L2TotalBaseTokens()
	// all goes to storage deposit and to total base tokens on chain
	totalSpent := totalBaseTokensInfo.TotalStorageDeposit + totalBaseTokensInfo.TotalBaseTokensInL2Accounts
	t.Logf("total on chain: storage deposit: %d, total base tokens on chain: %d, total spent: %d",
		totalBaseTokensInfo.TotalStorageDeposit, totalBaseTokensOnChain, totalSpent)
	// what has left on L1 address
	env.AssertL1BaseTokens(ch.OriginatorAddress, utxodb.FundsFromFaucetAmount-totalSpent)

	// let's analise storage deposit on origin and init transactions
	vByteCostInit := GetStorageDeposit(initTx)[0]
	storageDepositCosts := transaction.NewStorageDepositEstimate()
	// what we spent is only for storage deposits for those 2 transactions
	require.EqualValues(t, int(totalSpent), int(storageDepositCosts.AnchorOutput+vByteCostInit))

	// check if there's a single alias output on chain's address
	aliasOutputs := env.L1Ledger().GetAliasOutputs(ch.ChainID.AsAddress())
	require.EqualValues(t, 1, len(aliasOutputs))
	var aliasOut *iotago.AliasOutput
	for _, out := range aliasOutputs {
		aliasOut = out
	}

	// check total on chain assets
	totalAssets := ch.L2TotalAssets()
	// no native tokens expected
	require.EqualValues(t, 0, len(totalAssets.NativeTokens))
	// what spent all goes to the alias output
	require.EqualValues(t, int(totalSpent), int(aliasOut.Amount))
	// total base tokens on L2 must be equal to alias output base tokens - storage deposit
	ch.AssertL2TotalBaseTokens(aliasOut.Amount - storageDepositCosts.AnchorOutput)

	// all storage deposit of the init request goes to the user account
	ch.AssertL2BaseTokens(ch.OriginatorAgentID, vByteCostInit)
	// common account is empty
	require.EqualValues(t, 0, ch.L2CommonAccountBaseTokens())
}

// TestNoTargetPostOnLedger test what happens when sending requests to non-existent contract or entry point
func TestNoTargetPostOnLedger(t *testing.T) {
	t.Run("no contract,originator==user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		defer func() {
			_ = ch.Log().Sync()
		}()

		totalBaseTokensBefore := ch.L2TotalBaseTokens()
		originatorsL2BaseTokensBefore := ch.L2BaseTokens(ch.OriginatorAgentID)
		originatorsL1BaseTokensBefore := env.L1BaseTokens(ch.OriginatorAddress)
		require.EqualValues(t, 0, ch.L2CommonAccountBaseTokens())

		req := solo.NewCallParams("dummyContract", "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, nil)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrContractNotFound.Create(isc.Hn("dummyContract")).Error())

		totalBaseTokensAfter := ch.L2TotalBaseTokens()
		commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()

		reqStorageDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()

		// total base tokens on chain increase by the storage deposit from the request tx
		require.EqualValues(t, int(totalBaseTokensBefore+reqStorageDeposit), int(totalBaseTokensAfter))
		// user on L1 is charged with storage deposit
		env.AssertL1BaseTokens(ch.OriginatorAddress, originatorsL1BaseTokensBefore-reqStorageDeposit)
		// originator (user) is charged with gas fee on L2
		ch.AssertL2BaseTokens(ch.OriginatorAgentID, originatorsL2BaseTokensBefore+reqStorageDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountBaseTokensAfter)
	})
	t.Run("no contract,originator!=user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		defer func() {
			_ = ch.Log().Sync()
		}()

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := isc.NewAgentID(senderAddr)

		totalBaseTokensBefore := ch.L2TotalBaseTokens()
		originatorsL2BaseTokensBefore := ch.L2BaseTokens(ch.OriginatorAgentID)
		originatorsL1BaseTokensBefore := env.L1BaseTokens(ch.OriginatorAddress)
		env.AssertL1BaseTokens(senderAddr, utxodb.FundsFromFaucetAmount)
		require.EqualValues(t, 0, ch.L2CommonAccountBaseTokens())

		req := solo.NewCallParams("dummyContract", "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrContractNotFound.Create(isc.Hn("dummyContract")).Error())

		totalBaseTokensAfter := ch.L2TotalBaseTokens()
		commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()

		reqStorageDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()

		// total base tokens on chain increase by the storage deposit from the request tx
		require.EqualValues(t, int(totalBaseTokensBefore+reqStorageDeposit), int(totalBaseTokensAfter))
		// originator on L1 does not change
		env.AssertL1BaseTokens(ch.OriginatorAddress, originatorsL1BaseTokensBefore)
		// user on L1 is charged with storage deposit
		env.AssertL1BaseTokens(senderAddr, utxodb.FundsFromFaucetAmount-reqStorageDeposit)
		// originator account does not change
		ch.AssertL2BaseTokens(ch.OriginatorAgentID, originatorsL2BaseTokensBefore)
		// user is charged with gas fee on L2
		ch.AssertL2BaseTokens(senderAgentID, reqStorageDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountBaseTokensAfter)
	})
	t.Run("no EP,originator==user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		defer func() {
			_ = ch.Log().Sync()
		}()

		totalBaseTokensBefore := ch.L2TotalBaseTokens()
		originatorsL2BaseTokensBefore := ch.L2BaseTokens(ch.OriginatorAgentID)
		originatorsL1BaseTokensBefore := env.L1BaseTokens(ch.OriginatorAddress)
		require.EqualValues(t, 0, ch.L2CommonAccountBaseTokens())

		req := solo.NewCallParams(root.Contract.Name, "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, nil)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrTargetEntryPointNotFound.Error())

		totalBaseTokensAfter := ch.L2TotalBaseTokens()
		commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()

		reqStorageDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()

		// total base tokens on chain increase by the storage deposit from the request tx
		require.EqualValues(t, int(totalBaseTokensBefore+reqStorageDeposit), int(totalBaseTokensAfter))
		// user on L1 is charged with storage deposit
		env.AssertL1BaseTokens(ch.OriginatorAddress, originatorsL1BaseTokensBefore-reqStorageDeposit)
		// originator (user) is charged with gas fee on L2
		ch.AssertL2BaseTokens(ch.OriginatorAgentID, originatorsL2BaseTokensBefore+reqStorageDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountBaseTokensAfter)
	})
	t.Run("no EP,originator!=user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		defer func() {
			_ = ch.Log().Sync()
		}()

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := isc.NewAgentID(senderAddr)

		totalBaseTokensBefore := ch.L2TotalBaseTokens()
		originatorsL2BaseTokensBefore := ch.L2BaseTokens(ch.OriginatorAgentID)
		originatorsL1BaseTokensBefore := env.L1BaseTokens(ch.OriginatorAddress)
		env.AssertL1BaseTokens(senderAddr, utxodb.FundsFromFaucetAmount)
		require.EqualValues(t, 0, ch.L2CommonAccountBaseTokens())

		req := solo.NewCallParams(root.Contract.Name, "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrTargetEntryPointNotFound.Error())

		totalBaseTokensAfter := ch.L2TotalBaseTokens()
		commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()

		reqStorageDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()
		// total base tokens on chain increase by the storage deposit from the request tx
		require.EqualValues(t, int(totalBaseTokensBefore+reqStorageDeposit), int(totalBaseTokensAfter))
		// originator on L1 does not change
		env.AssertL1BaseTokens(ch.OriginatorAddress, originatorsL1BaseTokensBefore)
		// user on L1 is charged with storage deposit
		env.AssertL1BaseTokens(senderAddr, utxodb.FundsFromFaucetAmount-reqStorageDeposit)
		// originator account does not change
		ch.AssertL2BaseTokens(ch.OriginatorAgentID, originatorsL2BaseTokensBefore)
		// user is charged with gas fee on L2
		ch.AssertL2BaseTokens(senderAgentID, reqStorageDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountBaseTokensAfter)
	})
}

func TestNoTargetView(t *testing.T) {
	t.Run("no contract view", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		chain := env.NewChain()
		chain.AssertControlAddresses()

		_, err := chain.CallView("dummyContract", "dummyEP")
		require.Error(t, err)
	})
	t.Run("no EP view", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		chain := env.NewChain()
		chain.AssertControlAddresses()

		_, err := chain.CallView(root.Contract.Name, "dummyEP")
		require.Error(t, err)
	})
}

func TestOkCall(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name).
		WithGasBudget(100_000)
	_, err := ch.PostRequestSync(req, nil)
	require.NoError(t, err)
}

func TestEstimateGas(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain()
	ch.MustDepositBaseTokensToL2(10000, nil)
	err := ch.DeployContract(nil, sbtestsc.Contract.Name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	callParams := func() *solo.CallParams {
		return solo.NewCallParams(sbtestsc.Contract.Name, sbtestsc.FuncCalcFibonacciIndirectStoreValue.Name,
			sbtestsc.ParamN, uint64(10),
		)
	}

	getResult := func() int64 {
		res, err := ch.CallView(sbtestsc.Contract.Name, sbtestsc.FuncViewCalcFibonacciResult.Name)
		require.NoError(t, err)
		n, err := codec.DecodeInt64(res.MustGet(sbtestsc.ParamN), 0)
		require.NoError(t, err)
		return n
	}

	var estimatedGas, estimatedGasFee uint64
	{
		keyPair, _ := env.NewKeyPairWithFunds()

		// we can call EstimateGas even with 0 base tokens in L2 account
		estimatedGas, estimatedGasFee, err = ch.EstimateGasOffLedger(callParams(), keyPair, true)
		require.NoError(t, err)
		require.NotZero(t, estimatedGas)
		require.NotZero(t, estimatedGasFee)
		t.Logf("estimatedGas: %d, estimatedGasFee: %d", estimatedGas, estimatedGasFee)

		// test that EstimateGas did not actually commit changes in the state
		require.EqualValues(t, 0, getResult())
	}

	for _, testCase := range []struct {
		Desc          string
		L2Balance     uint64
		GasBudget     uint64
		ExpectedError string
	}{
		{
			Desc:          "0 base tokens in L2 balance to cover gas fee",
			L2Balance:     0,
			GasBudget:     estimatedGas,
			ExpectedError: "gas budget exceeded",
		},
		{
			Desc:          "insufficient base tokens in L2 balance to cover gas fee",
			L2Balance:     estimatedGasFee - 1,
			GasBudget:     estimatedGas,
			ExpectedError: "gas budget exceeded",
		},
		{
			Desc:          "insufficient gas budget",
			L2Balance:     estimatedGasFee,
			GasBudget:     estimatedGas - 1,
			ExpectedError: "gas budget exceeded",
		},
		{
			Desc:      "success",
			L2Balance: estimatedGasFee,
			GasBudget: estimatedGas,
		},
	} {
		t.Run(testCase.Desc, func(t *testing.T) {
			keyPair, addr := env.NewKeyPairWithFunds()
			agentID := isc.NewAgentID(addr)

			if testCase.L2Balance > 0 {
				// deposit must come from another user so that we have exactly the funds we need on the test account (can't send lower than storage deposit)
				anotherKeyPair, _ := env.NewKeyPairWithFunds()
				err = ch.TransferAllowanceTo(
					isc.NewAssetsBaseTokens(testCase.L2Balance),
					isc.NewAgentID(addr),
					anotherKeyPair,
				)
				require.NoError(t, err)
				balance := ch.L2BaseTokens(agentID)
				require.Equal(t, testCase.L2Balance, balance)
			}

			_, err := ch.PostRequestOffLedger(
				callParams().WithGasBudget(testCase.GasBudget),
				keyPair,
			)
			rec := ch.LastReceipt()
			println(rec)
			if testCase.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError)
			} else {
				require.NoError(t, err)
				// changes committed to the state
				require.NotZero(t, getResult())
			}
		})
	}
}

func TestRepeatInit(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		err := ch.DepositBaseTokensToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(root.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, root.ErrChainInitConditionsFailed)
		ch.CheckAccountLedger()
	})
	t.Run("accounts", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		err := ch.DepositBaseTokensToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(accounts.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrRepeatingInitCall)
		ch.CheckAccountLedger()
	})
	t.Run("blocklog", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		err := ch.DepositBaseTokensToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(blocklog.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrRepeatingInitCall)
		ch.CheckAccountLedger()
	})
	t.Run("blob", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		err := ch.DepositBaseTokensToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(blob.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrRepeatingInitCall)
		ch.CheckAccountLedger()
	})
	t.Run("governance", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch := env.NewChain()
		err := ch.DepositBaseTokensToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(governance.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrRepeatingInitCall)
		ch.CheckAccountLedger()
	})
}

func TestDeployNativeContract(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(sbtestsc.Processor)

	ch := env.NewChain()

	senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
	// userAgentID := isc.NewAgentID(userAddr, 0)

	err := ch.DepositBaseTokensToL2(10_000, senderKeyPair)
	require.NoError(t, err)

	// get more base tokens for originator
	originatorBalance := env.L1Assets(ch.OriginatorAddress).BaseTokens
	_, err = env.L1Ledger().GetFundsFromFaucet(ch.OriginatorAddress)
	require.NoError(t, err)
	env.AssertL1BaseTokens(ch.OriginatorAddress, originatorBalance+utxodb.FundsFromFaucetAmount)

	req := solo.NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name,
		root.ParamDeployer, isc.NewAgentID(senderAddr)).
		AddBaseTokens(100_000).
		WithGasBudget(100_000)
	_, err = ch.PostRequestSync(req, nil)
	require.NoError(t, err)

	err = ch.DeployContract(senderKeyPair, "sctest", sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)
	//
	//req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name)
	//_, err := ch.PostRequestSync(req, nil)
}

func TestFeeBasic(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	chain := env.NewChain()
	feePolicy := chain.GetGasFeePolicy()
	require.True(t, isc.IsEmptyNativeTokenID(feePolicy.GasFeeTokenID))
	require.EqualValues(t, 0, feePolicy.ValidatorFeeShare)
}

func TestBurnLog(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(30_000, nil)
	rec := ch.LastReceipt()
	t.Logf("receipt 1:\n%s", rec)
	t.Logf("burn log 1:\n%s", rec.GasBurnLog)

	_, err := ch.UploadBlob(nil, "field", strings.Repeat("dummy data", 1000))
	require.NoError(t, err)

	rec = ch.LastReceipt()
	t.Logf("receipt 2:\n%s", rec)
	t.Logf("burn log 2:\n%s", rec.GasBurnLog)
}

func TestMessageSize(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		Debug:                    true,
		PrintStackTrace:          true,
	}).
		WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(10000, nil)

	err := ch.DeployContract(nil, sbtestsc.Contract.Name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	initialBlockIndex := ch.GetLatestBlockInfo().BlockIndex

	reqSize := 5_000 // bytes
	storageDeposit := 1 * isc.Million

	maxRequestsPerBlock := parameters.L1().MaxPayloadSize / reqSize

	reqs := make([]isc.Request, maxRequestsPerBlock+1)
	for i := 0; i < len(reqs); i++ {
		req, err := solo.NewIscRequestFromCallParams(
			ch,
			solo.NewCallParams(sbtestsc.Contract.Name, sbtestsc.FuncSendLargeRequest.Name,
				sbtestsc.ParamSize, uint32(reqSize),
			).
				AddBaseTokens(storageDeposit).
				AddAllowanceBaseTokens(storageDeposit).
				WithMaxAffordableGasBudget(),
			nil,
		)
		require.NoError(t, err)
		reqs[i] = req
	}

	env.AddRequestsToChainMempoolWaitUntilInbufferEmpty(ch, reqs)
	ch.WaitUntilMempoolIsEmpty()

	// request outputs are so large that they have to be processed in two separate blocks
	require.Equal(t, initialBlockIndex+2, ch.GetLatestBlockInfo().BlockIndex)

	for _, req := range reqs {
		receipt, err := ch.GetRequestReceipt(req.ID())
		require.Nil(t, err)
		require.Nil(t, receipt.Error)
	}
}
