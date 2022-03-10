package test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlibclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient"
	coreaccountsclient "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
)

func seedToAddress(mySeed string, index uint64) wasmtypes.ScAddress {
	seedBytes, err := base58.Decode(mySeed)
	if err != nil {
		panic(err)
	}
	address := seed.NewSeed(seedBytes).Address(index)
	return wasmtypes.AddressFromBytes(address.AddressBytes[:])
}

func setupClientV2(t *testing.T) *testwasmlibclient.TestWasmLibService {
	// for now skip client tests
	t.SkipNow()

	require.True(t, wasmclient.SeedIsValid(mySeed))
	require.True(t, wasmclient.ChainIsValid(myChainID))

	// we're testing against wasp-cluster, so defaults will do
	svcClient := wasmclient.DefaultServiceClient()

	// create the service for the testwasmlib smart contract
	svc, err := testwasmlibclient.NewTestWasmLibService(svcClient, myChainID)
	require.NoError(t, err)

	// we'll use the first address in the seed to sign requests
	svc.SignRequests(wasmclient.SeedToKeyPair(mySeed, 0))
	return svc
}

func TestClientV2Events(t *testing.T) {
	svc := setupClientV2(t)
	events := svc.NewEventHandler()
	events.OnTestWasmLibTest(func(e *testwasmlibclient.EventTest) {
		fmt.Printf("Name is %s\n", e.Name)
	})
	svc.Register(events)

	address0 := seedToAddress(mySeed, 0)
	address1 := seedToAddress(mySeed, 1)

	// get new triggerEvent interface, pass params, and post the request
	f := testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue("Lala")
	f.Params.Address().SetValue(address0)
	f.Func.Post()
	require.NoError(t, svc.Err)

	err := svc.WaitRequest(svc.Req)
	require.NoError(t, err)

	// get new triggerEvent interface, pass params, and post the request
	f = testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue("Trala")
	f.Params.Address().SetValue(address1)
	f.Func.Post()
	require.NoError(t, svc.Err)

	err = svc.WaitRequest(svc.Req)
	require.NoError(t, err)
}

func TestClientV2Random(t *testing.T) {
	svc := setupClientV2(t)

	// generate new random value
	f := testwasmlib.ScFuncs.Random(svc)
	f.Func.Post()
	require.NoError(t, svc.Err)

	err := svc.WaitRequest(svc.Req)
	require.NoError(t, err)

	// get current random value
	v := testwasmlib.ScFuncs.GetRandom(svc)
	v.Func.Call()
	require.NoError(t, svc.Err)
	rnd := v.Results.Random().Value()
	require.GreaterOrEqual(t, rnd, uint64(0))
	fmt.Println("Random: ", rnd)
}

func TestClientV2Array(t *testing.T) {
	svc := setupClientV2(t)

	v := testwasmlib.ScFuncs.ArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 0, v.Results.Length().Value())

	f := testwasmlib.ScFuncs.ArrayAppend(svc)
	f.Params.Name().SetValue("Bands")
	f.Params.Value().SetValue("Dire Straits")
	f.Func.Post()
	require.NoError(t, svc.Err)
	err := svc.WaitRequest(svc.Req)
	require.NoError(t, err)

	v = testwasmlib.ScFuncs.ArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 1, v.Results.Length().Value())

	c := testwasmlib.ScFuncs.ArrayClear(svc)
	c.Params.Name().SetValue("Bands")
	c.Func.Post()
	require.NoError(t, svc.Err)
	err = svc.WaitRequest(svc.Req)
	require.NoError(t, err)

	v = testwasmlib.ScFuncs.ArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 0, v.Results.Length().Value())
}

func TestClientV2AccountBalance(t *testing.T) {
	// note: this calls core accounts contract instead of testwasmlib

	// for now skip client tests
	t.SkipNow()

	// we're testing against wasp-cluster, so defaults will do
	svcClient := wasmclient.DefaultServiceClient()

	// create the service for the testwasmlib smart contract
	svc, err := coreaccountsclient.NewCoreAccountsService(svcClient, myChainID)
	require.NoError(t, err)

	// we'll use the first address in the seed to sign requests
	svc.SignRequests(wasmclient.SeedToKeyPair(mySeed, 0))

	bal := coreaccounts.ScFuncs.Balance(svc)
	address0 := seedToAddress(mySeed, 0)
	bal.Params.AgentID().SetValue(address0.AsAgentID())
	bal.Func.Call()
	require.NoError(t, svc.Err)
	balances := bal.Results.Balances()
	fmt.Printf("Balance: %v\n", balances.GetInt64(wasmtypes.IOTA))
}
