// excluded temporarily because of compilation errors

package testcore

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestInitLoad(t *testing.T) {
	env := solo.New(t)
	user, userAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(12))
	env.AssertL1BaseTokens(userAddr, iotaclient.FundsFromFaucetAmount)
	var originAmount coin.Value = 10 * isc.Million
	ch, _ := env.NewChainExt(user, originAmount, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)
	_ = ch.Log().Sync()

	cassets := ch.L2CommonAccountAssets()
	require.EqualValues(t,
		originAmount,
		cassets.BaseTokens())
	require.EqualValues(t, 1, len(cassets.Coins))

	t.Logf("common base tokens: %d", ch.L2CommonAccountBaseTokens())
	require.True(t, cassets.BaseTokens() >= governance.DefaultMinBaseTokensOnCommonAccount)

	testdbhash.VerifyDBHash(env, t.Name())
}

// TestLedgerBaseConsistency deploys chain and check consistency of L1 and L2 ledgers
func TestLedgerBaseConsistency(t *testing.T) {
	env := solo.New(t)
	ch, _ := env.NewChainExt(nil, 10*isc.Million, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

	ch.CheckChain()

	l2Total1 := ch.L2TotalAssets().BaseTokens()
	someUserWallet, _ := env.NewKeyPairWithFunds()
	ch.DepositBaseTokensToL2(1*isc.Million, someUserWallet)
	l2Total2 := ch.L2TotalAssets().BaseTokens()
	require.Equal(t, l2Total1+1*isc.Million, l2Total2)

	ch.CheckChain()
}

// TestNoTargetPostOnLedger test what happens when sending requests to non-existent contract or entry point
func TestNoTargetPostOnLedger(t *testing.T) {
	t.Run("no contract", func(t *testing.T) {
		env := solo.New(t)
		ch, _ := env.NewChainExt(nil, governance.DefaultMinBaseTokensOnCommonAccount, "chain", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := isc.NewAddressAgentID(senderAddr)

		l2TotalBaseTokensBefore := ch.L2TotalBaseTokens()
		senderL1BaseTokensBefore := env.L1BaseTokens(senderAddr)
		senderL2BaseTokensBefore := ch.L2BaseTokens(senderAgentID)
		originatorL2BaseTokensBefore := ch.L2BaseTokens(ch.OriginatorAgentID)
		commonAccountBaseTokensBefore := ch.L2CommonAccountBaseTokens()

		require.EqualValues(t, governance.DefaultMinBaseTokensOnCommonAccount, commonAccountBaseTokensBefore)

		_, l1Res, _, err := ch.PostRequestSyncTx(
			solo.NewCallParamsEx("dummyContract", "dummyEP").
				WithAssets(isc.NewAssets(1*isc.Million)).
				WithGasBudget(100_000),
			senderKeyPair,
		)

		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrContractNotFound.Create(uint32(isc.Hn("dummyContract"))).Error())

		l2TotalBaseTokensAfter := ch.L2TotalBaseTokens()
		senderL1BaseTokensAfter := env.L1BaseTokens(senderAddr)
		senderL2BaseTokensAfter := ch.L2BaseTokens(senderAgentID)
		originatorL2BaseTokensAfter := ch.L2BaseTokens(ch.OriginatorAgentID)
		commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()
		l1GasFee := coin.Value(l1Res.Effects.Data.GasFee())
		l2GasFee := ch.LastReceipt().GasFeeCharged

		// sender deposited 1mil to L2 and spent L1 gas fees (which goes to a black hole?)
		require.Equal(t, senderL1BaseTokensBefore-1*isc.Million-l1GasFee, senderL1BaseTokensAfter)
		// total L2 assets is increased by 1mil
		require.Equal(t, l2TotalBaseTokensBefore+1*isc.Million, l2TotalBaseTokensAfter)
		// sender got 1mil minus l2GasFee
		require.Equal(t, senderL2BaseTokensBefore+1*isc.Million-l2GasFee, senderL2BaseTokensAfter)
		// l2GasFee goes to payoutAgentID (originator by default)
		require.Equal(t, originatorL2BaseTokensBefore+l2GasFee, originatorL2BaseTokensAfter)
		// common account is left untouched
		require.Equal(t, commonAccountBaseTokensBefore, commonAccountBaseTokensAfter)
	})
	t.Run("no EP,originator==user", func(t *testing.T) {
		env := solo.New(t)
		ch, _ := env.NewChainExt(nil, 0, "chain", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

		totalBaseTokensBefore := ch.L2TotalBaseTokens()
		originatorsL2BaseTokensBefore := ch.L2BaseTokens(ch.OriginatorAgentID)
		originatorsL1BaseTokensBefore := env.L1BaseTokens(ch.OriginatorAddress)
		require.EqualValues(t, governance.DefaultMinBaseTokensOnCommonAccount, ch.L2CommonAccountBaseTokens())

		req := solo.NewCallParamsEx(root.Contract.Name, "dummyEP").
			WithGasBudget(100_000)
		_, _, _, err := ch.PostRequestSyncTx(req, nil)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrTargetEntryPointNotFound.Error())

		totalBaseTokensAfter := ch.L2TotalBaseTokens()
		commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()

		// total base tokens on chain increase by the storage deposit from the request tx
		require.EqualValues(t, int(totalBaseTokensBefore), int(totalBaseTokensAfter))
		// user on L1 is charged with storage deposit
		env.AssertL1BaseTokens(ch.OriginatorAddress, originatorsL1BaseTokensBefore)
		// originator (user) is charged with gas fee on L2
		ch.AssertL2BaseTokens(ch.OriginatorAgentID, originatorsL2BaseTokensBefore)
		// all gas fee goes to the common account
		require.EqualValues(t, governance.DefaultMinBaseTokensOnCommonAccount, commonAccountBaseTokensAfter)
	})
	t.Run("no EP,originator!=user", func(t *testing.T) {
		env := solo.New(t)
		ch, _ := env.NewChainExt(nil, 0, "chain", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := isc.NewAddressAgentID(senderAddr)

		totalBaseTokensBefore := ch.L2TotalBaseTokens()
		originatorsL2BaseTokensBefore := ch.L2BaseTokens(ch.OriginatorAgentID)
		originatorsL1BaseTokensBefore := env.L1BaseTokens(ch.OriginatorAddress)
		env.AssertL1BaseTokens(senderAddr, iotaclient.FundsFromFaucetAmount)
		require.EqualValues(t, governance.DefaultMinBaseTokensOnCommonAccount, ch.L2CommonAccountBaseTokens())

		req := solo.NewCallParamsEx(root.Contract.Name, "dummyEP").
			WithGasBudget(100_000)
		_, _, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		// expecting specific error
		require.Contains(t, err.Error(), vm.ErrTargetEntryPointNotFound.Error())

		totalBaseTokensAfter := ch.L2TotalBaseTokens()
		commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()

		rec := ch.LastReceipt()
		// total base tokens on chain increase by the storage deposit from the request tx
		require.EqualValues(t, int(totalBaseTokensBefore), int(totalBaseTokensAfter))
		// originator on L1 does not change
		env.AssertL1BaseTokens(ch.OriginatorAddress, originatorsL1BaseTokensBefore)
		// user on L1 is charged with storage deposit
		env.AssertL1BaseTokens(senderAddr, iotaclient.FundsFromFaucetAmount)
		// originator account does not change
		ch.AssertL2BaseTokens(ch.OriginatorAgentID, originatorsL2BaseTokensBefore+rec.GasFeeCharged)
		// user is charged with gas fee on L2
		ch.AssertL2BaseTokens(senderAgentID, -rec.GasFeeCharged)
		// all gas fee goes to the common account
		require.EqualValues(t,
			governance.DefaultMinBaseTokensOnCommonAccount,
			commonAccountBaseTokensAfter,
		)
	})
}

func TestNoTargetView(t *testing.T) {
	t.Run("no contract view", func(t *testing.T) {
		env := solo.New(t)
		chain := env.NewChain()
		_, err := chain.CallViewEx("dummyContract", "dummyEP")
		require.Error(t, err)
	})
	t.Run("no EP view", func(t *testing.T) {
		env := solo.New(t)
		chain := env.NewChain()
		_, err := chain.CallViewEx(root.Contract.Name, "dummyEP")
		require.Error(t, err)
	})
}

func TestEstimateGas(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()
	ch.MustDepositBaseTokensToL2(10000, nil)

	callParams := func() *solo.CallParams {
		return solo.NewCallParams(sbtestsc.FuncCalcFibonacciIndirectStoreValue.Message(10))
	}

	getResult := func() uint64 {
		r, err := sbtestsc.FuncViewCalcFibonacciResult.Call(ch.CallView)
		require.NoError(t, err)
		return r
	}

	var estimatedGas uint64
	var estimatedGasFee coin.Value
	{
		keyPair, _ := env.NewKeyPairWithFunds()

		req := callParams().WithFungibleTokens(isc.NewAssets(1 * isc.Million).Coins).WithMaxAffordableGasBudget()
		_, estimate, err2 := ch.EstimateGasOnLedger(req, keyPair)
		estimatedGas = estimate.GasBurned
		estimatedGasFee = estimate.GasFeeCharged
		require.NoError(t, err2)
		require.NotZero(t, estimatedGasFee)
		require.NotZero(t, estimatedGasFee)
		t.Logf("estimatedGas: %d, estimatedGasFee: %d", estimatedGas, estimatedGasFee)

		// test that EstimateGas did not actually commit changes in the state
		require.EqualValues(t, 0, getResult())
	}

	for _, testCase := range []struct {
		Desc          string
		L2Balance     coin.Value
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
			agentID := isc.NewAddressAgentID(addr)

			if testCase.L2Balance > 0 {
				// deposit must come from another user so that we have exactly the funds we need on the test account (can't send lower than storage deposit)
				anotherKeyPair, _ := env.NewKeyPairWithFunds()
				err := ch.TransferAllowanceTo(
					isc.NewAssets(testCase.L2Balance),
					isc.NewAddressAgentID(addr),
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
			fmt.Println(rec)
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

func TestFeeBasic(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()
	feePolicy := chain.GetGasFeePolicy()
	require.EqualValues(t, 0, feePolicy.ValidatorFeeShare)
}

func TestBurnLog(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(30_000, nil)
	rec := ch.LastReceipt()
	t.Logf("receipt 1:\n%s", rec)
	t.Logf("burn log 1:\n%s", rec.GasBurnLog)

	rec = ch.LastReceipt()
	t.Logf("receipt 2:\n%s", rec)
	t.Logf("burn log 2:\n%s", rec.GasBurnLog)
}

func TestMessageSize(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(10000, nil)

	initialBlockIndex := ch.GetLatestBlockInfo().BlockIndex

	reqSize := 5_000 // bytes
	var storageDeposit coin.Value = 1 * isc.Million

	maxRequestsPerBlock := parameters.L1().MaxPayloadSize / reqSize

	reqs := make([]isc.Request, maxRequestsPerBlock+1)
	for i := 0; i < len(reqs); i++ {
		req, err := solo.ISCRequestFromCallParams(
			ch,
			solo.NewCallParams(sbtestsc.FuncSendLargeRequest.Message(uint64(reqSize)), sbtestsc.Contract.Name).
				AddBaseTokens(storageDeposit).
				AddAllowanceBaseTokens(storageDeposit).
				WithMaxAffordableGasBudget(),
			nil,
		)
		require.NoError(t, err)
		reqs[i] = req
	}

	env.AddRequestsToMempool(ch, reqs)
	ch.WaitUntilMempoolIsEmpty()

	// request outputs are so large that they have to be processed in two separate blocks
	require.Equal(t, initialBlockIndex+2, ch.GetLatestBlockInfo().BlockIndex)

	for _, req := range reqs {
		receipt, _ := ch.GetRequestReceipt(req.ID())
		require.Nil(t, receipt.Error)
	}
}

func TestInvalidSignatureRequestsAreNotProcessed(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()
	req := isc.NewOffLedgerRequest(ch.ID(), isc.NewMessage(isc.Hn("contract"), isc.Hn("entrypoint"), nil), 0, math.MaxUint64)
	badReqBytes := req.(*isc.OffLedgerRequestData).EssenceBytes()
	// append 33 bytes to the req essence to simulate a bad signature (32 bytes for the pubkey + 1 for 0 length signature)
	for i := 0; i < 33; i++ {
		badReqBytes = append(badReqBytes, 0x00)
	}
	badReq, err := isc.RequestFromBytes(badReqBytes)
	require.NoError(t, err)
	env.AddRequestsToMempool(ch, []isc.Request{badReq})
	time.Sleep(200 * time.Millisecond)
	// request won't be processed
	_, ok := ch.GetRequestReceipt(badReq.ID())
	require.False(t, ok)
}

func TestBatchWithSkippedRequestsReceipts(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()
	user, _ := env.NewKeyPairWithFunds()
	err := ch.DepositAssetsToL2(isc.NewAssets(10*isc.Million), user)
	require.NoError(t, err)

	// create a request with an invalid nonce that must be skipped
	skipReq := isc.NewOffLedgerRequest(ch.ID(), isc.NewMessage(isc.Hn("contract"), isc.Hn("entrypoint"), nil), 0, math.MaxUint64).WithNonce(9999).Sign(user)
	validReq := isc.NewOffLedgerRequest(ch.ID(), isc.NewMessage(isc.Hn("contract"), isc.Hn("entrypoint"), nil), 0, math.MaxUint64).WithNonce(0).Sign(user)

	ch.RunRequestsSync([]isc.Request{skipReq, validReq}, "")

	// block has been created with only 1 request, calling 	`GetRequestReceiptsForBlock` must yield 1 receipt as expected
	bi := ch.GetLatestBlockInfo()
	require.EqualValues(t, 1, bi.TotalRequests)
	receipts := ch.GetRequestReceiptsForBlock(bi.BlockIndex)
	require.Len(t, receipts, 1)
}
