// excluded temporarily because of compilation errors

package testcore

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/v2/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestInitLoad(t *testing.T) {
	env := solo.New(t)
	user, userAddr := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	env.AssertL1BaseTokens(userAddr, iotaclient.FundsFromFaucetAmount)
	var originAmount coin.Value = 10 * isc.Million
	ch, _ := env.NewChainExt(user, originAmount, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

	cassets := ch.L2CommonAccountAssets()
	require.EqualValues(t,
		originAmount,
		cassets.BaseTokens())
	require.EqualValues(t, 1, cassets.Coins.Size())

	testdbhash.VerifyDBHash(env, t.Name())
}

// TestLedgerBaseConsistency deploys chain and check consistency of L1 and L2 ledgers
func TestLedgerBaseConsistency(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
	ch, _ := env.NewChainExt(nil, 0, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

	ch.CheckChain()

	l2Total1 := ch.L2TotalAssets().BaseTokens()
	someUserWallet, _ := env.NewKeyPairWithFunds()
	ch.DepositBaseTokensToL2(1*isc.Million, someUserWallet)
	l2Total2 := ch.L2TotalAssets().BaseTokens()
	require.Equal(t, l2Total1+1*isc.Million, l2Total2)

	ch.CheckChain()
}

// TestLedgerBaseConsistencyWithRequiredTopUpFee deploys a chain and checks the consistency of L1 and L2 ledgers after topping up fees
func TestLedgerBaseConsistencyWithRequiredTopUpFee(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})

	const initialCommonAccountBalance = isc.GasCoinTargetValue / 2

	ch, _ := env.NewChainExt(nil, initialCommonAccountBalance, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)
	ch.CheckChain()

	// common account has initialCommonAccountBalance
	require.EqualValues(t,
		initialCommonAccountBalance,
		ch.L2BaseTokens(accounts.CommonAccount()),
	)
	// chain admin's account is empty
	require.Zero(t, ch.L2BaseTokens(ch.AdminAgentID()))

	// total owned by the chain is initialCommonAccountBalance
	require.EqualValues(t,
		coin.Value(initialCommonAccountBalance),
		ch.L2TotalBaseTokens(),
	)

	gasCoinValueBefore := ch.GetLatestGasCoin().Value

	const depositedByUser = isc.GasCoinTargetValue * 4
	someUserWallet, someUserAddr := env.NewKeyPairWithFunds()
	// user deposits some tokens
	_, _, vmRes, ptbRes, err := ch.PostRequestSyncTx(
		solo.NewCallParams(accounts.FuncDeposit.Message()).
			AddBaseTokens(depositedByUser).
			WithGasBudget(math.MaxUint64),
		someUserWallet,
	)
	t.Logf("PTB gas fee: %d", ptbRes.Effects.Data.GasFee())
	require.NoError(t, err)
	ch.CheckChain()

	gasCoinValueAfter := ch.GetLatestGasCoin().Value
	t.Logf("gasCoinValueBefore: %d, gasCoinValueAfter: %d", gasCoinValueBefore, gasCoinValueAfter)

	// the common account is topped up to isc.GasCoinTargetValue,
	// taking from the L2 gas fee
	addedToCommonAccount := min(
		isc.GasCoinTargetValue-initialCommonAccountBalance,
		vmRes.Receipt.GasFeeCharged,
	)
	// the gas coin is topped up to GasCoinTargetValue, taking from the
	// common account
	deductedForGasCoin := min(
		initialCommonAccountBalance+addedToCommonAccount,
		isc.GasCoinTargetValue-gasCoinValueBefore,
	)
	require.EqualValues(t,
		initialCommonAccountBalance+addedToCommonAccount-deductedForGasCoin,
		ch.L2BaseTokens(accounts.CommonAccount()),
	)

	// the total deposited tokens in the chain
	totalDeposited := coin.Value(initialCommonAccountBalance + depositedByUser)
	// the tokens that remain in L2
	totalL2Tokens := totalDeposited - deductedForGasCoin
	require.EqualValues(t, totalL2Tokens, ch.L2TotalBaseTokens())

	// the user pays for the L2 gas fee
	require.EqualValues(t,
		depositedByUser-vmRes.Receipt.GasFeeCharged,
		ch.L2BaseTokens(isc.NewAddressAgentID(someUserAddr)),
	)

	// the gas coin is topped up to GasCoinTargetValue, and then it is used
	// to pay for L1 gas fee
	require.EqualValues(t,
		gasCoinValueBefore+deductedForGasCoin-coin.Value(ptbRes.Effects.Data.GasFee()),
		gasCoinValueAfter,
	)

	// the collected fees go to the payout (chain admin by default), minus the
	// amount used to top up the common account
	require.EqualValues(t,
		vmRes.Receipt.GasFeeCharged-addedToCommonAccount,
		ch.L2BaseTokens(ch.AdminAgentID()),
	)
}

