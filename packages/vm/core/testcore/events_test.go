package testcore

import (
	"math"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

var (
	manyEventsContractName = "ManyEventsContract"
	manyEventsContract     = coreutil.NewContract(manyEventsContractName)

	funcManyEvents = manyEventsContract.Func("manyevents")
	funcBigEvent   = manyEventsContract.Func("bigevent")

	manyEventsContractProcessor = manyEventsContract.Processor(nil,
		funcManyEvents.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			n := codec.Uint32.MustDecode(ctx.Params().Get("n"))
			for i := uint32(0); i < n; i++ {
				ctx.Event("event.test", codec.Uint32.Encode(n))
			}
			return nil
		}),
		funcBigEvent.WithHandler(func(ctx isc.Sandbox) dict.Dict {
			n := codec.Uint32.MustDecode(ctx.Params().Get("n"))
			ctx.Event("event.big", make([]byte, n))
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
	receipt, _ := ch.GetRequestReceipt(reqs[0].ID())
	return receipt.GasBurned, nil
}

func TestManyEvents(t *testing.T) {
	ch := setupTest(t)

	postEvents := func(n uint32) (uint64, error) {
		// post a request that issues too many events (nEvents)
		tx, _, err := ch.PostRequestSyncTx(
			solo.NewCallParamsEx(manyEventsContract.Name, funcManyEvents.Name, "n", n).
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
			solo.NewCallParamsEx(manyEventsContract.Name, funcBigEvent.Name, "n", n).
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
		solo.NewCallParams(inccounter.FuncIncCounter.Message(nil)).WithGasBudget(math.MaxUint64),
		nil,
	)
	require.NoError(t, err)
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(t, err)
	return reqs[0].ID()
}

func getEventsForRequest(t *testing.T, chain *solo.Chain, reqID isc.RequestID) []*isc.Event {
	res, err := chain.CallView(blocklog.ViewGetEventsForRequest.Message(reqID))
	require.NoError(t, err)
	return lo.Must(blocklog.ViewGetEventsForRequest.Output.Decode(res))
}

func getEventsForBlock(t *testing.T, chain *solo.Chain, blockNumber ...uint32) []*isc.Event {
	res, err := chain.CallView(blocklog.ViewGetEventsForBlock.Message(coreutil.Optional(blockNumber...)))
	require.NoError(t, err)
	return lo.Must(blocklog.ViewGetEventsForRequest.Output.Decode(res))
}

func getEventsForSC(t *testing.T, chain *solo.Chain, fromBlock, toBlock uint32) []*isc.Event {
	res, err := chain.CallView(blocklog.ViewGetEventsForContract.Message(blocklog.EventsForContractQuery{
		Contract:   inccounter.Contract.Hname(),
		BlockRange: &blocklog.BlockRange{From: fromBlock, To: toBlock},
	}))
	require.NoError(t, err)
	return lo.Must(blocklog.ViewGetEventsForContract.Output.Decode(res))
}

func TestGetEvents(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(inccounter.Processor)
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(10_000, nil)
	require.NoError(t, err)

	err = ch.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash, nil)
	require.NoError(t, err)

	// block 1 = ch init
	// block 2 = inccounter contract deployment
	reqID1 := incrementSCCounter(t, ch) // #block 3
	reqID2 := incrementSCCounter(t, ch) // #block 4
	reqID3 := incrementSCCounter(t, ch) // #block 5

	testdbhash.VerifyContractStateHash(env, blocklog.Contract, "", t.Name())

	events := getEventsForRequest(t, ch, reqID1)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 1)

	events = getEventsForRequest(t, ch, reqID2)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 2)

	events = getEventsForRequest(t, ch, reqID3)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 3)

	events = getEventsForBlock(t, ch, 3)
	require.Len(t, events, 2)
	checkEventCounter(t, events[0], 0)

	events = getEventsForBlock(t, ch, 4)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 1)

	events = getEventsForBlock(t, ch, 5)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 2)

	events = getEventsForBlock(t, ch)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 3)

	events = getEventsForSC(t, ch, 0, 1000)
	require.Len(t, events, 4)
	checkEventCounter(t, events[0], 0)
	checkEventCounter(t, events[1], 1)
	checkEventCounter(t, events[2], 2)
	checkEventCounter(t, events[3], 3)

	events = getEventsForSC(t, ch, 2, 3)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 0)
}

func checkEventCounter(t *testing.T, event *isc.Event, value uint64) {
	counter, err := codec.Uint64.Decode(event.Payload)
	require.NoError(t, err)
	require.EqualValues(t, counter, value)
}
