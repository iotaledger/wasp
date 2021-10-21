package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/stretchr/testify/require"
)

func TestBlockInfoLatest(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	bi := chain.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 1, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
}

func TestBlockInfo(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	bi, err := chain.GetBlockInfo(0)
	require.NoError(t, err)
	require.NotNil(t, bi)
	require.EqualValues(t, 0, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())

	bi, err = chain.GetBlockInfo(1)
	require.NoError(t, err)
	require.NotNil(t, bi)
	require.EqualValues(t, 1, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())
}

func TestBlockInfoLatestSeveral(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	err = chain.DeployContract(nil, "testCore", hwasm)
	require.NoError(t, err)

	bi := chain.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 3, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
}

func TestBlockInfoSeveral(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	err = chain.DeployContract(nil, "testCore", hwasm)
	require.NoError(t, err)

	bi := chain.GetLatestBlockInfo()

	for blockIndex := uint32(0); blockIndex <= bi.BlockIndex; blockIndex++ {
		bi1, err := chain.GetBlockInfo(blockIndex)
		require.NoError(t, err)
		require.NotNil(t, bi1)
		require.EqualValues(t, blockIndex, bi1.BlockIndex)
		require.EqualValues(t, 1, bi1.TotalRequests)
		require.EqualValues(t, 1, bi1.NumSuccessfulRequests)
		require.EqualValues(t, 0, bi1.NumOffLedgerRequests)
		t.Logf("%s", bi1.String())
	}
}

func TestRequestIsProcessed(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 0, governance.ParamValidatorFee, 0).WithIotas(1)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))
}

func TestRequestLogRecord(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 0, governance.ParamValidatorFee, 0).WithIotas(1)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	rec, blockIndex, requestIndex, ok := chain.GetRequestReceipt(reqs[0].ID())
	require.True(t, ok)
	require.EqualValues(t, reqs[0].ID(), rec.RequestID)
	require.False(t, rec.OffLedger)
	require.EqualValues(t, 0, len(rec.Error))
	require.EqualValues(t, 2, blockIndex)
	require.EqualValues(t, 0, requestIndex)
}

func TestRequestLogRecordsForBlocks(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 0, governance.ParamValidatorFee, 0).WithIotas(1)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	recs := chain.GetRequestReceiptsForBlock(2)
	require.EqualValues(t, 1, len(recs))
	require.EqualValues(t, reqs[0].ID(), recs[0].RequestID)
}

func TestRequestIDsForBlocks(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 0, governance.ParamValidatorFee, 0).WithIotas(1)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	ids := chain.GetRequestIDsForBlock(2)
	require.EqualValues(t, 1, len(ids))
	require.EqualValues(t, reqs[0].ID(), ids[0])
}

/// region test events ----------------------------------------------------------------

func incrementSCCounter(t *testing.T, chain *solo.Chain) iscp.RequestID {
	tx, _, err := chain.PostRequestSyncTx(
		solo.NewCallParams(inccounter.Contract.Name, inccounter.FuncIncCounter.Name).WithIotas(1),
		nil,
	)
	require.NoError(t, err)
	reqs, err := chain.Env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	return reqs[0].ID()
}

func getEventsForRequest(t *testing.T, chain *solo.Chain, reqID iscp.RequestID) []string {
	res, err := chain.CallView(blocklog.Contract.Name, blocklog.FuncGetEventsForRequest.Name,
		blocklog.ParamRequestID, reqID,
	)
	require.NoError(t, err)
	events, err := EventsViewResultToStringArray(res)
	require.NoError(t, err)
	return events
}

func getEventsForBlock(t *testing.T, chain *solo.Chain, blockNumber int32) []string {
	res, err := chain.CallView(blocklog.Contract.Name, blocklog.FuncGetEventsForBlock.Name,
		blocklog.ParamBlockIndex, blockNumber,
	)
	require.NoError(t, err)
	events, err := EventsViewResultToStringArray(res)
	require.NoError(t, err)
	return events
}

func getEventsForSC(t *testing.T, chain *solo.Chain, fromBlock, toBlock int32) []string {
	res, err := chain.CallView(blocklog.Contract.Name, blocklog.FuncGetEventsForContract.Name,
		blocklog.ParamContractHname, inccounter.Contract.Hname(),
		blocklog.ParamFromBlock, fromBlock,
		blocklog.ParamToBlock, toBlock,
	)
	require.NoError(t, err)
	events, err := EventsViewResultToStringArray(res)
	require.NoError(t, err)
	return events
}

func TestGetEvents(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(inccounter.Processor)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash, inccounter.VarCounter, 0)
	require.NoError(t, err)

	// block 1 = chain init
	// block 2 = inccounter contract deployment
	reqID1 := incrementSCCounter(t, chain) // #block 3
	reqID2 := incrementSCCounter(t, chain) // #block 4
	reqID3 := incrementSCCounter(t, chain) // #block 5

	events := getEventsForRequest(t, chain, reqID1)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 1")
	events = getEventsForRequest(t, chain, reqID2)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 2")
	events = getEventsForRequest(t, chain, reqID3)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 3")

	events = getEventsForBlock(t, chain, 3)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 1")
	events = getEventsForBlock(t, chain, 4)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 2")
	events = getEventsForBlock(t, chain, 5)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 3")

	events = getEventsForSC(t, chain, 0, 1000)
	require.Len(t, events, 4)
	require.Contains(t, events[0], "counter = 0")
	require.Contains(t, events[1], "counter = 1")
	require.Contains(t, events[2], "counter = 2")
	require.Contains(t, events[3], "counter = 3")
	events = getEventsForSC(t, chain, 2, 3)
	require.Len(t, events, 2)
	require.Contains(t, events[0], "counter = 0")
	require.Contains(t, events[1], "counter = 1")
}

/// end region ----------------------------------------------------------------
