// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlibimpl"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient/iscclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/iotaledger/wasp/tools/cluster/templates"
	clustertests "github.com/iotaledger/wasp/tools/cluster/tests"
)

const (
	useCluster    = false
	useDisposable = false
	mySeed        = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3"
	waspAPI       = "http://localhost:19090"
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
	for i := 0; i < 100 && proc.name == "" && ctx.Err == nil; i++ {
		time.Sleep(100 * time.Millisecond)
	}
	require.NoError(t, ctx.Err)
	require.EqualValues(t, name, proc.name)
	proc.name = ""
}

func subSeed(seed string, index uint32) *iscclient.Keypair {
	return iscclient.KeyPairFromSubSeed(wasmtypes.BytesFromString(seed), index)
}

func setupClient(t *testing.T) *wasmclient.WasmClientContext {
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
	env := clustertests.SetupWithChain(t)
	templates.WaspConfig = strings.ReplaceAll(templates.WaspConfig, "mapdb", "rocksdb")
	keyPair := subSeed(mySeed, 0)
	pk, _ := cryptolib.PrivateKeyFromBytes(keyPair.GetPrivateKey())
	wallet := cryptolib.KeyPairFromPrivateKey(pk)

	// request funds to the wallet that the wasm client will use
	err := env.Clu.RequestFunds(wallet.Address())
	require.NoError(t, err)

	// deposit funds to the on-chain account
	chClient := chainclient.New(env.Clu.L1Client(), env.Clu.WaspClient(0), env.Chain.ChainID, wallet)
	reqTx, err := chClient.DepositFunds(10_000_000)
	require.NoError(t, err)
	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(t, err)

	// deploy the contract
	wasm, err := os.ReadFile("../rs/testwasmlibwasm/pkg/testwasmlibwasm_bg.wasm")
	require.NoError(t, err)

	_, err = env.Chain.DeployWasmContract("testwasmlib", wasm, nil)
	require.NoError(t, err)

	svc := wasmclient.NewWasmClientService("http://localhost:19090")
	err = svc.SetCurrentChainID(env.Chain.ChainID.String())
	require.NoError(t, err)
	return newClient(t, svc, keyPair)
}

func setupClientDisposable(t testing.TB) *wasmclient.WasmClientContext {
	configBytes, err := os.ReadFile("wasp-cli.json")
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(configBytes, &config)
	require.NoError(t, err)

	cfgChains := config["chains"].(map[string]interface{})
	chainID := cfgChains["mychain"].(string)

	cfgWallet := config["wallet"].(map[string]interface{})
	cfgSeed := cfgWallet["seed"].(string)

	cfgWasp := config["wasp"].(map[string]interface{})
	cfgWaspAPI := cfgWasp["0"].(string)

	// we'll use the seed keypair to sign requests
	keyPair := subSeed(cfgSeed, 0)

	svc := wasmclient.NewWasmClientService(cfgWaspAPI)
	require.True(t, svc.IsHealthy())
	err = svc.SetCurrentChainID(chainID)
	require.NoError(t, err)
	return newClient(t, svc, keyPair)
}

func setupClientSolo(t testing.TB) *wasmclient.WasmClientContext {
	ctx := wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlibimpl.OnDispatch)
	chainID := ctx.Chain.ChainID.String()
	keyPair := iscclient.KeyPairFromSeed(ctx.Chain.OriginatorPrivateKey.GetPrivateKey().AsBytes()[:32])

	// use Solo as fake Wasp cluster
	return newClient(t, wasmsolo.NewSoloClientService(ctx, chainID), keyPair)
}

