package testcore

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

var (
	manyEventsContractName = "ManyEventsContract"
	manyEventsContract     = coreutil.NewContract(manyEventsContractName, "many events contract")

	funcManyEvents = coreutil.Func("manyevents")

	manyEventsContractProcessor = manyEventsContract.Processor(nil,
		funcManyEvents.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
			for i := 0; i < 300; i++ {
				ctx.Event(fmt.Sprintf("testing many events %d", i))
			}
			return nil, nil
		}),
	)
)

func TestManyEvents(t *testing.T) {
	env := solo.New(t, true, false).WithNativeContract(manyEventsContractProcessor)
	ch := env.NewChain(nil, "ch")
	err := ch.DeployContract(nil, manyEventsContract.Name, manyEventsContract.ProgramHash)
	require.NoError(t, err)

	// post a request to many events which issues too many events
	req := solo.NewCallParams(manyEventsContract.Name, funcManyEvents.Name).WithIotas(1)
	_, reqid, err := ch.RequestFromParamsToLedger(req, ch.OriginatorKeyPair)
	require.NoError(t, err)

	_, err = ch.PostRequestSync(
		req,
		nil,
	)
	require.Error(t, err) // error expected (too many events)

	// fetch events from blocklog
	events, err := ch.GetEventsForContract(manyEventsContractName)
	require.NoError(t, err)
	require.Len(t, events, 0)

	events, err = ch.GetEventsForRequest(reqid)
	require.NoError(t, err)
	require.Len(t, events, 0)
}
