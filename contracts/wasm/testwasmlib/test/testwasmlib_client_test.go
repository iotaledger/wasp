package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/mr-tron/base58"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	cluster_tests "github.com/iotaledger/wasp/tools/cluster/tests"
)

const (
	useDisposable = false
	useSoloClient = true
)

func setupClient(t *testing.T) *wasmclient.WasmClientContext {
	if useDisposable {
		return setupClientDisposable(t)
	}

	if useSoloClient {
		return setupClientSolo(t)
	}

	return setupClientCluster(t)
}

func setupClientCluster(t *testing.T) *wasmclient.WasmClientContext {
	e := cluster_tests.SetupWithChain(t)
	chainID := wasmtypes.ChainIDFromBytes(e.Chain.ChainID.Bytes())
	wallet := cryptolib.NewKeyPair()

	// request funds to the wallet that the wasmclient will use
	err := e.Clu.RequestFunds(wallet.Address())
	require.NoError(t, err)

	// deposit funds to the on-chain account
	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, wallet)
	reqTx, err := chClient.DepositFunds(10_000_000)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	// deploy the contract
	wasm, err := os.ReadFile("../pkg/testwasmlib_bg.wasm")
	require.NoError(t, err)

	_, err = e.Chain.DeployWasmContract("testwasmlib", "Test WasmLib", wasm, nil)
	require.NoError(t, err)

	// we're testing against wasp-cluster, so defaults will do
	return newClient(t, wasmclient.DefaultWasmClientService(), chainID, wallet)
}

func setupClientDisposable(t *testing.T) *wasmclient.WasmClientContext {
	viper.SetConfigFile("wasp-cli.json")
	err := viper.ReadInConfig()
	require.NoError(t, err)

	chain := viper.GetString("chains." + viper.GetString("chain"))
	chID, err := isc.ChainIDFromString(chain)
	require.NoError(t, err)
	chainID := wasmtypes.ChainIDFromBytes(chID.Bytes())

	// we'll use the seed keypair to sign requests
	seedBytes, err := base58.Decode(viper.GetString("wallet.seed"))
	require.NoError(t, err)
	seed := cryptolib.NewSeedFromBytes(seedBytes)
	wallet := cryptolib.NewKeyPairFromSeed(seed.SubSeed(0))

	// we're testing against disposable wasp-cluster, so defaults will do
	return newClient(t, wasmclient.DefaultWasmClientService(), chainID, wallet)
}

func setupClientSolo(t *testing.T) *wasmclient.WasmClientContext {
	ctx := wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlib.OnLoad)
	chainID := ctx.CurrentChainID()
	wallet := ctx.Chain.OriginatorPrivateKey

	// use Solo as fake Wasp cluster
	return newClient(t, wasmsolo.NewSoloClientService(ctx), chainID, wallet)
}

func newClient(t *testing.T, svcClient wasmclient.IClientService, chainID wasmtypes.ScChainID, wallet *cryptolib.KeyPair) *wasmclient.WasmClientContext {
	svc := wasmclient.NewWasmClientContext(svcClient, chainID, testwasmlib.ScName)
	require.NoError(t, svc.Err)
	svc.SignRequests(wallet)
	return svc
}

func TestClientEvents(t *testing.T) {
	svc := setupClient(t)
	events := &testwasmlib.TestWasmLibEventHandlers{}
	name := ""
	events.OnTestWasmLibTest(func(e *testwasmlib.EventTest) {
		name = e.Name
	})
	svc.Register(events)

	testClientEventsParam(t, svc, "Lala", &name)
	testClientEventsParam(t, svc, "Trala", &name)
	testClientEventsParam(t, svc, "Bar|Bar", &name)
	testClientEventsParam(t, svc, "Bar~|~Bar", &name)
	testClientEventsParam(t, svc, "Tilde~Tilde", &name)
	testClientEventsParam(t, svc, "Tilde~~ Bar~/ Space~_", &name)
}

func testClientEventsParam(t *testing.T, svc *wasmclient.WasmClientContext, param string, name *string) {
	// get new triggerEvent interface, pass params, and post the request
	f := testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue(param)
	f.Params.Address().SetValue(svc.CurrentChainID().Address())
	f.Func.Post()
	require.NoError(t, svc.Err)

	err := svc.WaitRequest()
	require.NoError(t, err)

	// make sure we wait for the event to show up
	err = svc.WaitEvent()
	require.NoError(t, err)

	require.EqualValues(t, param, *name)
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
