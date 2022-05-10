package tests

import (
	"os"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

type env struct {
	t   *testing.T
	Clu *cluster.Cluster
}

type ChainEnv struct {
	*env
	Chain       *cluster.Chain
	scOwner     *cryptolib.KeyPair
	scOwnerAddr iotago.Address
}

func newChainEnv(t *testing.T, clu *cluster.Cluster, chain *cluster.Chain) *ChainEnv {
	keyPair, addr, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	return &ChainEnv{
		env:         &env{t: t, Clu: clu},
		Chain:       chain,
		scOwner:     keyPair,
		scOwnerAddr: addr,
	}
}

type contractEnv struct {
	*ChainEnv
	programHash hashing.HashValue
}

type contractWithMessageCounterEnv struct {
	*contractEnv
	counter *cluster.MessageCounter
}

func (e *ChainEnv) deployContract(wasmName, scDescription string, initParams map[string]interface{}) *contractEnv {
	ret := &contractEnv{ChainEnv: e}

	wasmPath := "wasm/" + wasmName + "_bg.wasm"

	wasm, err := os.ReadFile(wasmPath)
	require.NoError(e.t, err)
	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, e.Chain.OriginatorKeyPair)

	reqTx, err := chClient.DepositFunds(1000000)
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, 30*time.Second)
	require.NoError(e.t, err)

	ph, err := e.Chain.DeployWasmContract(wasmName, scDescription, wasm, initParams)
	require.NoError(e.t, err)
	ret.programHash = ph
	e.t.Logf("deployContract: proghash = %s\n", ph.String())
	return ret
}

func (e *ChainEnv) createNewClient() *scclient.SCClient {
	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	client := e.Chain.SCClient(iscp.Hn(incCounterSCName), keyPair)
	return client
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
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 60*time.Second)
	require.NoError(e.t, err)
	if !e.counter.WaitUntilExpectationsMet() {
		e.t.Fatal()
	}
}

func setupWithNoChain(t *testing.T, opt ...waspClusterOpts) *env {
	return &env{t: t, Clu: newCluster(t, opt...)}
}

func SetupWithChain(t *testing.T, opt ...waspClusterOpts) *ChainEnv {
	e := setupWithNoChain(t, opt...)
	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)
	return newChainEnv(e.t, e.Clu, chain)
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

	return &contractWithMessageCounterEnv{contractEnv: cEnv, counter: counter}
}

func (e *ChainEnv) chainClient() *chainclient.Client {
	return chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, e.scOwner)
}
