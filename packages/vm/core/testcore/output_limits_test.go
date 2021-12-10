package testcore

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/stretchr/testify/require"
)

const paramShouldEmitEvent = "e"

var (
	manyOutputsContract = coreutil.NewContract("Many outputs", "")

	funcSplitFunds = coreutil.Func("splitFunds")

	manyOutputsProcessor = manyOutputsContract.Processor(nil,
		// splits incoming puts into many outputs (chunks of 1000)
		funcSplitFunds.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
			a := assert.NewAssert(ctx.Log())
			par := kvdecoder.New(ctx.Params())
			shouldEmitEvent := par.MustGetInt16(paramShouldEmitEvent, 0) == 1
			iotas := ctx.IncomingTransfer()[colored.IOTA]
			for i := uint64(0); i < iotas; i += 1000 {
				ret := ctx.Send(
					ctx.Caller().Address(),
					colored.NewBalancesForColor(colored.IOTA, 1000),
					nil,
				)
				a.Require(ret == true, "failed to send funds")
				if shouldEmitEvent {
					ctx.Event(fmt.Sprintf("reqid: %s, sent funds", ctx.Request().ID()))
				}
			}
			return nil, nil
		}),
	)
)

func TestTooManyOutputsInASingleCall(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(manyOutputsProcessor)
	ch := env.NewChain(nil, "chain1")
	err := ch.DeployContract(nil, manyOutputsContract.Name, manyOutputsContract.ProgramHash)
	require.NoError(t, err)

	// send 1 tx will 1_000_000 iotas which should result in too mant outputs, so the request must fail
	wallet, address := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))
	initialBalance := env.GetAddressBalance(address, colored.IOTA)

	_, err = ch.PostRequestSync(
		solo.NewCallParams(manyOutputsContract.Name, funcSplitFunds.Name).WithIotas(1000000),
		wallet,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeded max number of allowed outputs")

	finalBalance := env.GetAddressBalance(address, colored.IOTA)
	require.Equal(t, finalBalance, initialBalance)
}

func TestTooManyOutputsInBlock(t *testing.T) {
	// Set the maximum number of outputs allowed in a block to a low value, to force the batch to overflow in produced outputs (we need this to test that the "extra" requests don't get rejected and instead are processed in the next batch)
	vmcontext.MaxBlockOutputCount = 10

	env := solo.New(t, false, false).WithNativeContract(manyOutputsProcessor)
	ch := env.NewChain(nil, "chain1")
	err := ch.DeployContract(nil, manyOutputsContract.Name, manyOutputsContract.ProgramHash)
	require.NoError(t, err)

	wallet, address := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))
	initialBalance := env.GetAddressBalance(address, colored.IOTA)

	// send 500 tx with 1000 iotas each. all should be processed successfully, but must be split into different blocks (async post)
	nReqs := 500
	txs := make([]*ledgerstate.Transaction, nReqs)
	for i := 0; i < nReqs; i++ {
		req := solo.NewCallParams(
			manyOutputsContract.Name, funcSplitFunds.Name,
			paramShouldEmitEvent, uint16(1),
		).WithIotas(1000)
		txs[i], _, err = ch.RequestFromParamsToLedger(req, wallet)
		require.NoError(t, err)
	}

	for i := 0; i < nReqs; i++ {
		go ch.Env.EnqueueRequests(txs[i])
	}
	require.True(t, ch.WaitForRequestsThrough(nReqs+2, 10*time.Second)) // + 2 because of sc deploy and core contracts deployment

	contractEvents, err := ch.GetEventsForContract(manyOutputsContract.Name)
	require.NoError(t, err)
	require.Len(t, contractEvents, nReqs)

	for _, tx := range txs {
		reqsInTx, err := ch.Env.RequestsForChain(tx, ch.ChainID)
		reqID := reqsInTx[0].ID()
		require.NoError(t, err)
		events, err := ch.GetEventsForRequest(reqID)
		require.NoError(t, err)
		require.Len(t, events, 1)

		rec, _, _, ok := ch.GetRequestReceipt(reqID)
		require.True(ch.Env.T, ok)
		require.Len(t, rec.Error, 0)
	}

	lastBlock := ch.GetLatestBlockInfo()
	// block #3 is the first block that includes "non-chain-start" events
	for i := uint32(3); i <= lastBlock.BlockIndex; i++ {
		blockInfo, _ := ch.GetBlockInfo(i)
		// all requests must be successful
		require.Equal(t, blockInfo.TotalRequests, blockInfo.NumSuccessfulRequests)
		require.LessOrEqual(t, blockInfo.TotalRequests, uint16(vmcontext.MaxBlockOutputCount))
	}

	finalBalance := env.GetAddressBalance(address, colored.IOTA)

	// check balance matches the expected
	require.Equal(t, initialBalance, finalBalance) // all iotas should have been sent back
}
