package test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/mr-tron/base58"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const useSoloClient = true

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

	// TODO cannot run using time as nonce
	viper.SetConfigFile("wasp-cli.json")
	err := viper.ReadInConfig()
	require.NoError(t, err)

	chain := viper.GetString("chains." + viper.GetString("chain"))
	chainBytes, err := hex.DecodeString(chain)
	require.NoError(t, err)
	chainID := wasmtypes.ChainIDFromBytes(chainBytes)

	// we're testing against wasp-cluster, so defaults will do
	svcClient := wasmclient.DefaultWasmClientService()

	// create the service for the testwasmlib smart contract
	svc := wasmclient.NewWasmClientContext(svcClient, &chainID, testwasmlib.ScName)
	require.NoError(t, svc.Err)

	// we'll use the seed keypair to sign requests
	seedBytes, err := base58.Decode(viper.GetString("wallet.seed"))
	require.NoError(t, err)
	svc.SignRequests(cryptolib.NewKeyPairFromSeed(cryptolib.NewSeedFromBytes(seedBytes)))
	return svc
}

func TestClientEvents(t *testing.T) {
	svc := setupClient(t)
	events := &testwasmlib.TestWasmLibEventHandlers{}
	events.OnTestWasmLibTest(func(e *testwasmlib.EventTest) {
		fmt.Printf("Name is %s\n", e.Name)
	})
	svc.Register(events)

	// get new triggerEvent interface, pass params, and post the request
	f := testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue("Lala")
	f.Params.Address().SetValue(svc.ChainID().Address())
	f.Func.Post()
	require.NoError(t, svc.Err)

	err := svc.WaitRequest()
	require.NoError(t, err)

	// get new triggerEvent interface, pass params, and post the request
	f = testwasmlib.ScFuncs.TriggerEvent(svc)
	f.Params.Name().SetValue("Trala")
	f.Params.Address().SetValue(svc.ChainID().Address())
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
