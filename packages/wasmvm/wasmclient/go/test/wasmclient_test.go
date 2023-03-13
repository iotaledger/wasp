// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/tools/cluster/templates"
	clustertests "github.com/iotaledger/wasp/tools/cluster/tests"
)

const (
	myChainID = "atoi1pqtg32l9m53m0uv69636ch474xft5uzkn54er85hc05333hvfxfj6gm6lpx"
	mySeed    = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3"
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

func setupClient(t *testing.T) *wasmclient.WasmClientContext {
	// note that testing the WasmClient code requires a running cluster
	// with a preloaded chain that contains the TestWasmLib demo contract
	// therefore we skip all these tests when in the GitHub repo
	// to run these tests, set up the chain, update myChainID, and uncomment the next line
	t.SkipNow()

	svc := wasmclient.NewWasmClientService("http://localhost:19090", myChainID)
	ctx := wasmclient.NewWasmClientContext(svc, testwasmlib.ScName)
	require.NoError(t, ctx.Err)
	seed := cryptolib.NewSeedFromBytes(wasmtypes.BytesFromString(mySeed))
	wallet := cryptolib.NewKeyPairFromSeed(seed.SubSeed(0))
	ctx.SignRequests(wallet)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestSetup(t *testing.T) {
	ctx := setupClient(t)
	require.NoError(t, ctx.Err)
}

func TestRandom(t *testing.T) {
	ctx := setupClient(t)

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
}

func TestDeploy(t *testing.T) {
	t.SkipNow()
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

	time.Sleep(time.Hour)
}
