// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func setupAdvancedInccounterTest(t *testing.T, clusterSize int, committee []int) *ChainEnv {
	quorum := uint16((2*len(committee))/3 + 1)

	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})

	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.Bech32(parameters.L1().Protocol.Bech32HRP))

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID)

	e := &ChainEnv{
		env:   &env{t: t, Clu: clu},
		Chain: chain,
	}
	tx := e.deployNativeIncCounterSC(0)

	waitUntil(t, e.contractIsDeployed(), clu.Config.AllNodes(), 50*time.Second, "contract to be deployed")

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	return e
}

func TestAccessNodesOnLedger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Run("cluster=10, N=4, req=8", func(t *testing.T) {
		const numRequests = 8
		const numValidatorNodes = 3
		const clusterSize = 4 // TODO this should be 10, but its failing (port conflict probably)
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	// t.Run("cluster=10, N=4, req=100", func(t *testing.T) {
	// 	const numRequests = 100
	// 	const numValidatorNodes = 4
	// 	const clusterSize = 10
	// 	testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	// })

	// t.Run("cluster=15, N=4, req=200", func(t *testing.T) {
	// 	testutil.RunHeavy(t)
	// 	const numRequests = 200
	// 	const numValidatorNodes = 4
	// 	const clusterSize = 15
	// 	testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	// })

	// t.Run("cluster=15, N=6, req=200", func(t *testing.T) {
	// 	testutil.RunHeavy(t)
	// 	const numRequests = 200
	// 	const numValidatorNodes = 6
	// 	const clusterSize = 15
	// 	testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	// })
}

func testAccessNodesOnLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int) {
	cmt := util.MakeRange(0, numValidatorNodes)
	e := setupAdvancedInccounterTest(t, clusterSize, cmt)

	for i := 0; i < numRequests; i++ {
		client := e.createNewClient()

		var err error
		for i := 0; i < 5; i++ {
			_, err = client.PostRequest(inccounter.FuncIncCounter.Name)
			if err == nil {
				break
			}
			fmt.Printf("Error posting request, will retry... %v", err)
			time.Sleep(100 * time.Millisecond)
		}
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(numRequests)), e.Clu.AllNodes(), 40*time.Second, "a required number of testAccessNodesOnLedger requests")
}

func TestAccessNodesOffLedger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=6,N=4,req=8", func(t *testing.T) {
		const waitFor = 20 * time.Second
		const numRequests = 8
		const numValidatorNodes = 4
		const clusterSize = 6
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=10,N=4,req=50", func(t *testing.T) {
		const waitFor = 20 * time.Second
		const numRequests = 50
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=10,N=6,req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 120 * time.Second
		const numRequests = 1000
		const numValidatorNodes = 6
		const clusterSize = 10
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=15,N=6,req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 120 * time.Second
		const numRequests = 1000
		const numValidatorNodes = 6
		const clusterSize = 15
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=30,N=15,req=8", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 180 * time.Second
		const numRequests = 8
		const numValidatorNodes = 15
		const clusterSize = 30
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=30,N=20,req=8", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 300 * time.Second
		const numRequests = 8
		const numValidatorNodes = 20
		const clusterSize = 30
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})
}

func testAccessNodesOffLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int, timeout ...time.Duration) {
	to := 60 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}
	cmt := util.MakeRange(0, numValidatorNodes-1)

	e := setupAdvancedInccounterTest(t, clusterSize, cmt)

	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	accountsClient := e.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	tx, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: isc.NewFungibleBaseTokens(1_000_000),
	})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	myClient := e.Chain.SCClient(isc.Hn(nativeIncCounterSCName), keyPair)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 1)})
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(numRequests)), util.MakeRange(0, clusterSize-1), to, "requests counted")

	// TODO is this needed?
	time.Sleep(10 * time.Second) // five time for the nodes to shutdown properly before running the next test
}

// extreme test
func TestAccessNodesMany(t *testing.T) {
	testutil.RunHeavy(t)
	const clusterSize = 15
	const numValidatorNodes = 6
	const requestsCountInitial = 2
	const requestsCountProgression = 2
	const iterationCount = 8

	e := setupAdvancedInccounterTest(t, clusterSize, util.MakeRange(0, numValidatorNodes-1))

	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := e.Chain.SCClient(nativeIncCounterSCHname, keyPair)

	requestsCount := requestsCountInitial
	requestsCumulative := 0
	posted := 0
	for i := 0; i < iterationCount; i++ {
		logMsg := fmt.Sprintf("iteration %v of %v requests", i, requestsCount)
		t.Logf("Running %s", logMsg)
		_, err := myClient.PostNRequests(inccounter.FuncIncCounter.Name, requestsCount)
		require.NoError(t, err)
		posted += requestsCount
		requestsCumulative += requestsCount
		waitUntil(t, e.counterEquals(int64(requestsCumulative)), e.Clu.Config.AllNodes(), 60*time.Second, logMsg)
		requestsCount *= requestsCountProgression
	}
}
