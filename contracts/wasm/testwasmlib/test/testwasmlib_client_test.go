package test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlibimpl"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/iotaledger/wasp/tools/cluster/templates"
	clustertests "github.com/iotaledger/wasp/tools/cluster/tests"
)

const (
	useDocker     = true
	useDisposable = true
	useSoloClient = false
)

// to run with docker, set useDisposable to true and run with the following parameters:
// -layer1-api="http://localhost:14265" -layer1-faucet="http://localhost:8091"

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
	templates.WaspConfig = strings.ReplaceAll(templates.WaspConfig, "rocksdb", "mapdb")
	e := clustertests.SetupWithChain(t)
	templates.WaspConfig = strings.ReplaceAll(templates.WaspConfig, "mapdb", "rocksdb")
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
	wasm, err := os.ReadFile("../rs/testwasmlibwasm/pkg/testwasmlibwasm_bg.wasm")
	require.NoError(t, err)

	_, err = e.Chain.DeployWasmContract("testwasmlib", "Test WasmLib", wasm, nil)
	require.NoError(t, err)

	// we're testing against wasp-cluster, so defaults will do
	chainID := e.Chain.ChainID.String()
	return newClient(t, wasmclient.DefaultWasmClientService(), chainID, wallet)
}

func setupClientDisposable(t solo.TestContext) *wasmclient.WasmClientContext {
	// load config file
	configBytes, err := os.ReadFile("wasp-cli.json")
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(configBytes, &config)
	require.NoError(t, err)

	cfgChain := config["chain"].(string)
	cfgChains := config["chains"].(map[string]interface{})
	chain := cfgChains[cfgChain].(string)

	cfgWallet := config["wallet"].(map[string]interface{})
	cfgSeed := cfgWallet["seed"].(string)

	// we'll use the seed keypair to sign requests
	seedBytes, err := iotago.DecodeHex(cfgSeed)
	require.NoError(t, err)

	seed := cryptolib.NewSeedFromBytes(seedBytes)
	wallet := cryptolib.NewKeyPairFromSeed(seed.SubSeed(0))

	// we're testing against disposable wasp-cluster, so defaults will do
	service := wasmclient.DefaultWasmClientService()
	if useDocker {
		// test against Docker container, make sure to pass the correct args to test (top of file)
		service = wasmclient.NewWasmClientService("127.0.0.1:9090", "127.0.0.1:5550")
	}
	return newClient(t, service, chain, wallet)
}

func setupClientSolo(t solo.TestContext) *wasmclient.WasmClientContext {
	ctx := wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlibimpl.OnDispatch)
	chain := ctx.CurrentChainID().String()
	wallet := ctx.Chain.OriginatorPrivateKey

	// use Solo as fake Wasp cluster
	return newClient(t, wasmsolo.NewSoloClientService(ctx), chain, wallet)
}

func newClient(t solo.TestContext, svcClient wasmclient.IClientService, chain string, wallet *cryptolib.KeyPair) *wasmclient.WasmClientContext {
	svc := wasmclient.NewWasmClientContext(svcClient, chain, testwasmlib.ScName)
	require.NoError(t, svc.Err)
	svc.SignRequests(wallet)
	return svc
}

func TestClientAccountBalance(t *testing.T) {
	svc := setupClient(t)
	wallet := svc.CurrentKeyPair()

	// note: this calls core accounts contract instead of testwasmlib
	svc = wasmclient.NewWasmClientContext(svc.CurrentSvcClient(), svc.CurrentChainID().String(), coreaccounts.ScName)
	svc.SignRequests(wallet)

	addr := isc.NewAgentID(wallet.Address())
	agent := wasmtypes.AgentIDFromBytes(addr.Bytes())

	bal := coreaccounts.ScFuncs.BalanceBaseToken(svc)
	bal.Params.AgentID().SetValue(agent)
	bal.Func.Call()
	require.NoError(t, svc.Err)
	balance := bal.Results.Balance()
	fmt.Printf("Balance: %d\n", balance.Value())
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
	svc.WaitRequest()
	require.NoError(t, svc.Err)

	v = testwasmlib.ScFuncs.StringMapOfStringArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 1, v.Results.Length().Value())

	c := testwasmlib.ScFuncs.StringMapOfStringArrayClear(svc)
	c.Params.Name().SetValue("Bands")
	c.Func.Post()
	require.NoError(t, svc.Err)
	svc.WaitRequest()
	require.NoError(t, svc.Err)

	v = testwasmlib.ScFuncs.StringMapOfStringArrayLength(svc)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, svc.Err)
	require.EqualValues(t, 0, v.Results.Length().Value())
}

func TestClientRandom(t *testing.T) {
	svc := setupClient(t)
	doit := func() {
		// generate new random value
		f := testwasmlib.ScFuncs.Random(svc)
		f.Func.Post()
		require.NoError(t, svc.Err)

		svc.WaitRequest()
		require.NoError(t, svc.Err)

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

func TestClientEvents(t *testing.T) {
	svc := setupClient(t)
	events := &testwasmlib.TestWasmLibEventHandlers{}
	name := ""
	events.OnTestWasmLibTest(func(e *testwasmlib.EventTest) {
		name = e.Name
	})
	svc.Register(events)

	event := func() string {
		return name
	}

	testClientEventsParam(t, svc, "Lala", event)
	testClientEventsParam(t, svc, "Trala", event)
	testClientEventsParam(t, svc, "Bar|Bar", event)
	testClientEventsParam(t, svc, "Bar~|~Bar", event)
	testClientEventsParam(t, svc, "Tilde~Tilde", event)
	testClientEventsParam(t, svc, "Tilde~~ Bar~/ Space~_", event)
}

func testClientEventsParam(t *testing.T, svc *wasmclient.WasmClientContext, name string, event func() string) {
	f := testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue(name)
	f.Params.Address().SetValue(svc.CurrentChainID().Address())
	f.Func.Post()
	require.NoError(t, svc.Err)

	svc.WaitRequest()
	require.NoError(t, svc.Err)

	svc.WaitEvent()
	require.NoError(t, svc.Err)

	require.EqualValues(t, name, event())
}
