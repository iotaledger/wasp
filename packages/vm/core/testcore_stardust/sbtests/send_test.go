package sbtests

import (
	"math/big"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/stretchr/testify/require"
)

func TestTooManyOutputsInASingleCall(t *testing.T) { run2(t, testTooManyOutputsInASingleCall) }
func testTooManyOutputsInASingleCall(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	// send 1 tx will 1_000_000 iotas which should result in too many outputs, so the request must fail
	wallet, address := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))
	_, err := ch.Env.GetFundsFromFaucet(address, 10_000_000)
	require.NoError(t, err)

	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFunds.Name).
		AddAssetsIotas(10_000_000).
		AddIotaAllowance(40_000). // 40k iotas = 200 outputs
		WithGasBudget(10_000_000)
	_, err = ch.PostRequestSync(req, wallet)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vmcontext.ErrExceededPostedOutputLimit)
	require.NotContains(t, err.Error(), "skipped")
}

func TestSeveralOutputsInASingleCall(t *testing.T) { run2(t, testSeveralOutputsInASingleCall) }
func testSeveralOutputsInASingleCall(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, walletAddr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	err := ch.DepositIotasToL2(100_000, wallet)
	require.NoError(t, err)

	beforeWallet := ch.L1L2Funds(walletAddr)
	t.Logf("----- BEFORE wallet: %s", beforeWallet)

	// this will SUCCEED because it will result in 4 = 800/200 outputs in the single call
	const allowance = 800
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFunds.Name).
		AddIotaAllowance(allowance).
		WithGasBudget(200_000)
	tx, _, err := ch.PostRequestSyncTx(req, wallet)
	require.NoError(t, err)

	dustDeposit := tx.Essence.Outputs[0].Deposit()
	ch.Env.AssertL1Iotas(walletAddr, beforeWallet.AssetsL1.Iotas+allowance-dustDeposit)
}

func TestSeveralOutputsInASingleCallFail(t *testing.T) { run2(t, testSeveralOutputsInASingleCallFail) }
func testSeveralOutputsInASingleCallFail(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, walletAddr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	err := ch.DepositIotasToL2(100_000, wallet)
	require.NoError(t, err)

	beforeWallet := ch.L1L2Funds(walletAddr)
	t.Logf("----- BEFORE wallet: %s", beforeWallet)

	// this will FAIL because it will result in 1000/200 = 5 outputs in the single call
	const allowance = 1000
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFunds.Name).
		AddIotaAllowance(allowance).
		WithGasBudget(200_000)
	_, err = ch.PostRequestSync(req, wallet)
	testmisc.RequireErrorToBe(t, err, vmcontext.ErrExceededPostedOutputLimit)
	require.NotContains(t, err.Error(), "skipped")
}

func TestSplitTokensFail(t *testing.T) { run2(t, testSplitTokensFail) }
func testSplitTokensFail(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, _ := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	err := ch.DepositIotasToL2(100_000, wallet)
	require.NoError(t, err)

	sn, tokenID, err := ch.NewFoundryParams(100).
		WithUser(wallet).
		CreateFoundry()
	require.NoError(t, err)
	err = ch.MintTokens(sn, big.NewInt(100), wallet)
	require.NoError(t, err)

	// this will FAIL because it will result in 100 outputs in the single call
	allowance := iscp.NewAssetsIotas(100_000).AddNativeTokens(tokenID, 100)
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFundsNativeTokens.Name).
		AddAllowance(allowance).
		AddAssetsIotas(100_000).
		WithGasBudget(200_000)
	_, err = ch.PostRequestSync(req, wallet)
	testmisc.RequireErrorToBe(t, err, vmcontext.ErrExceededPostedOutputLimit)
	require.NotContains(t, err.Error(), "skipped")
}

