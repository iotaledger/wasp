package testcore

import (
	"math"
	"strings"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
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
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func GetStorageDeposit(tx *iotago.Transaction) []uint64 {
	ret := make([]uint64, len(tx.Essence.Outputs))
	for i, out := range tx.Essence.Outputs {
		ret[i] = parameters.L1.Protocol.RentStructure.MinRent(out)
	}
	return ret
}

func TestInitLoad(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	user, userAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(12))
	env.AssertL1Iotas(userAddr, utxodb.FundsFromFaucetAmount)
	ch, _, _ := env.NewChainExt(user, 10_000, "chain1")
	_ = ch.Log().Sync()

	dustCosts := transaction.NewStorageDepositEstimate()
	cassets := ch.L2CommonAccountAssets()
	require.EqualValues(t, 10_000-dustCosts.AnchorOutput, cassets.Iotas)
	require.EqualValues(t, 0, len(cassets.Tokens))

	t.Logf("common iotas: %d", ch.L2CommonAccountIotas())
	require.True(t, cassets.Iotas >= accounts.MinimumIotasOnCommonAccount)
}

// TestLedgerBaseConsistency deploys chain and check consistency of L1 and L2 ledgers
func TestLedgerBaseConsistency(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	genesisAddr := env.L1Ledger().GenesisAddress()
	assets := env.L1Assets(genesisAddr)
	require.EqualValues(t, env.L1Ledger().Supply(), assets.Iotas)

	// create chain
	ch, _, initTx := env.NewChainExt(nil, 0, "chain1")
	defer func() {
		_ = ch.Log().Sync()
	}()
	ch.AssertControlAddresses()
	t.Logf("originator address iotas: %d (spent %d)",
		env.L1Iotas(ch.OriginatorAddress), utxodb.FundsFromFaucetAmount-env.L1Iotas(ch.OriginatorAddress))

	// get all native tokens. Must be empty
	nativeTokenIDs := ch.GetOnChainTokenIDs()
	require.EqualValues(t, 0, len(nativeTokenIDs))

	// query dust parameters of the latest block
	totalIotasInfo := ch.GetTotalIotaInfo()
	totalIotasOnChain := ch.L2TotalIotas()
	// all goes to dust and to total iotas on chain
	totalSpent := totalIotasInfo.TotalDustDeposit + totalIotasInfo.TotalIotasInL2Accounts
	t.Logf("total on chain: dust deposit: %d, total iotas on chain: %d, total spent: %d",
		totalIotasInfo.TotalDustDeposit, totalIotasOnChain, totalSpent)
	// what has left on L1 address
	env.AssertL1Iotas(ch.OriginatorAddress, utxodb.FundsFromFaucetAmount-totalSpent)

	// let's analise dust deposit on origin and init transactions
	vByteCostInit := GetStorageDeposit(initTx)[0]
	dustCosts := transaction.NewStorageDepositEstimate()
	// what we spent is only for dust deposits for those 2 transactions
	require.EqualValues(t, int(totalSpent), int(dustCosts.AnchorOutput+vByteCostInit))

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
	require.EqualValues(t, 0, len(totalAssets.Tokens))
	// what spent all goes to the alias output
	require.EqualValues(t, int(totalSpent), int(aliasOut.Amount))
	// total iotas on L2 must be equal to alias output iotas - dust deposit
	ch.AssertL2TotalIotas(aliasOut.Amount - dustCosts.AnchorOutput)

	// all dust deposit of the init request goes to the user account
	ch.AssertL2Iotas(ch.OriginatorAgentID, vByteCostInit)
	// common account is empty
	require.EqualValues(t, 0, ch.L2CommonAccountIotas())
}

