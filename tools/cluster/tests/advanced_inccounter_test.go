// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func setupAdvancedInccounterTest(t *testing.T, clusterSize int, committee []int) *ChainEnv {
	quorum := uint16((2*len(committee))/3 + 1)

	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})

	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.Bech32(parameters.L1.Protocol.Bech32HRP))

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID)

	chEnv := &ChainEnv{
		env:   &env{t: t, Clu: clu},
		Chain: chain,
	}
	tx := chEnv.deployNativeIncCounterSC(0)

	waitUntil(t, chEnv.contractIsDeployed(nativeIncCounterSCName), clu.Config.AllNodes(), 50*time.Second, "contract to be deployed")

	_, err = chEnv.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	return chEnv
}

func (e *ChainEnv) printBlocks(expected int) {
	recs, err := e.Chain.GetAllBlockInfoRecordsReverse()
	require.NoError(e.t, err)

	sum := 0
	for _, rec := range recs {
		e.t.Logf("---- block #%d: total: %d, off-ledger: %d, success: %d", rec.BlockIndex, rec.TotalRequests, rec.NumOffLedgerRequests, rec.NumSuccessfulRequests)
		sum += int(rec.TotalRequests)
	}
	e.t.Logf("Total requests processed: %d", sum)
	require.EqualValues(e.t, expected, sum)
}

func TestAccessNodesOnLedger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Run("cluster=10, N=4, req=8", func(t *testing.T) {
		const numRequests = 8
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	t.Run("cluster=10, N=4, req=100", func(t *testing.T) {
		const numRequests = 100
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	t.Run("cluster=15, N=4, req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const numRequests = 1000
		const numValidatorNodes = 4
		const clusterSize = 15
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	t.Run("cluster=15, N=6, req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const numRequests = 1000
		const numValidatorNodes = 6
		const clusterSize = 15
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})
}

func testAccessNodesOnLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int) {
	cmt := util.MakeRange(0, numValidatorNodes)
	e := setupAdvancedInccounterTest(t, clusterSize, cmt)

	for i := 0; i < numRequests; i++ {
		client := e.createNewClient()

		_, err := client.PostRequest(inccounter.FuncIncCounter.Name)
		for i := 0; i < 5 && (err != nil); i++ {
			fmt.Printf("Error posting request, will retry... %v", err)
			time.Sleep(100 * time.Millisecond)
			_, err = client.PostRequest(inccounter.FuncIncCounter.Name)
		}
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(numRequests)), util.MakeRange(0, clusterSize), 40*time.Second, "a required number of testAccessNodesOnLedger requests")

	e.printBlocks(
		numRequests + // The actual IncCounter requests.
			3 + // Initial State + IncCounter SC Deploy + ???
			clusterSize, // Access node applications.
	)
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
	cmt := util.MakeRange(0, numValidatorNodes)

	e := setupAdvancedInccounterTest(t, clusterSize, cmt)

	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	accountsClient := e.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	tx, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: iscp.NewTokensIotas(1_000_000),
	})
	require.NoError(t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	myClient := e.Chain.SCClient(iscp.Hn(nativeIncCounterSCName), keyPair)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 1)})
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(numRequests)), util.MakeRange(0, clusterSize), to, "requests counted")

	e.printBlocks(
		numRequests + // The actual IncCounter requests.
			4 + // ???
			clusterSize, // Access nodes applications.
	)
	time.Sleep(10 * time.Second) // five time for the nodes to shutdown properly before running the next test
}

// extreme test
func TestAccessNodesMany(t *testing.T) {
	testutil.RunHeavy(t)
	const clusterSize = 15
	const numValidatorNodes = 6
	const requestsCountInitial = 2
	const requestsCountProgression = 2
	const iterationCount = 9

	e := setupAdvancedInccounterTest(t, clusterSize, util.MakeRange(0, numValidatorNodes))

	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := e.Chain.SCClient(nativeIncCounterSCHname, keyPair)

	requestsCount := requestsCountInitial
	requestsCumulative := 0
	posted := 0
	for i := 0; i < iterationCount; i++ {
		logMsg := fmt.Sprintf("iteration %v of %v requests", i, requestsCount)
		t.Logf("Running %s", logMsg)
		_, err := myClient.PostRequests(inccounter.FuncIncCounter.Name, requestsCount)
		require.NoError(t, err)
		posted += requestsCount
		requestsCumulative += requestsCount
		waitUntil(t, e.counterEquals(int64(requestsCumulative)), e.Clu.Config.AllNodes(), 60*time.Second, logMsg)
		requestsCount *= requestsCountProgression
	}
	e.printBlocks(
		posted + // The actual SC requests.
			3 + // ???
			clusterSize, // GOV: Access Node Applications.
	)
}

