package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/testcore"
	"github.com/stretchr/testify/require"
)

// TODO test analysis not finished

func incrementSCCounter(t *testing.T, ch *solo.Chain) iscp.RequestID {
	tx, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(inccounter.Contract.Name, inccounter.FuncIncCounter.Name),
		nil,
	)
	require.NoError(t, err)
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(t, err)
	return reqs[0].ID()
}

func getEventsForRequest(t *testing.T, chain *solo.Chain, reqID iscp.RequestID) []string {
	res, err := chain.CallView(blocklog.Contract.Name, blocklog.FuncGetEventsForRequest.Name,
		blocklog.ParamRequestID, reqID,
	)
	require.NoError(t, err)
	events, err := testcore.EventsViewResultToStringArray(res)
	require.NoError(t, err)
	return events
}

func getEventsForBlock(t *testing.T, chain *solo.Chain, blockNumber ...int32) []string {
	var res dict.Dict
	var err error
	if len(blockNumber) > 0 {
		res, err = chain.CallView(blocklog.Contract.Name, blocklog.FuncGetEventsForBlock.Name,
			blocklog.ParamBlockIndex, blockNumber[0],
		)
	} else {
		res, err = chain.CallView(blocklog.Contract.Name, blocklog.FuncGetEventsForBlock.Name)
	}
	require.NoError(t, err)
	events, err := testcore.EventsViewResultToStringArray(res)
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
	events, err := testcore.EventsViewResultToStringArray(res)
	require.NoError(t, err)
	return events
}

func TestGetEvents(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true}).
		WithNativeContract(inccounter.Processor)
	ch := env.NewChain(nil, "chain1")

	err := ch.DepositIotasToL2(10_000, nil)
	require.NoError(t, err)

	err = ch.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash, inccounter.VarCounter, 0)
	require.NoError(t, err)

	// block 1 = ch init
	// block 2 = inccounter contract deployment
	reqID1 := incrementSCCounter(t, ch) // #block 3
	reqID2 := incrementSCCounter(t, ch) // #block 4
	reqID3 := incrementSCCounter(t, ch) // #block 5

	events := getEventsForRequest(t, ch, reqID1)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 1")
	events = getEventsForRequest(t, ch, reqID2)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 2")
	events = getEventsForRequest(t, ch, reqID3)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 3")

	events = getEventsForBlock(t, ch, 3)
	require.Len(t, events, 2)
	require.Contains(t, events[0], "counter = 0")
	events = getEventsForBlock(t, ch, 4)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 1")
	events = getEventsForBlock(t, ch, 5)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 2")
	events = getEventsForBlock(t, ch)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 3")

	events = getEventsForSC(t, ch, 0, 1000)
	require.Len(t, events, 4)
	require.Contains(t, events[0], "counter = 0")
	require.Contains(t, events[1], "counter = 1")
	require.Contains(t, events[2], "counter = 2")
	require.Contains(t, events[3], "counter = 3")
	events = getEventsForSC(t, ch, 2, 3)
	require.Len(t, events, 1)
	require.Contains(t, events[0], "counter = 0")
}