func TestSplitTokensSuccess(t *testing.T) { run2(t, testSplitTokensSuccess) }
func testSplitTokensSuccess(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, addr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))
	agentID := iscp.NewAgentID(addr, 0)

	err := ch.DepositIotasToL2(100_000, wallet)
	require.NoError(t, err)

	amountMintedTokens := int64(100)
	sn, tokenID, err := ch.NewFoundryParams(amountMintedTokens).
		WithUser(wallet).
		CreateFoundry()
	require.NoError(t, err)
	err = ch.MintTokens(sn, big.NewInt(amountMintedTokens), wallet)
	require.NoError(t, err)

	amountTokensToSend := int64(3)
	// this will FAIL because it will result in 100 outputs in the single call
	allowance := iscp.NewAssetsIotas(100_000).AddNativeTokens(tokenID, amountTokensToSend)
	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFundsNativeTokens.Name).
		AddAllowance(allowance).
		AddAssetsIotas(100_000).
		WithGasBudget(200_000)
	_, err = ch.PostRequestSync(req, wallet)
	require.NoError(t, err)
	require.Equal(t, ch.L2NativeTokens(agentID, &tokenID).Int64(), amountMintedTokens-amountTokensToSend)
	require.Equal(t, ch.Env.L1NativeTokens(addr, &tokenID).Int64(), amountTokensToSend)
}

// TestPingIotas1 sends some iotas to SC and receives the whole allowance sent back to L1 as on-ledger request
func TestPingIotas1(t *testing.T) { run2(t, testPingIotas1) }

func testPingIotas1(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	user, userAddr := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(1))

	userFundsBefore := ch.L1L2Funds(userAddr)
	commonBefore := ch.L2CommonAccountAssets()
	t.Logf("----- BEFORE -----\nUser funds left: %s\nCommon account: %s", userFundsBefore, commonBefore)

	const expectedBack = 1_000
	ch.Env.AssertL1Iotas(userAddr, solo.Saldo)

	reqEstimate := solo.NewCallParams(ScName, sbtestsc.FuncPingAllowanceBack.Name).
		AddAssetsIotas(100_000).
		AddIotaAllowance(expectedBack).
		WithGasBudget(100_000)

	gasEstimate, feeEstimate, err := ch.EstimateGasOnLedger(reqEstimate, user)
	require.NoError(t, err)
	t.Logf("gasEstimate: %d, feeEstimate: %d", gasEstimate, feeEstimate)

	req := solo.NewCallParams(ScName, sbtestsc.FuncPingAllowanceBack.Name).
		AddAssetsIotas(feeEstimate + expectedBack).
		AddIotaAllowance(expectedBack).
		WithGasBudget(gasEstimate)

	_, err = ch.PostRequestSync(req, user)
	require.NoError(t, err)
	rec := ch.LastReceipt()

	userFundsAfter := ch.L1L2Funds(userAddr)
	commonAfter := ch.L2CommonAccountAssets()
	t.Logf("------ AFTER ------\nReceipt: %s\nUser funds left: %s\nCommon account: %s", rec, userFundsAfter, commonAfter)

	require.EqualValues(t, userFundsAfter.AssetsL1.Iotas, solo.Saldo-feeEstimate)
	require.EqualValues(t, int(commonBefore.Iotas+rec.GasFeeCharged), int(commonAfter.Iotas))
	require.EqualValues(t, feeEstimate-rec.GasFeeCharged, int(userFundsAfter.AssetsL2.Iotas))
}

func TestEstimateMinimumDust(t *testing.T) { run2(t, testEstimateMinimumDust) }
func testEstimateMinimumDust(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	wallet, _ := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(20))

	// should fail without enough iotas to pay for a L1 transaction dust
	allowance := iscp.NewAssetsIotas(1)
	req := solo.NewCallParams(ScName, sbtestsc.FuncEstimateMinDust.Name).
		AddAllowance(allowance).
		AddAssetsIotas(100_000).
		WithGasBudget(100_000)

	_, err := ch.PostRequestSync(req, wallet)
	require.Error(t, err)

	// should succeed with enough iotas to pay for a L1 transaction dust
	allowance = iscp.NewAssetsIotas(100_000)
	req = solo.NewCallParams(ScName, sbtestsc.FuncEstimateMinDust.Name).
		AddAllowance(allowance).
		AddAssetsIotas(100_000).
		WithGasBudget(100_000)

	_, err = ch.PostRequestSync(req, wallet)
	require.NoError(t, err)
}
