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
	"github.com/iotaledger/wasp/clients/chainclient"
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
	useCluster    = false
	useDisposable = false
)

var params = []string{
	"Lala",
	"Trala",
	"Bar|Bar",
	"Bar~|~Bar",
	"Tilde~Tilde",
	"Tilde~~ Bar~/ Space~_",
}

type EventProcessor struct {
	name string
}

func (proc *EventProcessor) sendClientEventsParam(t *testing.T, ctx *wasmclient.WasmClientContext, name string) {
	f := testwasmlib.ScFuncs.TriggerEvent(ctx)
	f.Params.Name().SetValue(name)
	f.Params.Address().SetValue(ctx.CurrentChainID().Address())
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func (proc *EventProcessor) waitClientEventsParam(t *testing.T, ctx *wasmclient.WasmClientContext, name string) {
	ctx.WaitEvent()
	require.NoError(t, ctx.Err)

	require.EqualValues(t, name, proc.name)
}

func setupClient(t *testing.T) *wasmclient.WasmClientContext {
	wasmclient.HrpForClient = ""
	if useCluster {
		return setupClientCluster(t)
	}

	if useDisposable {
		return setupClientDisposable(t)
	}

	// fall back on rudimentary basic testing by using SoloClientService
	return setupClientSolo(t)
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

	cfgWasp := config["wasp"].(map[string]interface{})
	cfgWasp0 := cfgWasp["0"].(string)

	// we'll use the seed keypair to sign requests
	seedBytes, err := iotago.DecodeHex(cfgSeed)
	require.NoError(t, err)

	seed := cryptolib.NewSeedFromBytes(seedBytes)
	wallet := cryptolib.NewKeyPairFromSeed(seed.SubSeed(0))

	// we're testing against disposable wasp-cluster, so defaults will do
	service := wasmclient.DefaultWasmClientService()
	if cfgWasp0[len(cfgWasp0)-6:] != ":19090" {
		// test against Docker container, make sure to pass the correct args to test (top of file)
		service = wasmclient.NewWasmClientService("http://localhost:9090", "127.0.0.1:5550")
	}
	return newClient(t, service, chain, wallet)
}

func setupClientSolo(t solo.TestContext) *wasmclient.WasmClientContext {
	ctx := wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlibimpl.OnDispatch)
	chain := ctx.Chain.ChainID.String()
	wallet := ctx.Chain.OriginatorPrivateKey

	// use Solo as fake Wasp cluster
	return newClient(t, wasmsolo.NewSoloClientService(ctx), chain, wallet)
}

func newClient(t solo.TestContext, svcClient wasmclient.IClientService, chain string, wallet *cryptolib.KeyPair) *wasmclient.WasmClientContext {
	ctx := wasmclient.NewWasmClientContext(svcClient, chain, testwasmlib.ScName)
	require.NoError(t, ctx.Err)
	ctx.SignRequests(wallet)
	return ctx
}

func TestClientAccountBalance(t *testing.T) {
	ctx := setupClient(t)
	wallet := ctx.CurrentKeyPair()

	// note: this calls core accounts contract instead of testwasmlib
	ctx = wasmclient.NewWasmClientContext(ctx.CurrentSvcClient(), ctx.CurrentChainID().String(), coreaccounts.ScName)
	ctx.SignRequests(wallet)

	addr := isc.NewAgentID(wallet.Address())
	agent := wasmtypes.AgentIDFromBytes(addr.Bytes())

	bal := coreaccounts.ScFuncs.BalanceBaseToken(ctx)
	bal.Params.AgentID().SetValue(agent)
	bal.Func.Call()
	require.NoError(t, ctx.Err)
	balance := bal.Results.Balance()
	fmt.Printf("Balance: %d\n", balance.Value())
}

func TestClientArray(t *testing.T) {
	ctx := setupClient(t)

	c := testwasmlib.ScFuncs.StringMapOfStringArrayClear(ctx)
	c.Params.Name().SetValue("Bands")
	c.Func.Post()
	require.NoError(t, ctx.Err)
	ctx.WaitRequest()
	require.NoError(t, ctx.Err)

	v := testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 0, v.Results.Length().Value())

	f := testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	f.Params.Name().SetValue("Bands")
	f.Params.Value().SetValue("Dire Straits")
	f.Func.Post()
	require.NoError(t, ctx.Err)
	ctx.WaitRequest()
	require.NoError(t, ctx.Err)

	v = testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 1, v.Results.Length().Value())

	c = testwasmlib.ScFuncs.StringMapOfStringArrayClear(ctx)
	c.Params.Name().SetValue("Bands")
	c.Func.Post()
	require.NoError(t, ctx.Err)
	ctx.WaitRequest()
	require.NoError(t, ctx.Err)

	v = testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	v.Params.Name().SetValue("Bands")
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 0, v.Results.Length().Value())
}

func TestClientRandom(t *testing.T) {
	ctx := setupClient(t)
	doit := func() {
		// generate new random value
		f := testwasmlib.ScFuncs.Random(ctx)
		f.Func.Post()
		require.NoError(t, ctx.Err)

		ctx.WaitRequest()
		require.NoError(t, ctx.Err)

		// get current random value
		v := testwasmlib.ScFuncs.GetRandom(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
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
	ctx := setupClient(t)

	events := &testwasmlib.TestWasmLibEventHandlers{}
	proc := new(EventProcessor)
	events.OnTestWasmLibTest(func(e *testwasmlib.EventTest) {
		proc.name = e.Name
	})
	ctx.Register(events)

	for _, param := range params {
		proc.sendClientEventsParam(t, ctx, param)
		proc.waitClientEventsParam(t, ctx, param)
	}

	//for _, param := range params {
	//	proc.sendClientEventsParam(param)
	//	ctx.WaitRequest()
	//	require.NoError(t, ctx.Err)
	//}
	//
	//for _, param := range params {
	//	proc.waitClientEventsParam(param)
	//}
}
