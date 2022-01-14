package testcore

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/stretchr/testify/require"
)

var (
	nEvents                = int(governance.DefaultMaxEventsPerRequest + 1000)
	bigEventSize           = int(governance.DefaultMaxEventSize + 1000)
	manyEventsContractName = "ManyEventsContract"
	manyEventsContract     = coreutil.NewContract(manyEventsContractName, "many events contract")

	funcManyEvents = coreutil.Func("manyevents")
	funcBigEvent   = coreutil.Func("bigevent")

	manyEventsContractProcessor = manyEventsContract.Processor(nil,
		funcManyEvents.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
			for i := 0; i < nEvents; i++ {
				ctx.Event(fmt.Sprintf("testing many events %d", i))
			}
			return nil, nil
		}),
		funcBigEvent.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
			buf := make([]byte, bigEventSize)
			ctx.Event(string(buf))
			return nil, nil
		}),
	)
)

func setupTest(t *testing.T) *solo.Chain {
	env := solo.New(t, true, false).WithNativeContract(manyEventsContractProcessor)
	ch := env.NewChain(nil, "ch")
	err := ch.DeployContract(nil, manyEventsContract.Name, manyEventsContract.ProgramHash)
	require.NoError(t, err)
	return ch
}

func checkNEvents(t *testing.T, ch *solo.Chain, reqid iscp.RequestID, n int) {
	// fetch events from blocklog
	events, err := ch.GetEventsForContract(manyEventsContractName)
	require.NoError(t, err)
	require.Len(t, events, n)

	events, err = ch.GetEventsForRequest(reqid)
	require.NoError(t, err)
	require.Len(t, events, n)
}

func TestManyEvents(t *testing.T) {
	ch := setupTest(t)

	// post a request that issues too many events (nEvents)
	tx, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(manyEventsContract.Name, funcManyEvents.Name).AddAssetsIotas(1),
		nil,
	)
	require.Error(t, err) // error expected (too many events)
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(t, err)
	reqID := reqs[0].ID()
	checkNEvents(t, ch, reqID, 0) // no events are saved

	// allow for more events per request in root contract
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name, governance.FuncSetChainInfo.Name,
			governance.ParamMaxEventsPerRequestUint16, uint16(nEvents),
		).AddAssetsIotas(1),
		nil,
	)
	require.NoError(t, err)

	// check events are now saved
	tx, _, err = ch.PostRequestSyncTx(
		solo.NewCallParams(manyEventsContract.Name, funcManyEvents.Name).AddAssetsIotas(1),
		nil,
	)
	require.NoError(t, err)
	reqs, err = ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(t, err)
	reqID = reqs[0].ID()
	checkNEvents(t, ch, reqID, nEvents)
}

func TestEventTooLarge(t *testing.T) {
	ch := setupTest(t)

	// post a request that issues an event too large
	tx, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(manyEventsContract.Name, funcBigEvent.Name).AddAssetsIotas(1),
		nil,
	)
	require.Error(t, err) // error expected (event too large)
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(t, err)
	reqID := reqs[0].ID()
	checkNEvents(t, ch, reqID, 0) // no events are saved

	// allow for bigger events in root contract
	_, err = ch.PostRequestSync(
		solo.NewCallParams(
			governance.Contract.Name, governance.FuncSetChainInfo.Name,
			governance.ParamMaxEventSizeUint16, uint16(bigEventSize),
		).AddAssetsIotas(1),
		nil,
	)
	require.NoError(t, err)

	// check event is now saved
	tx, _, err = ch.PostRequestSyncTx(
		solo.NewCallParams(manyEventsContract.Name, funcBigEvent.Name).AddAssetsIotas(1),
		nil,
	)
	require.NoError(t, err)
	reqs, err = ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(t, err)
	reqID = reqs[0].ID()
	checkNEvents(t, ch, reqID, 1)
}
