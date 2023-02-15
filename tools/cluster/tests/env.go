package tests

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/scclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/cluster"
)

// TODO remove this?
func setupWithNoChain(t *testing.T, opt ...waspClusterOpts) *ChainEnv {
	clu, _ := newCluster(t, opt...)
	return &ChainEnv{t: t, Clu: clu}
}

type ChainEnv struct {
	t        *testing.T
	Clu      *cluster.Cluster
	dataPath string
	Chain    *cluster.Chain
}

func newChainEnv(t *testing.T, clu *cluster.Cluster, chain *cluster.Chain) *ChainEnv {
	return &ChainEnv{
		t:     t,
		Clu:   clu,
		Chain: chain,
	}
}

type contractEnv struct {
	*ChainEnv
	programHash hashing.HashValue
}

func (e *ChainEnv) deployWasmContract(wasmName, scDescription string, initParams map[string]interface{}) *contractEnv {
	ret := &contractEnv{ChainEnv: e}

	wasmPath := "wasm/" + wasmName + "_bg.wasm"

	wasm, err := os.ReadFile(wasmPath)
	require.NoError(e.t, err)
	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, e.Chain.OriginatorKeyPair)

	reqTx, err := chClient.DepositFunds(1_000_000)
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
	keyPair, addr, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	retries := 0
	for {
		outs, err := e.Clu.L1Client().OutputMap(addr)
		require.NoError(e.t, err)
		if len(outs) > 0 {
			break
		}
		retries++
		if retries > 10 {
			panic("createNewClient - funds aren't available")
		}
		time.Sleep(300 * time.Millisecond)
	}
	return e.Chain.SCClient(isc.Hn(nativeIncCounterSCName), keyPair)
}

func SetupWithChain(t *testing.T, opt ...waspClusterOpts) *ChainEnv {
	e := setupWithNoChain(t, opt...)
	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)
	return newChainEnv(e.t, e.Clu, chain)
}

func (e *ChainEnv) NewChainClient() *chainclient.Client {
	wallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	return chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, wallet)
}

func (e *ChainEnv) DepositFunds(amount uint64, keyPair *cryptolib.KeyPair) {
	accountsClient := e.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	tx, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: isc.NewAssetsBaseTokens(amount),
	})
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(e.t, err)
}