func newClient(t testing.TB, svcClient wasmclient.IClientService, keyPair *iscclient.Keypair) *wasmclient.WasmClientContext {
	ctx := wasmclient.NewWasmClientContext(svcClient, testwasmlib.ScName)
	require.NoError(t, ctx.Err)
	ctx.SignRequests(keyPair)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestTimedDeactivation(t *testing.T) {
	if !useDisposable && !useCluster {
		t.SkipNow()
	}

	var ctxCluster *wasmclient.WasmClientContext
	if useCluster {
		ctxCluster = setupClient(t)
	}

	ctx := setupClientLib(t)
	require.NoError(t, ctx.Err)

	active := getActive(t, ctx)
	require.False(t, active)

	f := testwasmlib.ScFuncs.Activate(ctx)
	f.Params.Seconds().SetValue(420)
	f.Func.TransferBaseTokens(2_000_000).AllowanceBaseTokens(1_000_000).Post()
	require.NoError(t, ctx.Err)

	ctx.WaitRequest()
	require.NoError(t, ctx.Err)

	for i := 0; i < 100; i++ {
		active = getActive(t, ctx)
		seconds := 20
		fmt.Printf("TICK #%d: %v\n", i*seconds, active)
		if !active {
			break
		}
		factor := time.Duration(seconds)
		if useCluster {
			// time marches 10x faster
			factor /= 10
		}
		time.Sleep(factor * time.Second)
	}

	_ = ctxCluster
}

func getActive(t *testing.T, ctx *wasmclient.WasmClientContext) bool {
	a := testwasmlib.ScFuncs.GetActive(ctx)
	a.Func.Call()
	require.NoError(t, ctx.Err)
	return a.Results.Active().Value()
}

func TestClientAccountBalance(t *testing.T) {
	ctx := setupClient(t)
	keyPair := ctx.CurrentKeyPair()

	// note: this calls core accounts contract instead of testwasmlib
	ctx = wasmclient.NewWasmClientContext(ctx.CurrentSvcClient(), coreaccounts.ScName)
	ctx.SignRequests(keyPair)

	agent := wasmtypes.ScAgentIDFromAddress(keyPair.Address())

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

	for i := 0; i < 4; i++ {
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
		fmt.Println("Random: ", rnd)
		require.GreaterOrEqual(t, rnd, uint64(0))
	}
}

func TestClientEvents(t *testing.T) {
	ctx := setupClient(t)

	events := testwasmlib.NewTestWasmLibEventHandlers()
	proc := new(EventProcessor)
	events.OnTestWasmLibTest(func(e *testwasmlib.EventTest) {
		proc.name = e.Name
	})
	ctx.Register(events)
	require.NoError(t, ctx.Err)

	for _, param := range params {
		proc.sendClientEventsParam(t, ctx, param)
		proc.waitClientEventsParam(t, ctx, param)
	}

	ctx.Unregister(events.ID())
	require.NoError(t, ctx.Err)

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

func setupClientLib(t *testing.T) *wasmclient.WasmClientContext {
	if !useCluster && !useDisposable {
		t.SkipNow()
	}

	svc := wasmclient.NewWasmClientService(waspAPI)

	// note that testing the WasmClient code requires a running wasp-cluster
	// with a single preloaded chain that contains the TestWasmLib demo contract
	// therefore we skip all WasmClient tests when in the GitHub repo
	if !svc.IsHealthy() {
		t.SkipNow()
	}

	err := svc.SetDefaultChainID()
	require.NoError(t, err)

	ctx := wasmclient.NewWasmClientContext(svc, testwasmlib.ScName)
	require.NoError(t, ctx.Err)

	keyPair := subSeed(mySeed, 0)
	ctx.SignRequests(keyPair)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestAPICallView(t *testing.T) {
	ctxCluster := setupClient(t)

	ctx := setupClientLib(t)
	require.NoError(t, ctx.Err)

	rnd := testAPICallView(t, ctx)
	require.Equal(t, rnd, uint64(0))

	_ = ctxCluster
}

func TestAPIPostRequest(t *testing.T) {
	ctxCluster := setupClient(t)

	ctx := setupClientLib(t)

	testAPIPostRequest(t, ctx)

	rnd := testAPICallView(t, ctx)
	require.NotEqual(t, rnd, uint64(0))

	_ = ctxCluster
}

func testAPIPostRequest(t *testing.T, ctx *wasmclient.WasmClientContext) {
	// generate new random value
	f := testwasmlib.ScFuncs.Random(ctx)
	f.Func.Post()
	require.NoError(t, ctx.Err)

	ctx.WaitRequest()
	require.NoError(t, ctx.Err)
}

func testAPICallView(t *testing.T, ctx *wasmclient.WasmClientContext) uint64 {
	v := testwasmlib.ScFuncs.GetRandom(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	rnd := v.Results.Random().Value()
	fmt.Println("Random: ", rnd)
	return rnd
}

func TestAPIErrorHandling(t *testing.T) {
	ctxCluster := setupClient(t)

	ctx := setupClientLib(t)
	require.NoError(t, ctx.Err)

	testAPIErrorHandling(t, ctx)

	_ = ctxCluster
}

func testAPIErrorHandling(t *testing.T, ctx *wasmclient.WasmClientContext) {
	fmt.Println("check missing mandatory string parameter")
	v := testwasmlib.ScFuncs.CheckString(ctx)
	v.Func.Call()
	require.Error(t, ctx.Err)
	fmt.Println("Error: " + ctx.Err.Error())

	// // wait for nonexisting request id (time out)
	// ctx.WaitRequest(wasmtypes.RequestIDFromBytes(nil))
	// require.Error(t, ctx.Err)
	// fmt.Println("Error: " + ctx.Err.Error())

	fmt.Println("check sign with wrong key pair")
	keyPair := subSeed(mySeed, 1)
	ctx.SignRequests(keyPair)
	f := testwasmlib.ScFuncs.Random(ctx)
	f.Func.Post()
	require.Error(t, ctx.Err)
	fmt.Println("Error: " + ctx.Err.Error())

	fmt.Println("check wait for request on wrong chain")
	chainBytes := wasmtypes.ChainIDToBytes(ctx.CurrentChainID())
	chainBytes[2]++
	badChainID := wasmtypes.ChainIDToString(wasmtypes.ChainIDFromBytes(chainBytes))

	svc := wasmclient.NewWasmClientService(waspAPI)
	ctx.Err = svc.SetCurrentChainID(badChainID)
	require.NoError(t, ctx.Err)
	ctx = wasmclient.NewWasmClientContext(svc, testwasmlib.ScName)
	require.NoError(t, ctx.Err)
	ctx.SignRequests(keyPair)
	require.NoError(t, ctx.Err)
	ctx.WaitRequest(wasmtypes.RequestIDFromBytes(nil))
	require.Error(t, ctx.Err)
	fmt.Println("Error: " + ctx.Err.Error())
}

func TestAPIAsyncInvoke(t *testing.T) {
	// t.SkipNow()
	ctxCluster := setupClient(t)

	done1 := make(chan bool)

	// note that we'll need two contexts because one is never going to
	// get ctx.Err set and the other will get ctx.Err set for sure

	ctx := setupClientLib(t)
	require.NoError(t, ctx.Err)
	ctx2 := setupClientLib(t)
	require.NoError(t, ctx2.Err)

	lastRnd := testAPICallView(t, ctx)

	var doneLock sync.Mutex
	doneLock.Lock()
	go func() {
		for !doneLock.TryLock() {
			fmt.Println("CALL")
			rnd := testAPICallView(t, ctx)
			fmt.Println("CALL DONE")
			require.Equal(t, rnd, lastRnd)
		}
		done1 <- true
	}()
	go func() {
		testAPIErrorHandling(t, ctx2)
		doneLock.Unlock()
	}()

	fmt.Println("APIWAIT WAIT")
	<-done1
	fmt.Println("APIWAIT DONE")

	_ = ctxCluster
}

func TestRunAsWaspClusterForRustTesting(t *testing.T) {
	// only enable this test when you know why you want to enable it, because
	// this test will run an endless loop emulating the wasp-cluster tool
	t.SkipNow()
	require.True(t, useCluster)

	// This will start a cluster test with preloaded testwasmlib SC.
	// Essentially this replaces starting wasp-cluster and then
	// running deploy.sh to set up the tests for client lib testing.
	ctxCluster := setupClient(t)

	// part of setting up is a Random request and a GetRandom call

	ctx := setupClientLib(t)

	testAPIPostRequest(t, ctx)

	rnd := testAPICallView(t, ctx)
	require.NotEqual(t, rnd, uint64(0))

	// now wait forever and allow external client lib tests to run
	done1 := make(chan bool)
	<-done1

	_ = ctxCluster
}
