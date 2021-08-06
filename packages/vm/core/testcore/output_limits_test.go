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
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

var n = 0

var (
	manyOutputsContract = coreutil.NewContract("Many outputs", "")

	funcSplitFunds = coreutil.Func("splitFunds")

	manyOutputsProcessor = manyOutputsContract.Processor(nil,
		// splits incoming puts into many outputs (1 per iota)
		funcSplitFunds.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
			a := assert.NewAssert(ctx.Log())
			iotas := ctx.IncomingTransfer()[colored.IOTA]
			for i := uint64(0); i < iotas; i++ {
				ret := ctx.Send(
					ctx.Caller().Address(),
					colored.NewBalancesForColor(colored.IOTA, 1),
					nil,
				)
				a.Require(ret == true, "failed to send funds")
				ctx.Event(fmt.Sprintf("sent funds, i: %d", i))
			}
			println("!!!!!!!!!!!!!!!!")
			println(n)
			println("!!!!!!!!!!!!!!!!")
			n++
			return nil, nil
		}),
	)
)

func TestTooManyOutputsInASingleCall(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(manyOutputsProcessor)
	ch := env.NewChain(nil, "chain1")
	err := ch.DeployContract(nil, manyOutputsContract.Name, manyOutputsContract.ProgramHash)
	require.NoError(t, err)
	// send 1 tx will 1000 iotas which should result in 999 outputs, so the request must fail
	_, err = ch.PostRequestSync(
		solo.NewCallParams(manyOutputsContract.Name, funcSplitFunds.Name).WithIotas(1000),
		nil,
	)
	require.Error(t, err)
}

func TestTooManyOutputsInBlock(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(manyOutputsProcessor)
	ch := env.NewChain(nil, "chain1")
	err := ch.DeployContract(nil, manyOutputsContract.Name, manyOutputsContract.ProgramHash)
	require.NoError(t, err)

	// send 100 tx with 10 iotas each. all should be processed successfully, but must be split into different blocks (async post)
	nReqs := 100
	txs := make([]*ledgerstate.Transaction, nReqs)
	for i := 0; i < nReqs; i++ {
		req := solo.NewCallParams(manyOutputsContract.Name, funcSplitFunds.Name).WithIotas(10)
		txs[i], _, err = ch.RequestFromParamsToLedger(req, nil)
		require.NoError(t, err)
	}

	for i := 0; i < nReqs; i++ {
		go ch.Env.EnqueueRequests(txs[i])
	}
	require.True(t, ch.WaitForRequestsThrough(nReqs+2, 10*time.Second)) // + 2 because of sc deploy and core contracts deployment
	events, err := ch.GetEventsForContract(manyOutputsContract.Name)
	require.NoError(t, err)
	println(events)

	for i, tx := range txs {
		reqsInTx, err := ch.Env.RequestsForChain(tx, ch.ChainID)
		reqID := reqsInTx[0].ID()
		require.NoError(t, err)
		events, err := ch.GetEventsForRequest(reqID)
		require.NoError(t, err)
		println(fmt.Sprintf("events for req #%d: %v", i, events))
	}

	lastBlock := ch.GetLatestBlockInfo()
	// block #3 is the first block that includes "non-chain-start" events
	for i := uint32(3); i <= lastBlock.BlockIndex; i++ {
		blockInfo, _ := ch.GetBlockInfo(lastBlock.BlockIndex - 1)
		// all requests must be successful
		require.Equal(t, blockInfo.TotalRequests, blockInfo.NumSuccessfulRequests)
		events, err := ch.GetEventsForBlock(i)
		require.NoError(t, err)
		println(events)
	}
}