// TestNoTargetPostOnLedger test what happens when sending requests to non-existent contract or entry point
func TestNoTargetPostOnLedger(t *testing.T) {
	for _, test := range []struct {
		Name               string
		Req                *solo.CallParams
		SenderIsOriginator bool
		ExpectedError      string
	}{
		{
			Name:               "no contract, sender == originator",
			Req:                solo.NewCallParamsEx("dummyContract", "dummyEP"),
			SenderIsOriginator: true,
			ExpectedError:      vm.ErrContractNotFound.Create(uint32(isc.Hn("dummyContract"))).Error(),
		},
		{
			Name:               "no contract, sender != originator",
			Req:                solo.NewCallParamsEx("dummyContract", "dummyEP"),
			SenderIsOriginator: false,
			ExpectedError:      vm.ErrContractNotFound.Create(uint32(isc.Hn("dummyContract"))).Error(),
		},
		{
			Name:               "no EP, sender == originator",
			Req:                solo.NewCallParamsEx(root.Contract.Name, "dummyEP"),
			SenderIsOriginator: true,
			ExpectedError:      vm.ErrTargetEntryPointNotFound.Error(),
		},
		{
			Name:               "no EP, sender != originator",
			Req:                solo.NewCallParamsEx(root.Contract.Name, "dummyEP"),
			SenderIsOriginator: false,
			ExpectedError:      vm.ErrTargetEntryPointNotFound.Error(),
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			env := solo.New(t, &solo.InitOptions{
				Debug:           true,
				PrintStackTrace: true,
			})
			ch, _ := env.NewChainExt(nil, 0, "chain", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

			senderKeyPair, senderAddr := ch.ChainAdmin, ch.AdminAddress()
			if !test.SenderIsOriginator {
				senderKeyPair, senderAddr = env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
			}
			senderAgentID := isc.NewAddressAgentID(senderAddr)

			gasCoinL1Before := ch.GetLatestGasCoin().Value
			l2TotalBaseTokensBefore := ch.L2TotalBaseTokens()
			senderL1BaseTokensBefore := env.L1BaseTokens(senderAddr)
			senderL2BaseTokensBefore := ch.L2BaseTokens(senderAgentID)
			originatorL2BaseTokensBefore := ch.L2BaseTokens(ch.AdminAgentID())
			commonAccountBaseTokensBefore := ch.L2CommonAccountBaseTokens()

			require.EqualValues(t, 0, commonAccountBaseTokensBefore)

			_, l1Res, _, _, err := ch.PostRequestSyncTx(
				test.Req.
					WithAssets(isc.NewAssets(1*isc.Million)).
					WithGasBudget(100_000),
				senderKeyPair,
			)

			require.Contains(t, err.Error(), test.ExpectedError)

			gasCoinL1After := ch.GetLatestGasCoin().Value
			t.Logf("gasCoinL1Before: %d, gasCoinL1After: %d", gasCoinL1Before, gasCoinL1After)
			l2TotalBaseTokensAfter := ch.L2TotalBaseTokens()
			t.Logf("l2TotalBaseTokensBefore: %d, l2TotalBaseTokensAfter: %d", l2TotalBaseTokensBefore, l2TotalBaseTokensAfter)
			senderL1BaseTokensAfter := env.L1BaseTokens(senderAddr)
			t.Logf("senderL1BaseTokensBefore: %d, senderL1BaseTokensAfter: %d", senderL1BaseTokensBefore, senderL1BaseTokensAfter)
			senderL2BaseTokensAfter := ch.L2BaseTokens(senderAgentID)
			t.Logf("senderL2BaseTokensBefore: %d, senderL2BaseTokensAfter: %d", senderL2BaseTokensBefore, senderL2BaseTokensAfter)
			commonAccountBaseTokensAfter := ch.L2CommonAccountBaseTokens()
			t.Logf("commonAccountBaseTokensBefore: %d, commonAccountBaseTokensAfter: %d", commonAccountBaseTokensBefore, commonAccountBaseTokensAfter)
			originatorL2BaseTokensAfter := ch.L2BaseTokens(ch.AdminAgentID())
			t.Logf("originatorL2BaseTokensBefore: %d, originatorL2BaseTokensAfter: %d", originatorL2BaseTokensBefore, originatorL2BaseTokensAfter)
			l1GasFee := coin.Value(l1Res.Effects.Data.GasFee())
			l2GasFee := ch.LastReceipt().GasFeeCharged
			t.Logf("l1GasFee: %d, l2GasFee: %d", l1GasFee, l2GasFee)

			require.NotZero(ch.Env.T, l2GasFee)
			// the common account is topped up to isc.GasCoinTargetValue,
			// taking from the L2 gas fee
			addedToCommonAccount := min(l2GasFee, isc.GasCoinTargetValue-commonAccountBaseTokensBefore)
			// the gas coin is topped up to GasCoinTargetValue, taking from the
			// common account
			deductedForGasCoin := min(commonAccountBaseTokensBefore+addedToCommonAccount, isc.GasCoinTargetValue-gasCoinL1Before)
			// total L2 assets is increased by 1 mil minus the amount deducted
			// to top up the gas coin
			require.Equal(t, l2TotalBaseTokensBefore+1*isc.Million-deductedForGasCoin, l2TotalBaseTokensAfter)
			require.Equal(t, commonAccountBaseTokensBefore+addedToCommonAccount-deductedForGasCoin, commonAccountBaseTokensAfter)
			if test.SenderIsOriginator {
				// sender got 1mil (rest of l2GasFee goes to originator which is the sender)
				require.Equal(t, senderL2BaseTokensBefore+1*isc.Million-addedToCommonAccount, senderL2BaseTokensAfter)
			} else {
				// sender got 1mil minus l2GasFee
				require.Equal(t, senderL2BaseTokensBefore+1*isc.Million-l2GasFee, senderL2BaseTokensAfter)
				// l2GasFee goes to payoutAgentID (originator by default)
				require.Equal(t, originatorL2BaseTokensBefore+l2GasFee-addedToCommonAccount, originatorL2BaseTokensAfter)
			}
		})
	}
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

func TestSandboxStackOverflow(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
	chain := env.NewChain()
	_, err := chain.PostRequestSync(
		solo.NewCallParams(sbtestsc.FuncStackOverflow.Message()).
			WithGasBudget(math.MaxUint64),
		nil,
	)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
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
				// deposit must come from another user so that we have exactly
				// the funds we need on the test account
				anotherKeyPair, _ := env.NewKeyPairWithFunds()
				err := ch.TransferAllowanceTo(
					isc.NewAssets(testCase.L2Balance),
					agentID,
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
	t.Skipf("This test needs to be properly validated and fixed. Its only temporarily deactivated.")

	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(10000, nil)

	initialBlockIndex := ch.GetLatestBlockInfo().BlockIndex

	// Higher values cause execution errors, probably due to the Gas requirements.
	reqSize := 128                      // bytes
	MaxPayloadSize := 128 * 1024 * 1024 // 120kB
	var attachedBaseTokens coin.Value = 1

	maxRequestsPerBlock := MaxPayloadSize / reqSize

	reqs := make([]isc.Request, maxRequestsPerBlock+1)

	for i := range reqs {
		req, _, err := ch.SendRequest(
			solo.NewCallParams(errors.FuncRegisterError.Message(string(rune(i)))).
				AddBaseTokens(attachedBaseTokens).
				AddAllowanceBaseTokens(attachedBaseTokens).
				WithMaxAffordableGasBudget(),
			nil,
		)
		require.NoError(t, err)

		reqs[i] = req
	}

	// TODO properly test this:
	// request outputs are so large that they have to be processed in two separate blocks
	_, results := ch.RunRequestBatch(maxRequestsPerBlock)
	require.Len(t, results, maxRequestsPerBlock)
	_, results = ch.RunRequestBatch(maxRequestsPerBlock)
	require.Len(t, results, 1)
	require.Equal(t, initialBlockIndex+2, ch.GetLatestBlockInfo().BlockIndex)

	for _, req := range reqs {
		receipt, _ := ch.GetRequestReceipt(req.ID())
		require.Nil(t, receipt.Error)
	}
}

func TestInvalidSignatureRequestsAreNotProcessed(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	// produce a badly signed off-ledger request
	req := isc.NewOffLedgerRequest(
		ch.ChainID,
		isc.NewMessage(isc.Hn("contract"), isc.Hn("entrypoint"), nil),
		0,
		math.MaxUint64,
	).WithSender(ch.ChainAdmin.GetPublicKey())

	require.ErrorContains(t, req.VerifySignature(), "invalid signature")

	_, _, err := ch.RunOffLedgerRequest(req)
	require.ErrorContains(t, err, "request was skipped")
}

func TestInvalidAllowance(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	_, err := ch.Env.ISCMoveClient().CreateAndSendRequestWithAssets(
		ch.Env.Ctx(),
		&iscmoveclient.CreateAndSendRequestWithAssetsRequest{
			Signer:        ch.ChainAdmin,
			PackageID:     ch.Env.ISCPackageID(),
			AnchorAddress: ch.ID().AsAddress().AsIotaAddress(),
			Assets:        isc.NewAssets(1 * isc.Million).AsISCMove(),
			Message: &iscmove.Message{
				Contract: uint32(accounts.Contract.Hname()),
				Function: uint32(accounts.FuncDeposit.Hname()),
				Args:     [][]byte{},
			},
			AllowanceBCS:     []byte{1, 2, 3}, // invalid allowance
			OnchainGasBudget: math.MaxUint64,
			GasPrice:         iotaclient.DefaultGasPrice,
			GasBudget:        iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	n := ch.RunAllReceivedRequests(1)
	require.EqualValues(t, 1, n)
	receipt := ch.LastReceipt()
	require.ErrorContains(t, ch.ResolveVMError(receipt.Error).AsGoError(), "invalid allowance")
}

func TestBatchWithSkippedRequestsReceipts(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()
	user, _ := env.NewKeyPairWithFunds()
	err := ch.DepositAssetsToL2(isc.NewAssets(10*isc.Million), user)
	require.NoError(t, err)

	// create a request with an invalid nonce that must be skipped
	skipReq := isc.NewOffLedgerRequest(ch.ChainID, isc.NewMessage(isc.Hn("contract"), isc.Hn("entrypoint"), nil), 0, math.MaxUint64).WithNonce(9999).Sign(user)
	validReq := isc.NewOffLedgerRequest(ch.ChainID, isc.NewMessage(isc.Hn("contract"), isc.Hn("entrypoint"), nil), 0, math.MaxUint64).WithNonce(0).Sign(user)

	ch.RunRequestsSync([]isc.Request{skipReq, validReq})

	// block has been created with only 1 request, calling 	`GetRequestReceiptsForBlock` must yield 1 receipt as expected
	bi := ch.GetLatestBlockInfo()
	require.EqualValues(t, 1, bi.TotalRequests)
	receipts := ch.GetRequestReceiptsForBlock(bi.BlockIndex)
	require.Len(t, receipts, 1)
}
