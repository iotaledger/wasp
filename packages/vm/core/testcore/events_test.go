package testcore

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

var (
	manyEventsContractName = "ManyEventsContract"
	manyEventsContract     = coreutil.NewContract(manyEventsContractName, "many events contract")

	funcManyEvents = coreutil.Func("manyevents")
	funcBigEvent   = coreutil.Func("bigevent")

	manyEventsContractProcessor = manyEventsContract.Processor(nil,
		funcManyEvents.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			n := int(codec.MustDecodeUint32(ctx.Params().MustGet("n")))
			for i := 0; i < n; i++ {
				ctx.Event(fmt.Sprintf("testing many events %d", i))
			}
			return nil
		}),
		funcBigEvent.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			n := int(codec.MustDecodeUint32(ctx.Params().MustGet("n")))
			buf := make([]byte, n)
			ctx.Event(string(buf))
			return nil
		}),
	)
)

func setupTest(t *testing.T) *solo.Chain {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true}).
		WithNativeContract(manyEventsContractProcessor)
	ch := env.NewChain()
	err := ch.DeployContract(nil, manyEventsContract.Name, manyEventsContract.ProgramHash)
	require.NoError(t, err)

	// allow "infinite" gas per request
	limits := ch.GetGasLimits()
	limits.MaxGasPerBlock = math.MaxUint64
	limits.MaxGasPerRequest = math.MaxUint64
	ch.SetGasLimits(ch.OriginatorPrivateKey, limits)

	// set gas very cheap
	fp := ch.GetGasFeePolicy()
	fp.GasPerToken.A = 1000000
	ch.SetGasFeePolicy(ch.OriginatorPrivateKey, fp)

	ch.MustDepositBaseTokensToL2(10_000_000, nil)
	return ch
}

func checkNEvents(t *testing.T, ch *solo.Chain, reqid isc.RequestID, n int) {
	events, err := ch.GetEventsForRequest(reqid)
	require.NoError(t, err)
	require.Len(t, events, n)
}

func getBurnedGas(ch *solo.Chain, tx *iotago.Transaction, err error) (uint64, error) {
	reqs, err2 := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err2)
	require.EqualValues(ch.Env.T, 1, len(reqs))
	if err != nil {
		return 0, err
	}
	receipt, err2 := ch.GetRequestReceipt(reqs[0].ID())
	require.NoError(ch.Env.T, err2)

	return receipt.GasBurned, nil
}

func TestManyEvents(t *testing.T) {
	ch := setupTest(t)

	postEvents := func(n uint32) (uint64, error) {
		// post a request that issues too many events (nEvents)
		tx, _, err := ch.PostRequestSyncTx(
			solo.NewCallParams(manyEventsContract.Name, funcManyEvents.Name, "n", n).
				WithMaxAffordableGasBudget(),
			nil,
		)
		return getBurnedGas(ch, tx, err)
	}

	gas1000, err := postEvents(1000)
	require.NoError(t, err)
	checkNEvents(t, ch, ch.LastReceipt().DeserializedRequest().ID(), 1000)

	gas2000, err := postEvents(2000)
	require.NoError(t, err)
	checkNEvents(t, ch, ch.LastReceipt().DeserializedRequest().ID(), 2000)

	t.Log(gas1000, gas2000)
	require.Greater(t, gas2000, gas1000)

	_, err = postEvents(math.MaxUint16)
	require.NoError(t, err)
	checkNEvents(t, ch, ch.LastReceipt().DeserializedRequest().ID(), math.MaxUint16)

	_, err = postEvents(math.MaxUint16 + 1)
	require.ErrorContains(t, err, "too many events")
	checkNEvents(t, ch, ch.LastReceipt().DeserializedRequest().ID(), 0)
}

func TestEventTooLarge(t *testing.T) {
	ch := setupTest(t)

	postEvent := func(n uint32) (uint64, error) {
		// post a request that issues too many events (nEvents)
		tx, _, err := ch.PostRequestSyncTx(
			solo.NewCallParams(manyEventsContract.Name, funcBigEvent.Name, "n", n).
				WithMaxAffordableGasBudget(),
			nil,
		)
		return getBurnedGas(ch, tx, err)
	}

	gas1k, err := postEvent(100_000)
	require.NoError(t, err)
	checkNEvents(t, ch, ch.LastReceipt().DeserializedRequest().ID(), 1)

	gas2k, err := postEvent(200_000)
	require.NoError(t, err)
	checkNEvents(t, ch, ch.LastReceipt().DeserializedRequest().ID(), 1)

	t.Log(gas1k, gas2k)
	require.Greater(t, gas2k, gas1k)
}

func incrementSCCounter(t *testing.T, ch *solo.Chain) isc.RequestID {
	tx, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(inccounter.Contract.Name, inccounter.FuncIncCounter.Name).WithGasBudget(math.MaxUint64),
		nil,
	)
	require.NoError(t, err)
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(t, err)
	return reqs[0].ID()
}

func getEventsForRequest(t *testing.T, chain *solo.Chain, reqID isc.RequestID) []string {
	res, err := chain.CallView(blocklog.Contract.Name, blocklog.ViewGetEventsForRequest.Name,
		blocklog.ParamRequestID, reqID,
	)
	require.NoError(t, err)
	events, err := EventsViewResultToStringArray(res)
	require.NoError(t, err)
	return events
}

func getEventsForBlock(t *testing.T, chain *solo.Chain, blockNumber ...int32) []string {
	var res dict.Dict
	var err error
	if len(blockNumber) > 0 {
		res, err = chain.CallView(blocklog.Contract.Name, blocklog.ViewGetEventsForBlock.Name,
			blocklog.ParamBlockIndex, blockNumber[0],
		)
	} else {
		res, err = chain.CallView(blocklog.Contract.Name, blocklog.ViewGetEventsForBlock.Name)
	}
	require.NoError(t, err)
	events, err := EventsViewResultToStringArray(res)
	require.NoError(t, err)
	return events
}

func getEventsForSC(t *testing.T, chain *solo.Chain, fromBlock, toBlock int32) []string {
	res, err := chain.CallView(blocklog.Contract.Name, blocklog.ViewGetEventsForContract.Name,
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
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(inccounter.Processor)
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(10_000, nil)
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
