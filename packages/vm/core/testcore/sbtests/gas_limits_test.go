package sbtests

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/sbtests/sbtestsc"
)

func infiniteLoopRequest(ch *solo.Chain, gasBudget ...uint64) (*solo.CallParams, *cryptolib.KeyPair) {
	budget := uint64(math.MaxUint64)
	if len(gasBudget) > 0 {
		budget = gasBudget[0]
	}
	wallet, address := ch.Env.NewKeyPairWithFunds()
	baseTokensToSend := ch.Env.L1BaseTokens(address) / 10

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncInfiniteLoop.Name).
		AddBaseTokens(baseTokensToSend).
		WithGasBudget(budget)
	return req, wallet
}

func TestTxWithGasOverLimit(t *testing.T) {
	// create a TX that would use more than max gas limit, assert that only the maximum will be used
	_, ch := setupChain(t)
	setupTestSandboxSC(t, ch, nil)

	req, wallet := infiniteLoopRequest(ch)
	_, err := ch.PostRequestSync(req, wallet)
	require.Error(t, err) // tx expected to run out of gas
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
	receipt := ch.LastReceipt()
	// assert that the submitted gas budget was limited to the max per call
	require.Less(t, receipt.GasBurned, req.GasBudget())
	require.GreaterOrEqual(t, receipt.GasBurned, ch.GetGasLimits().MaxGasPerRequest) // should exceed MaxGasPerRequest() by 1 operation
}

func TestBlockGasOverflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// queue many transactions with enough gas to fill a block, assert that they are split across blocks
	_, ch := setupChain(t)
	setupTestSandboxSC(t, ch, nil)
	initialBlockInfo := ch.GetLatestBlockInfo()

	// produce 1 request over the block gas limit (each request uses the maximum amount of gas a call can use)
	limits := ch.GetGasLimits()
	nRequests := int(limits.MaxGasPerBlock/limits.MaxGasPerRequest) + 1

	for i := 0; i < nRequests; i++ {
		req, wallet := infiniteLoopRequest(ch)
		ch.SendRequest(req, wallet)
	}

	const maxRequestsPerBlock = 50
	ch.RunAllReceivedRequests(maxRequestsPerBlock)

	// we should have produced 2 blocks
	require.EqualValues(t, initialBlockInfo.BlockIndex+2, ch.LatestBlockIndex())

	fullGasBlockInfo, err := ch.GetBlockInfo(initialBlockInfo.BlockIndex + 1)
	require.NoError(t, err)
	// the request number #{nRequests} should overflow the block and be moved to the next one
	require.Equal(t, nRequests-1, int(fullGasBlockInfo.TotalRequests))
	// gas burned will be sightly below the limit
	require.LessOrEqual(t, fullGasBlockInfo.GasBurned, limits.MaxGasPerBlock)

	// 1 requests should be moved to the next block
	followingBlockInfo, err := ch.GetBlockInfo(initialBlockInfo.BlockIndex + 2)
	require.NoError(t, err)
	require.Equal(t, uint16(1), followingBlockInfo.TotalRequests)
}

func TestGasBudget(t *testing.T) {
	// create a TX with not enough gas, assert the receipt is as expected
	_, ch := setupChain(t)
	setupTestSandboxSC(t, ch, nil)

	limits := ch.GetGasLimits()
	gasBudget := limits.MaxGasPerRequest / 2
	req, wallet := infiniteLoopRequest(ch, gasBudget)
	_, err := ch.PostRequestSync(req, wallet)
	require.Error(t, err) // tx expected to run out of gas
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
	receipt := ch.LastReceipt()
	require.EqualValues(t, gasBudget, receipt.GasBudget)

	// repeat with gas budget 0
	gasBudget = 0
	req, wallet = infiniteLoopRequest(ch, gasBudget)
	_, err = ch.PostRequestSync(req, wallet)
	require.Error(t, err) // tx expected to run out of gas
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
	receipt = ch.LastReceipt()
	require.EqualValues(t, limits.MinGasPerRequest, receipt.GasBudget) // gas budget should be adjusted to the minimum
}

func TestViewGasLimit(t *testing.T) {
	_, ch := setupChain(t)
	setupTestSandboxSC(t, ch, nil)
	_, err := ch.CallViewEx(sbtestsc.Contract.Name, sbtestsc.FuncInfiniteLoopView.Name)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
}
