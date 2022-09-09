package sbtests

import (
	"math"
	"testing"

	"github.com/iotaledger/wasp/packages/vm"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/stretchr/testify/require"
)

func maxGasRequest(ch *solo.Chain, seedIndex int) (*solo.CallParams, *cryptolib.KeyPair) {
	wallet, address := ch.Env.NewKeyPairWithFunds(ch.Env.NewSeedFromIndex(seedIndex))
	baseTokensToSend := ch.Env.L1BaseTokens(address)

	req := solo.NewCallParams(ScName, sbtestsc.FuncInfiniteLoop.Name).
		AddBaseTokens(baseTokensToSend).
		WithGasBudget(math.MaxUint64)
	return req, wallet
}

// create a TX that would use more than max gas limit, assert that only the maximum will be used
func TestTxWithGasOverLimit(t *testing.T) { run2(t, testTxWithGasOverLimit) }

func testTxWithGasOverLimit(t *testing.T, w bool) {
	if w { // TODO the WASM version of this must be tested.
		t.SkipNow()
	}
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)

	req, wallet := maxGasRequest(ch, 2)
	_, err := ch.PostRequestSync(req, wallet)
	require.Error(t, err) // tx expected to run out of gas
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
	receipt := ch.LastReceipt()
	// assert that the submitted gas budget was limited to the max per call
	require.Less(t, receipt.GasBurned, req.GasBudget())
	require.GreaterOrEqual(t, receipt.GasBurned, gas.MaxGasPerCall) // should exceed MaxGasPerCall by 1 operation
}

// queue many transactions with enough gas to fill a block, assert that they are split across blocks
func TestBlockGasOverflow(t *testing.T) { run2(t, testBlockGasOverflow) }

func testBlockGasOverflow(t *testing.T, w bool) {
	if forceGoNoWasm && w {
		t.SkipNow()
	}
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)
	initialBlockInfo := ch.GetLatestBlockInfo()

	// produce n requests over the block gas limit (each request uses the maximum amount of gas a call can use)
	nRequests := int(gas.MaxGasPerBlock / gas.MaxGasPerCall)
	reqs := make([]isc.Request, nRequests)

	for i := 0; i < nRequests; i++ {
		req, wallet := maxGasRequest(ch, i)
		iscReq, err := solo.NewIscRequestFromCallParams(ch, req, wallet)
		require.NoError(t, err)
		reqs[i] = iscReq
	}

	ch.Env.AddRequestsToChainMempoolWaitUntilInbufferEmpty(ch, reqs)
	ch.WaitUntilMempoolIsEmpty()

	fullGasBlockInfo, err := ch.GetBlockInfo(initialBlockInfo.BlockIndex + 1)
	require.NoError(t, err)
	// the request number #{nRequests} should overflow the block and be moved to the next one
	require.Equal(t, int(fullGasBlockInfo.TotalRequests), nRequests-1)
	// gas burned will be sightly below the limit
	require.LessOrEqual(t, fullGasBlockInfo.GasBurned, gas.MaxGasPerBlock)

	// 1 requests should be moved to the next block
	followingBlockInfo, err := ch.GetBlockInfo(initialBlockInfo.BlockIndex + 2)
	require.NoError(t, err)
	require.Equal(t, followingBlockInfo.TotalRequests, uint16(1))

	// no further blocks should have been produced
	_, err = ch.GetBlockInfo(initialBlockInfo.BlockIndex + 3)
	require.Error(t, err)
}

func TestViewGasBlock(t *testing.T) { run2(t, testViewGasBlock) }
func testViewGasBlock(t *testing.T, w bool) {
	if forceGoNoWasm && w {
		t.SkipNow()
	}
	_, ch := setupChain(t, nil)
	setupTestSandboxSC(t, ch, nil, w)
	_, err := ch.CallView(sbtestsc.Contract.Name, sbtestsc.FuncInfiniteLoopView.Name)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
}