// TestNoTargetPostOnLedger test what happens when sending requests to non-existent contract or entry point
func TestNoTargetPostOnLedger(t *testing.T) {
	t.Run("no contract,originator==user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		defer func() {
			_ = ch.Log().Sync()
		}()

		totalIotasBefore := ch.L2TotalIotas()
		originatorsL2IotasBefore := ch.L2Iotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1Iotas(ch.OriginatorAddress)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams("dummyContract", "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, nil)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrContractNotFound.Create(iscp.Hn("dummyContract")).Error())

		totalIotasAfter := ch.L2TotalIotas()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()

		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// user on L1 is charged with dust deposit
		env.AssertL1Iotas(ch.OriginatorAddress, originatorsL1IotasBefore-reqDustDeposit)
		// originator (user) is charged with gas fee on L2
		ch.AssertL2Iotas(ch.OriginatorAgentID, originatorsL2IotasBefore+reqDustDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountIotasAfter)
	})
	t.Run("no contract,originator!=user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		defer func() {
			_ = ch.Log().Sync()
		}()

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := iscp.NewAgentID(senderAddr)

		totalIotasBefore := ch.L2TotalIotas()
		originatorsL2IotasBefore := ch.L2Iotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1Iotas(ch.OriginatorAddress)
		env.AssertL1Iotas(senderAddr, utxodb.FundsFromFaucetAmount)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams("dummyContract", "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrContractNotFound.Create(iscp.Hn("dummyContract")).Error())

		totalIotasAfter := ch.L2TotalIotas()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()

		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// originator on L1 does not change
		env.AssertL1Iotas(ch.OriginatorAddress, originatorsL1IotasBefore)
		// user on L1 is charged with dust deposit
		env.AssertL1Iotas(senderAddr, utxodb.FundsFromFaucetAmount-reqDustDeposit)
		// originator account does not change
		ch.AssertL2Iotas(ch.OriginatorAgentID, originatorsL2IotasBefore)
		// user is charged with gas fee on L2
		ch.AssertL2Iotas(senderAgentID, reqDustDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountIotasAfter)
	})
	t.Run("no EP,originator==user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		defer func() {
			_ = ch.Log().Sync()
		}()

		totalIotasBefore := ch.L2TotalIotas()
		originatorsL2IotasBefore := ch.L2Iotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1Iotas(ch.OriginatorAddress)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams(root.Contract.Name, "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, nil)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrTargetEntryPointNotFound.Error())

		totalIotasAfter := ch.L2TotalIotas()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()

		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// user on L1 is charged with dust deposit
		env.AssertL1Iotas(ch.OriginatorAddress, originatorsL1IotasBefore-reqDustDeposit)
		// originator (user) is charged with gas fee on L2
		ch.AssertL2Iotas(ch.OriginatorAgentID, originatorsL2IotasBefore+reqDustDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountIotasAfter)
	})
	t.Run("no EP,originator!=user", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		defer func() {
			_ = ch.Log().Sync()
		}()

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := iscp.NewAgentID(senderAddr)

		totalIotasBefore := ch.L2TotalIotas()
		originatorsL2IotasBefore := ch.L2Iotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1Iotas(ch.OriginatorAddress)
		env.AssertL1Iotas(senderAddr, utxodb.FundsFromFaucetAmount)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams(root.Contract.Name, "dummyEP").
			WithGasBudget(100_000)
		reqTx, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrTargetEntryPointNotFound.Error())

		totalIotasAfter := ch.L2TotalIotas()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := GetStorageDeposit(reqTx)[0]
		rec := ch.LastReceipt()
		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// originator on L1 does not change
		env.AssertL1Iotas(ch.OriginatorAddress, originatorsL1IotasBefore)
		// user on L1 is charged with dust deposit
		env.AssertL1Iotas(senderAddr, utxodb.FundsFromFaucetAmount-reqDustDeposit)
		// originator account does not change
		ch.AssertL2Iotas(ch.OriginatorAgentID, originatorsL2IotasBefore)
		// user is charged with gas fee on L2
		ch.AssertL2Iotas(senderAgentID, reqDustDeposit-rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t, int(rec.GasFeeCharged), commonAccountIotasAfter)
	})
}

func TestNoTargetView(t *testing.T) {
	t.Run("no contract view", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		chain := env.NewChain(nil, "chain1")
		chain.AssertControlAddresses()

		_, err := chain.CallView("dummyContract", "dummyEP")
		require.Error(t, err)
	})
	t.Run("no EP view", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		chain := env.NewChain(nil, "chain1")
		chain.AssertControlAddresses()

		_, err := chain.CallView(root.Contract.Name, "dummyEP")
		require.Error(t, err)
	})
}

func TestOkCall(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	ch := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name).
		WithGasBudget(100_000)
	_, err := ch.PostRequestSync(req, nil)
	require.NoError(t, err)
}

