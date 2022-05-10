package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	cluster_tests "github.com/iotaledger/wasp/tools/cluster/tests"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/require"
)

const (
	useSoloClient = true
	mySeed        = "6C6tRksZDWeDTCzX4Q7R2hbpyFV86cSGLVxdkFKSB3sv"
	mySeedIndex   = 0
)

func setupClient(t *testing.T) *wasmclient.WasmClientContext {
	if useSoloClient {
		*wasmsolo.TsWasm = true
		ctx := wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlib.OnLoad)
		svcClient := wasmsolo.NewSoloClientService(ctx)
		chainID := ctx.ChainID()
		svc := wasmclient.NewWasmClientContext(svcClient, &chainID, testwasmlib.ScName)
		require.NoError(t, svc.Err)

		// we'll use the first address in the seed to sign requests
		svc.SignRequests(ctx.Chain.OriginatorPrivateKey)
		return svc
	}
	// use cluster tool
	e := cluster_tests.SetupWithChain(t)

	// TODO wasmlib shouldn't use base58, just the to/from methods from regular chainID // chainIDStr := e.Chain.ChainID.String()
	chainIDStr := base58.Encode(e.Chain.ChainID[:])

	require.True(t, wasmclient.SeedIsValid(mySeed))
	require.True(t, wasmclient.ChainIsValid(chainIDStr))
	chainID := wasmtypes.ChainIDFromBytes(wasmclient.Base58Decode(chainIDStr))

	// request funds to the wallet that the wasmclient will use
	wallet := wasmclient.SeedToKeyPair(mySeed, mySeedIndex)
	e.Clu.RequestFunds(wallet.Address())

	// deposit funds to the on-chain account
	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, wallet)
	reqTx, err := chClient.DepositFunds(10_000_000)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	// deploy the contract
	wasm, err := os.ReadFile("./testwasmlib_bg.wasm")
	require.NoError(t, err)

	_, err = e.Chain.DeployWasmContract("testwasmlib", "contract to test wasmlib", wasm, nil)
	require.NoError(t, err)

	// we're testing against wasp-cluster, so defaults will do
	svcClient := wasmclient.DefaultWasmClientService()

	// create the service for the testwasmlib smart contract
	svc := wasmclient.NewWasmClientContext(svcClient, &chainID, testwasmlib.ScName)
	require.NoError(t, svc.Err)

	// we'll use the first address in the seed to sign requests
	svc.SignRequests(wasmclient.SeedToKeyPair(mySeed, mySeedIndex))
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
	doit := func() {
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
	doit()
	doit()
	doit()
	doit()
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
//	svcClient := wasmclient.DefaultWasmClientService()
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
