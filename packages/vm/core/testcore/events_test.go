// excluded temporarily because of compilation errors

package testcore

import (
	"math"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/manyevents"
)

func setupTest(t *testing.T) *solo.Chain {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain()

	// allow "infinite" gas per request
	limits := ch.GetGasLimits()
	limits.MaxGasPerBlock = math.MaxUint64
	limits.MaxGasPerRequest = math.MaxUint64
	ch.SetGasLimits(ch.ChainAdmin, limits)

	// set gas very cheap
	fp := ch.GetGasFeePolicy()
	fp.GasPerToken.A = 1000000
	ch.SetGasFeePolicy(ch.ChainAdmin, fp)

	ch.MustDepositBaseTokensToL2(10_000_000, nil)
	return ch
}

func checkNEvents(t *testing.T, ch *solo.Chain, reqid isc.RequestID, n int) {
	events, err := ch.GetEventsForRequest(reqid)
	require.NoError(t, err)
	require.Len(t, events, n)
}

func getBurnedGas(ch *solo.Chain, reqID isc.RequestID, err error) (uint64, error) {
	if err != nil {
		return 0, err
	}
	receipt, ok := ch.GetRequestReceipt(reqID)
	require.True(ch.Env.T, ok)
	return receipt.GasBurned, nil
}

func TestManyEvents(t *testing.T) {
	ch := setupTest(t)

	postEvents := func(n uint32) (uint64, error) {
		// post a request that issues too many events (nEvents)
		req, _, _, _, err := ch.PostRequestSyncTx(
			solo.NewCallParamsEx(manyevents.Contract.Name, manyevents.FuncManyEvents.Name, isc.NewCallArguments(
				codec.Encode(n),
			)).
				WithMaxAffordableGasBudget(),
			nil,
		)
		return getBurnedGas(ch, req.ID(), err)
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
		req, _, _, _, err := ch.PostRequestSyncTx(
			solo.NewCallParamsEx(manyevents.Contract.Name, manyevents.FuncBigEvent.Name, isc.NewCallArguments(
				codec.Encode(n),
			)).
				WithMaxAffordableGasBudget(),
			nil,
		)
		return getBurnedGas(ch, req.ID(), err)
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
	req, _, _, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(inccounter.FuncIncCounter.Message(nil)).WithGasBudget(math.MaxUint64),
		nil,
	)
	require.NoError(t, err)
	return req.ID()
}

func getEventsForRequest(t *testing.T, chain *solo.Chain, reqID isc.RequestID) []*isc.Event {
	res, err := chain.CallView(blocklog.ViewGetEventsForRequest.Message(reqID))
	require.NoError(t, err)
	return lo.Must(blocklog.ViewGetEventsForRequest.DecodeOutput(res))
}

func getEventsForBlock(t *testing.T, chain *solo.Chain, blockNumber ...uint32) []*isc.Event {
	res, err := chain.CallView(blocklog.ViewGetEventsForBlock.Message(coreutil.Optional(blockNumber...)))
	require.NoError(t, err)
	_, events := lo.Must2(blocklog.ViewGetEventsForBlock.DecodeOutput(res))
	return events
}

func TestGetEvents(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(10_000, nil)
	require.NoError(t, err)

	bi := ch.LatestBlockIndex()
	reqID1 := incrementSCCounter(t, ch) // #block bi+1
	reqID2 := incrementSCCounter(t, ch) // #block bi+2
	reqID3 := incrementSCCounter(t, ch) // #block bi+3

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

	events = getEventsForBlock(t, ch, bi+1)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 1)

	events = getEventsForBlock(t, ch, bi+2)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 2)

	events = getEventsForBlock(t, ch, bi+3)
	require.Len(t, events, 1)
	checkEventCounter(t, events[0], 3)
}

func checkEventCounter(t *testing.T, event *isc.Event, value uint64) {
	counter, err := codec.Decode[uint64](event.Payload)
	require.NoError(t, err)
	require.EqualValues(t, counter, value)
}