func TestEstimateGas(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true}).
		WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain(nil, "chain1")
	ch.MustDepositIotasToL2(10000, nil)
	err := ch.DeployContract(nil, sbtestsc.Contract.Name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	callParams := func() *solo.CallParams {
		return solo.NewCallParams(sbtestsc.Contract.Name, sbtestsc.FuncSetInt.Name,
			sbtestsc.ParamIntParamName, "v",
			sbtestsc.ParamIntParamValue, 42,
		)
	}

	getInt := func() int64 {
		v, err := ch.CallView(sbtestsc.Contract.Name, sbtestsc.FuncGetInt.Name,
			sbtestsc.ParamIntParamName, "v",
		)
		require.NoError(t, err)
		n, err := codec.DecodeInt64(v.MustGet("v"), 0)
		require.NoError(t, err)
		return n
	}

	var estimatedGas, estimatedGasFee uint64
	{
		keyPair, _ := env.NewKeyPairWithFunds()

		// we can call EstimateGas even with 0 iotas in L2 account
		estimatedGas, estimatedGasFee, err = ch.EstimateGasOffLedger(callParams(), keyPair, true)
		require.NoError(t, err)
		require.NotZero(t, estimatedGas)
		require.NotZero(t, estimatedGasFee)
		t.Logf("estimatedGas: %d, estimatedGasFee: %d", estimatedGas, estimatedGasFee)

		// test that EstimateGas did not actually commit changes in the state
		require.EqualValues(t, 0, getInt())
	}

	for _, testCase := range []struct {
		Desc          string
		L2Balance     uint64
		GasBudget     uint64
		ExpectedError string
	}{
		{
			Desc:          "0 iotas in L2 balance to cover gas fee",
			L2Balance:     0,
			GasBudget:     estimatedGas,
			ExpectedError: "gas budget exceeded",
		},
		{
			Desc:          "insufficient iotas in L2 balance to cover gas fee",
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
			agentID := iscp.NewAgentID(addr)

			if testCase.L2Balance > 0 {
				// deposit must come from another user so that we have exactly the funds we need on the test account (can't send lower than storage deposit)
				anotherKeyPair, _ := env.NewKeyPairWithFunds()
				req := solo.NewCallParams(
					accounts.Contract.Name,
					accounts.FuncTransferAllowanceTo.Name,
					dict.Dict{
						accounts.ParamAgentID:          codec.EncodeAgentID(iscp.NewAgentID(addr)),
						accounts.ParamForceOpenAccount: codec.EncodeBool(true),
					},
				).AddAllowance(iscp.NewAllowanceIotas(testCase.L2Balance)).
					AddIotas(10 * iscp.Mi).
					WithGasBudget(math.MaxUint64)

				_, err = ch.PostRequestSync(req, anotherKeyPair)
				require.NoError(t, err)
				balance := ch.L2Iotas(agentID)
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
				require.EqualValues(t, 42, getInt())
			}
		})
	}
}

func TestRepeatInit(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		err := ch.DepositIotasToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(root.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, root.ErrChainInitConditionsFailed)
		ch.CheckAccountLedger()
	})
	t.Run("accounts", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		err := ch.DepositIotasToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(accounts.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrRepeatingInitCall)
		ch.CheckAccountLedger()
	})
	t.Run("blocklog", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		err := ch.DepositIotasToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(blocklog.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrRepeatingInitCall)
		ch.CheckAccountLedger()
	})
	t.Run("blob", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		err := ch.DepositIotasToL2(10_000, nil)
		require.NoError(t, err)
		req := solo.NewCallParams(blob.Contract.Name, "init").
			WithGasBudget(100_000)
		_, err = ch.PostRequestSync(req, nil)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrRepeatingInitCall)
		ch.CheckAccountLedger()
	})
	t.Run("governance", func(t *testing.T) {
		env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
		ch := env.NewChain(nil, "chain1")
		err := ch.DepositIotasToL2(10_000, nil)
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
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true}).
		WithNativeContract(sbtestsc.Processor)

	ch := env.NewChain(nil, "chain1")

	senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
	// userAgentID := iscp.NewAgentID(userAddr, 0)

	err := ch.DepositIotasToL2(10_000, senderKeyPair)
	require.NoError(t, err)

	// get more iotas for originator
	originatorBalance := env.L1Assets(ch.OriginatorAddress).Iotas
	_, err = env.L1Ledger().GetFundsFromFaucet(ch.OriginatorAddress)
	require.NoError(t, err)
	env.AssertL1Iotas(ch.OriginatorAddress, originatorBalance+utxodb.FundsFromFaucetAmount)

	req := solo.NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name,
		root.ParamDeployer, iscp.NewAgentID(senderAddr)).
		AddIotas(100_000).
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
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "chain1")
	feePolicy := chain.GetGasFeePolicy()
	require.Nil(t, feePolicy.GasFeeTokenID)
	require.EqualValues(t, 0, feePolicy.ValidatorFeeShare)
}

func TestBurnLog(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	ch := env.NewChain(nil, "chain1")

	ch.MustDepositIotasToL2(30_000, nil)
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
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true, Debug: true, PrintStackTrace: true}).
		WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain(nil, "chain1")

	ch.MustDepositIotasToL2(10000, nil)

	err := ch.DeployContract(nil, sbtestsc.Contract.Name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	initialBlockIndex := ch.GetLatestBlockInfo().BlockIndex

	reqSize := 5_000 // bytes
	dust := 1 * iscp.Mi

	maxRequestsPerBlock := parameters.L1.MaxTransactionSize / reqSize

	reqs := make([]iscp.Request, maxRequestsPerBlock+1)
	for i := 0; i < len(reqs); i++ {
		req, err := solo.NewIscpRequestFromCallParams(
			ch,
			solo.NewCallParams(sbtestsc.Contract.Name, sbtestsc.FuncSendLargeRequest.Name,
				sbtestsc.ParamSize, uint32(reqSize),
			).
				AddIotas(dust).
				AddAllowanceIotas(dust).
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
		receipt, ok := ch.GetRequestReceipt(req.ID())
		require.True(t, ok)
		require.Nil(t, receipt.Error)
	}
}
