package tests

import (
	"math/rand"
	"os"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

var (
	wallet      = initSeed()
	scOwner     = cryptolib.NewKeyPairFromSeed(wallet.SubSeed(0))
	scOwnerAddr = tpkg.RandEd25519Address()
)

type env struct {
	t   *testing.T
	clu *cluster.Cluster
}

type chainEnv struct {
	*env
	chain        *cluster.Chain
	addressIndex uint64
}

func newChainEnv(t *testing.T, clu *cluster.Cluster, chain *cluster.Chain) *chainEnv {
	return &chainEnv{env: &env{t: t, clu: clu}, chain: chain}
}

type contractEnv struct {
	*chainEnv
	programHash hashing.HashValue
}

type contractWithMessageCounterEnv struct {
	*contractEnv
	counter *cluster.MessageCounter
}

func initSeed() cryptolib.Seed {
	return cryptolib.NewSeed()
}

func (e *chainEnv) deployContract(wasmName, scDescription string, initParams map[string]interface{}) *contractEnv {
	ret := &contractEnv{chainEnv: e}

	wasmPath := "wasm/" + wasmName + "_bg.wasm"

	wasm, err := os.ReadFile(wasmPath)
	require.NoError(e.t, err)
	chClient := chainclient.New(e.clu.L1Client(), e.clu.WaspClient(0), e.chain.ChainID, e.chain.OriginatorKeyPair)

	reqTx, err := chClient.DepositFunds(100)
	require.NoError(e.t, err)
	err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, reqTx, 30*time.Second)
	require.NoError(e.t, err)

	ph, err := e.chain.DeployWasmContract(wasmName, scDescription, wasm, initParams)
	require.NoError(e.t, err)
	ret.programHash = ph
	e.t.Logf("deployContract: proghash = %s\n", ph.String())
	return ret
}

func (e *chainEnv) createNewClient() *scclient.SCClient {
	keyPair, _ := e.getOrCreateAddress()
	client := e.chain.SCClient(iscp.Hn(incCounterSCName), keyPair)
	return client
}

func (e *chainEnv) getOrCreateAddress() (*cryptolib.KeyPair, iotago.Address) {
	const minTokenAmountBeforeRequestingNewFunds uint64 = 100

	randomAddress := rand.NewSource(time.Now().UnixNano())

	keyPair := cryptolib.NewKeyPairFromSeed(wallet.SubSeed(e.addressIndex))
	myAddress := keyPair.Address()

	funds := e.clu.AddressBalances(myAddress).Iotas

	if funds <= minTokenAmountBeforeRequestingNewFunds {
		// Requesting new token requires a new address

		e.addressIndex = rand.New(randomAddress).Uint64()
		e.t.Logf("Generating new address: %v", e.addressIndex)

		keyPair = cryptolib.NewKeyPairFromSeed(wallet.SubSeed(e.addressIndex))
		myAddress = keyPair.Address()

		e.requestFunds(myAddress, "myAddress")
		e.t.Logf("Funds: %v, addressIndex: %v", funds, e.addressIndex)
	}

	return keyPair, myAddress
}

func (e *contractWithMessageCounterEnv) postRequest(contract, entryPoint iscp.Hname, tokens int, params map[string]interface{}) {
	transfer := iscp.NewFungibleTokens(uint64(tokens), nil)
	e.postRequestFull(contract, entryPoint, transfer, params)
}

func (e *contractWithMessageCounterEnv) postRequestFull(contract, entryPoint iscp.Hname, transfer *iscp.FungibleTokens, params map[string]interface{}) {
	b := iscp.NewEmptyAssets()
	if transfer != nil {
		b = transfer
	}
	tx, err := e.chainClient().Post1Request(contract, entryPoint, chainclient.PostRequestParams{
		Transfer: b,
		Args:     codec.MakeDict(params),
	})
	require.NoError(e.t, err)
	err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, tx, 60*time.Second)
	require.NoError(e.t, err)
	if !e.counter.WaitUntilExpectationsMet() {
		e.t.Fail()
	}
}

func setupWithNoChain(t *testing.T, opt ...interface{}) *env {
	return &env{t: t, clu: newCluster(t, opt...)}
}

func setupWithChain(t *testing.T, opt ...interface{}) *chainEnv {
	e := setupWithNoChain(t, opt...)
	chain, err := e.clu.DeployDefaultChain()
	require.NoError(t, err)
	return newChainEnv(e.t, e.clu, chain)
}

func setupWithContractAndMessageCounter(t *testing.T, name, description string, nrOfRequests int) *contractWithMessageCounterEnv {
	clu := newCluster(t)

	expectations := map[string]int{
		"dismissed_committee": 0,
		"state":               3 + nrOfRequests,
		//"request_out":         3 + nrOfRequests,    // not always coming from all nodes, but from quorum only
	}

	var err error

	counter, err := clu.StartMessageCounter(expectations)
	require.NoError(t, err)
	t.Cleanup(counter.Close)

	chain, err := clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, clu, chain)

	cEnv := chEnv.deployContract(name, description, nil)
	require.NoError(t, err)

	chEnv.requestFunds(scOwnerAddr, "client")

	return &contractWithMessageCounterEnv{contractEnv: cEnv, counter: counter}
}

func (e *chainEnv) chainClient() *chainclient.Client {
	return chainclient.New(e.clu.L1Client(), e.clu.WaspClient(0), e.chain.ChainID, scOwner)
}