// cluster of 10 access nodes and two overlapping committees
func TestRotation(t *testing.T) {
	numRequests := 8

	cmt1 := []int{0, 1, 2, 3}
	cmt2 := []int{2, 3, 4, 5}

	clu := newCluster(t, waspClusterOpts{nNodes: 10})
	addr1, err := clu.RunDKG(cmt1, 3)
	require.NoError(t, err)
	addr2, err := clu.RunDKG(cmt2, 3)
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Bech32(parameters.L1.Protocol.Bech32HRP))
	t.Logf("addr2: %s", addr2.Bech32(parameters.L1.Protocol.Bech32HRP))

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt1, 3, addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID)

	e := newChainEnv(t, clu, chain)

	tx := e.deployNativeIncCounterSC(0)

	waitUntil(t, e.contractIsDeployed(nativeIncCounterSCName), clu.Config.AllNodes(), 50*time.Second, "contract to be deployed")

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)
	require.NoError(t, e.waitStateControllers(addr1, 5*time.Second))

	keyPair, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := chain.SCClient(nativeIncCounterSCHname, keyPair)

	_, err = myClient.PostRequests(inccounter.FuncIncCounter.Name, numRequests)
	require.NoError(t, err)

	waitUntil(t, e.counterEquals(int64(numRequests)), []int{0, 3, 8, 9}, 5*time.Second)

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair)

	params := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err = govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress.Name, *params)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 15*time.Second)
	require.NoError(t, err)
	require.True(t, isAllowedStateControllerAddress(t, chain, 0, addr2))
	require.NoError(t, e.waitStateControllers(addr1, 15*time.Second))

	params = chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err = govClient.PostRequest(governance.FuncRotateStateController.Name, *params)
	require.NoError(t, err)
	require.NoError(t, e.waitStateControllers(addr2, 15*time.Second))
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 15*time.Second)
	require.NoError(t, err)

	_, err = myClient.PostRequests(inccounter.FuncIncCounter.Name, numRequests)
	require.NoError(t, err)

	waitUntil(t, e.counterEquals(int64(2*numRequests)), clu.Config.AllNodes(), 15*time.Second)
}

type testRotationManyRotation struct {
	Committee []int
	Quorum    uint16
	Address   iotago.Address
}

func newTestRotationManyRotation(t *testing.T, clu *cluster.Cluster, committee []int, quorum uint16) testRotationManyRotation {
	address, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)
	return testRotationManyRotation{
		Committee: committee,
		Quorum:    quorum,
		Address:   address,
	}
}

func TestRotationMany(t *testing.T) {
	testutil.RunHeavy(t)
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	const numRequests = 2
	const waitTimeout = 180 * time.Second

	clu := newCluster(t, waspClusterOpts{nNodes: 10})
	rotations := []testRotationManyRotation{
		newTestRotationManyRotation(t, clu, []int{0, 1, 2, 3}, 3),
		newTestRotationManyRotation(t, clu, []int{2, 3, 4, 5}, 3),
		newTestRotationManyRotation(t, clu, []int{3, 4, 5, 6, 7, 8}, 5),
		newTestRotationManyRotation(t, clu, []int{9, 4, 5, 6, 7, 8, 3}, 5),
		newTestRotationManyRotation(t, clu, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 7),
	}

	t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotations[0].Committee, rotations[0].Quorum, rotations[0].Address)
	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), rotations[0].Committee, rotations[0].Quorum, rotations[0].Address)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID)

	e := newChainEnv(t, clu, chain)

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair)

	for _, rotation := range rotations {
		t.Logf("Adding address %s of committee %v to allowed state controller addresses", rotation.Address, rotation.Committee)
		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation.Address).WithIotas(1 * iscp.Mi)
		tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress.Name, *par)
		require.NoError(t, err)
		_, err = e.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, waitTimeout)
		require.NoError(t, err)
		require.NoError(t, e.checkAllowedStateControllerAddressInAllNodes(rotation.Address))
	}

	tx := e.deployNativeIncCounterSC(0)

	waitUntil(t, e.contractIsDeployed(nativeIncCounterSCName), e.Clu.Config.AllNodes(), 30*time.Second)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, waitTimeout)
	require.NoError(t, err)

	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := chain.SCClient(nativeIncCounterSCHname, keyPair)

	for i, rotation := range rotations {
		t.Logf("Rotating to %v-th committee %v with quorum %v and address %s", i, rotation.Committee, rotation.Quorum, rotation.Address)

		_, err = myClient.PostRequests(inccounter.FuncIncCounter.Name, numRequests)
		require.NoError(t, err)

		waitUntil(t, e.counterEquals(int64(numRequests*(i+1))), e.Clu.Config.AllNodes(), 30*time.Second)

		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation.Address).WithIotas(1 * iscp.Mi)
		tx, err := govClient.PostRequest(governance.FuncRotateStateController.Name, *par)
		require.NoError(t, err)
		require.NoError(t, e.waitStateControllers(rotation.Address, waitTimeout))
		_, err = e.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, waitTimeout)
		require.NoError(t, err)
	}
}

