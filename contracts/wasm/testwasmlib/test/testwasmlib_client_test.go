package test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

// hardcoded seed and chain ID, taken from wasp-cli.json
// note that normally the chain has already been set up and
// the contract has already been deployed in some way, so
// these values are usually available from elsewhere
const (
	useSoloClient = true
	myChainID     = "gkdhQDvLi23xxgpiLbmzodcayx3"
	mySeed        = "6C6tRksZDWeDTCzX4Q7R2hbpyFV86cSGLVxdkFKSB3sv"
)

func setupClient(t *testing.T) *wasmclient.WasmClientContext {
	if useSoloClient {
		ctx := wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlib.OnLoad)
		svcClient := wasmsolo.NewSoloClientService(ctx)
		chainID := ctx.ChainID()
		svc := wasmclient.NewWasmClientContext(svcClient, &chainID, testwasmlib.ScName)
		require.NoError(t, svc.Err)

		// we'll use the first address in the seed to sign requests
		svc.SignRequests(ctx.Chain.OriginatorPrivateKey)
		return svc
	}

	require.True(t, wasmclient.SeedIsValid(mySeed))
	require.True(t, wasmclient.ChainIsValid(myChainID))
	chainID := wasmtypes.ChainIDFromBytes(wasmclient.Base58Decode(myChainID))

	// we're testing against wasp-cluster, so defaults will do
	svcClient := wasmclient.DefaultServiceClient()

	// create the service for the testwasmlib smart contract
	svc := wasmclient.NewWasmClientContext(svcClient, &chainID, testwasmlib.ScName)
	require.NoError(t, svc.Err)

	// we'll use the first address in the seed to sign requests
	svc.SignRequests(wasmclient.SeedToKeyPair(mySeed, 0))
	return svc
}

func TestClientEvents(t *testing.T) {
	svc := setupClient(t)
	events := &testwasmlib.TestWasmLibEventHandlers{}
	events.OnTestWasmLibTest(func(e *testwasmlib.EventTest) {
		fmt.Printf("Name is %s\n", e.Name)
	})
	svc.Register(events)

	address0 := wasmclient.SeedToAddress(mySeed, 0)
	address1 := wasmclient.SeedToAddress(mySeed, 1)

	// get new triggerEvent interface, pass params, and post the request
	f := testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue("Lala")
	f.Params.Address().SetValue(address0)
	f.Func.Post()
	require.NoError(t, svc.Err)

	err := svc.WaitRequest()
	require.NoError(t, err)

	// get new triggerEvent interface, pass params, and post the request
	f = testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue("Trala")
	f.Params.Address().SetValue(address1)
	f.Func.Post()
	require.NoError(t, svc.Err)

	err = svc.WaitRequest()
	require.NoError(t, err)
}

func TestClientRandom(t *testing.T) {
	svc := setupClient(t)

	// generate new random value
	f := testwasmlib.ScFuncs.Random(svc)
	f.Func.Post()
	require.NoError(t, svc.Err)

	err := svc.WaitRequest()
	require.NoError(t, err)

	// get current random value
	v := testwasmlib.ScFuncs.GetRandom(svc)
	v.Func.Call()
	require.NoError(t, svc.Err)
	rnd := v.Results.Random().Value()
	require.GreaterOrEqual(t, rnd, uint64(0))
	fmt.Println("Random: ", rnd)
}

func TestClientArray(t *testing.T) {
	svc := setupClient(t)

	v := testwasmlib.ScFuncs.StringMapOfStringArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 0, v.Results.Length().Value())

	f := testwasmlib.ScFuncs.StringMapOfStringArrayAppend(svc)
	f.Params.Name().SetValue("Bands")
	f.Params.Value().SetValue("Dire Straits")
	f.Func.Post()
	require.NoError(t, svc.Err)
	err := svc.WaitRequest()
	require.NoError(t, err)

	v = testwasmlib.ScFuncs.StringMapOfStringArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 1, v.Results.Length().Value())

	c := testwasmlib.ScFuncs.StringMapOfStringArrayClear(svc)
	c.Params.Name().SetValue("Bands")
	c.Func.Post()
	require.NoError(t, svc.Err)
	err = svc.WaitRequest()
	require.NoError(t, err)

	v = testwasmlib.ScFuncs.StringMapOfStringArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 0, v.Results.Length().Value())
}

//func TestClientAccountBalance(t *testing.T) {
//	// note: this calls core accounts contract instead of testwasmlib
//
//	// for now skip client tests
//	t.SkipNow()
//
//	// we're testing against wasp-cluster, so defaults will do
//	svcClient := wasmclient.DefaultServiceClient()
//
//	// create the service for the testwasmlib smart contract
//	svc, err := coreaccountsclient.NewCoreAccountsService(svcClient, myChainID)
//	require.NoError(t, err)
//
//	// we'll use the first address in the seed to sign requests
//	svc.SignRequests(wasmclient.SeedToKeyPair(mySeed, 0))
//
//	bal := coreaccounts.ScFuncs.Balance(svc)
//	agent := wasmclient.SeedToAgentID(mySeed, 0)
//	bal.Params.AgentID().SetValue(agent)
//	bal.Func.Call()
//	require.NoError(t, svc.Err)
//	balances := bal.Results.Balances()
//	fmt.Printf("Balance: %v\n", balances.GetInt64(wasmtypes.IOTA))
//}
