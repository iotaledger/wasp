package sbtests

import (
	"testing"

	"github.com/iotaledger/wasp/packages/testutil/testmisc"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"github.com/stretchr/testify/require"
)

func TestTooManyOutputsInASingleCall(t *testing.T) { run2(t, testTooManyOutputsInASingleCall) }
func testTooManyOutputsInASingleCall(t *testing.T, w bool) {
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	// send 1 tx will 1_000_000 iotas which should result in too many outputs, so the request must fail
	wallet, address := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(1))
	_, err := ch.Env.GetFundsFromFaucet(address, 10_000_000)
	require.NoError(t, err)

	req := solo.NewCallParams(ScName, sbtestsc.FuncSplitFunds.Name).
		AddAssetsIotas(10_000_000).
		AddIotaAllowance(40_000). // 40k iotas = 200 outputs
		WithGasBudget(10_000_000)
	_, err = ch.PostRequestSync(req, wallet)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vmtxbuilder.ErrOutputLimitInSingleCallExceeded)
	require.NotContains(t, err.Error(), "skipped")
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
	ch.Env.AssertL1AddressIotas(userAddr, solo.Saldo)

	reqEstimate := solo.NewCallParams(ScName, sbtestsc.FuncPingAllowanceBack.Name).
		AddAssetsIotas(100_000).
		AddIotaAllowance(expectedBack).
		WithGasBudget(100_000)

	gasEstimate, feeEstimate, err := ch.EstimateGas(reqEstimate, user)
	require.NoError(t, err)
	t.Logf("gasEstimate: %d, feeEstimate: %d", gasEstimate, feeEstimate)

	req := solo.NewCallParams(ScName, sbtestsc.FuncPingAllowanceBack.Name).
		AddAssetsIotas(feeEstimate + expectedBack).
		AddIotaAllowance(expectedBack).
		WithGasBudget(gasEstimate)

	receipt, _, err := ch.PostRequestSyncReceipt(req, user)
	require.NoError(t, err)

	userFundsAfter := ch.L1L2Funds(userAddr)
	commonAfter := ch.L2CommonAccountAssets()
	t.Logf("------ AFTER ------\nReceipt: %s\nUser funds left: %s\nCommon account: %s", receipt, userFundsAfter, commonAfter)

	require.EqualValues(t, userFundsAfter.AssetsL1.Iotas, solo.Saldo-feeEstimate)
	require.EqualValues(t, int(commonBefore.Iotas+receipt.GasFeeCharged), int(commonAfter.Iotas))
	require.EqualValues(t, 0, int(userFundsAfter.AssetsL2.Iotas))
}