func (e *ChainEnv) waitBlockIndex(nodeIndex int, blockIndex uint32, timeout time.Duration) bool {
	return waitTrue(timeout, func() bool {
		i, err := e.callGetBlockIndex(nodeIndex)
		return err == nil && i >= blockIndex
	})
}

func (e *ChainEnv) callGetBlockIndex(nodeIndex int) (uint32, error) {
	ret, err := e.Chain.Cluster.WaspClient(nodeIndex).CallView(
		e.Chain.ChainID,
		blocklog.Contract.Hname(),
		blocklog.ViewGetLatestBlockInfo.Name,
		nil,
	)
	if err != nil {
		return 0, err
	}
	v, err := codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex))
	require.NoError(e.t, err)
	return v, nil
}

func (e *ChainEnv) waitStateControllers(addr iotago.Address, timeout time.Duration) error {
	for _, nodeIndex := range e.Clu.AllNodes() {
		if err := e.waitStateController(nodeIndex, addr, timeout); err != nil {
			return err
		}
	}
	return nil
}

func (e *ChainEnv) waitStateController(nodeIndex int, addr iotago.Address, timeout time.Duration) error {
	var err error
	result := waitTrue(timeout, func() bool {
		var a iotago.Address
		a, err = e.callGetStateController(nodeIndex)
		if err != nil {
			return false
		}
		return a.Equal(addr)
	})
	if err != nil {
		return err
	}
	if !result {
		return xerrors.New(fmt.Sprintf("Timeout waiting state controler change to %s in node %v", addr, nodeIndex))
	}
	return nil
}

func (e *ChainEnv) callGetStateController(nodeIndex int) (iotago.Address, error) {
	ret, err := e.Chain.Cluster.WaspClient(nodeIndex).CallView(
		e.Chain.ChainID,
		blocklog.Contract.Hname(),
		blocklog.ViewControlAddresses.Name,
		nil,
	)
	if err != nil {
		return nil, err
	}
	addr, err := codec.DecodeAddress(ret.MustGet(blocklog.ParamStateControllerAddress))
	require.NoError(e.t, err)
	return addr, nil
}

func (e *ChainEnv) checkAllowedStateControllerAddressInAllNodes(addr iotago.Address) error {
	for _, i := range e.Chain.AllPeers {
		if !isAllowedStateControllerAddress(e.t, e.Chain, i, addr) {
			return fmt.Errorf("Node %v state controller address is not %s", i, addr)
		}
	}
	return nil
}

func isAllowedStateControllerAddress(t *testing.T, chain *cluster.Chain, nodeIndex int, addr iotago.Address) bool {
	ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		governance.Contract.Hname(),
		governance.ViewGetAllowedStateControllerAddresses.Name,
		nil,
	)
	require.NoError(t, err)
	arr := collections.NewArray16ReadOnly(ret, string(governance.ParamAllowedStateControllerAddresses))
	arrlen := arr.MustLen()
	if arrlen == 0 {
		return false
	}
	for i := uint16(0); i < arrlen; i++ {
		a, err := codec.DecodeAddress(arr.MustGetAt(i))
		require.NoError(t, err)
		if a.Equal(addr) {
			return true
		}
	}
	return false
}
