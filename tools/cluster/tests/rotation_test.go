package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/cluster"
)

func TestBasicRotation(t *testing.T) {
	env := setupNativeInccounterTest(t, 6, []int{0, 1, 2, 3})

	newCmtAddr, err := env.Clu.RunDKG([]int{2, 3, 4, 5}, 3)
	require.NoError(t, err)

	kp, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := env.Chain.SCClient(isc.Hn(nativeIncCounterSCName), kp)

	// check the chain works
	tx, err := myClient.PostRequest(inccounter.FuncIncCounter.Name)
	isc.MustLogRequestsInTransaction(tx, t.Logf, "Posted request - FuncIncCounter (before rotation)")
	require.NoError(t, err)
	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, 20*time.Second)
	require.NoError(t, err)

	// change the committee to the new one

	govClient := env.Chain.SCClient(governance.Contract.Hname(), env.Chain.OriginatorKeyPair)

	tx, err = govClient.PostRequest(
		governance.FuncAddAllowedStateControllerAddress.Name,
		chainclient.PostRequestParams{
			Args: dict.Dict{
				governance.ParamStateControllerAddress: codec.Encode(newCmtAddr),
			},
		},
	)
	isc.MustLogRequestsInTransaction(tx, t.Logf, "Posted request - FuncAddAllowedStateControllerAddress")
	require.NoError(t, err)
	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, 20*time.Second)
	require.NoError(t, err)

	tx, err = govClient.PostRequest(
		coreutil.CoreEPRotateStateController,
		chainclient.PostRequestParams{
			Args: dict.Dict{
				governance.ParamStateControllerAddress: codec.Encode(newCmtAddr),
			},
		},
	)
	require.NoError(t, err)
	isc.MustLogRequestsInTransaction(tx, t.Logf, "Posted request - CoreEPRotateStateController")
	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, 20*time.Second)
	require.NoError(t, err)

	stateController, err := env.callGetStateController(0)
	require.NoError(t, err)
	require.True(t, stateController.Equal(newCmtAddr), "StateController, expected=%v, received=%v", newCmtAddr, stateController)

	// check the chain still works
	tx, err = myClient.PostRequest(inccounter.FuncIncCounter.Name)
	isc.MustLogRequestsInTransaction(tx, t.Logf, "Posted request - FuncIncCounter")
	require.NoError(t, err)
	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, 20*time.Second)
	require.NoError(t, err)

	require.EqualValues(t, 2, env.getNativeContractCounter(nativeIncCounterSCHname))
}

// cluster of 10 access nodes and two overlapping committees
func TestRotation(t *testing.T) {
	numRequests := 8

	clu := newCluster(t, waspClusterOpts{nNodes: 10})
	rotation1 := newTestRotationSingleRotation(t, clu, []int{0, 1, 2, 3}, 3)
	rotation2 := newTestRotationSingleRotation(t, clu, []int{2, 3, 4, 5}, 3)

	t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotation1.Committee, rotation1.Quorum, rotation1.Address)
	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), rotation1.Committee, rotation1.Quorum, rotation1.Address)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID)

	chEnv := newChainEnv(t, clu, chain)
	chEnv.deployNativeIncCounterSC(0)

	require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 5*time.Second))

	keyPair, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := chain.SCClient(nativeIncCounterSCHname, keyPair)

	_, err = myClient.PostNRequests(inccounter.FuncIncCounter.Name, numRequests)
	require.NoError(t, err)

	waitUntil(t, chEnv.counterEquals(int64(numRequests)), chEnv.Clu.Config.AllNodes(), 5*time.Second)

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair)

	t.Logf("Adding address %s of committee %v to allowed state controller addresses", rotation2.Address, rotation2.Committee)
	params := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation2.Address).WithBaseTokens(1 * isc.Million)
	tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress.Name, *params)
	require.NoError(t, err)
	_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, 15*time.Second)
	require.NoError(t, err)
	require.NoError(t, chEnv.checkAllowedStateControllerAddressInAllNodes(rotation2.Address))
	require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 15*time.Second))

	t.Logf("Rotating to committee %v with quorum %v and address %s", rotation2.Committee, rotation2.Quorum, rotation2.Address)
	params = chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation2.Address).WithBaseTokens(1 * isc.Million)
	tx, err = govClient.PostRequest(governance.FuncRotateStateController.Name, *params)
	require.NoError(t, err)
	require.NoError(t, chEnv.waitStateControllers(rotation2.Address, 15*time.Second))
	_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, 15*time.Second)
	require.NoError(t, err)

	_, err = myClient.PostNRequests(inccounter.FuncIncCounter.Name, numRequests)
	require.NoError(t, err)

	waitUntil(t, chEnv.counterEquals(int64(2*numRequests)), clu.Config.AllNodes(), 15*time.Second)
}

// cluster of 10 access nodes; chain is initialized by one node committee and then
// rotated for four other nodes committee. In parallel of doing this, simple inccounter
// requests are being posted. Test is designed in a way that some inccounter requests
// are approved by the one node committee and others by rotated four node committee.
// NOTE: the timeouts of the test are large, because all the nodes are checked. For
// a request to be marked processed, the node's state manager must be synchronized
// to any index after the transaction, which included the request. It might happen
// that some request is approved by committee for state index 8 and some (most likely
// access) node is constantly behind and catches up only when the test stops producing
// requests in state index 18. In that node, request index 8 is marked as processed
// only after state manager reaches state index 18 and publishes the transaction.
func TestRotationFromSingle(t *testing.T) {
	numRequests := 16

	clu := newCluster(t, waspClusterOpts{nNodes: 10})
	rotation1 := newTestRotationSingleRotation(t, clu, []int{0}, 1)
	rotation2 := newTestRotationSingleRotation(t, clu, []int{1, 2, 3, 4}, 3)

	t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotation1.Committee, rotation1.Quorum, rotation1.Address)
	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), rotation1.Committee, rotation1.Quorum, rotation1.Address)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID)

	chEnv := newChainEnv(t, clu, chain)
	chEnv.deployNativeIncCounterSC(0)

	require.NoError(t, err)
	require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 5*time.Second))
	incCounterResultChan := make(chan error)

	go func() {
		keyPair, _, err2 := clu.NewKeyPairWithFunds()
		if err2 != nil {
			incCounterResultChan <- fmt.Errorf("failed to create a key pair: %w", err2)
			return
		}
		myClient := chain.SCClient(nativeIncCounterSCHname, keyPair)
		for i := 0; i < numRequests; i++ {
			t.Logf("Posting inccounter request number %v", i)
			_, err2 = myClient.PostRequest(inccounter.FuncIncCounter.Name)
			if err2 != nil {
				incCounterResultChan <- fmt.Errorf("failed to post inccounter request number %v: %w", i, err2)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		incCounterResultChan <- nil
	}()

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair)

	time.Sleep(500 * time.Millisecond)
	t.Logf("Adding address %s of committee %v to allowed state controller addresses", rotation2.Address, rotation2.Committee)
	params := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation2.Address).WithBaseTokens(1 * isc.Million)
	tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress.Name, *params)
	require.NoError(t, err)
	_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)
	require.NoError(t, chEnv.checkAllowedStateControllerAddressInAllNodes(rotation2.Address))
	require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 15*time.Second))

	time.Sleep(500 * time.Millisecond)
	t.Logf("Rotating to committee %v with quorum %v and address %s", rotation2.Committee, rotation2.Quorum, rotation2.Address)
	params = chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation2.Address).WithBaseTokens(1 * isc.Million)
	tx, err = govClient.PostRequest(governance.FuncRotateStateController.Name, *params)
	require.NoError(t, err)
	require.NoError(t, chEnv.waitStateControllers(rotation2.Address, 30*time.Second))
	_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	select {
	case incCounterResult := <-incCounterResultChan:
		require.NoError(t, incCounterResult)
	case <-time.After(20 * time.Second):
		t.Fatal("Timeout waiting incCounterResult")
	}

	waitUntil(t, chEnv.counterEquals(int64(numRequests)), chEnv.Clu.Config.AllNodes(), 30*time.Second)
}

type testRotationSingleRotation struct {
	Committee []int
	Quorum    uint16
	Address   iotago.Address
}

func newTestRotationSingleRotation(t *testing.T, clu *cluster.Cluster, committee []int, quorum uint16) testRotationSingleRotation {
	address, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)
	return testRotationSingleRotation{
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
	rotations := []testRotationSingleRotation{
		newTestRotationSingleRotation(t, clu, []int{0, 1, 2, 3}, 3),
		newTestRotationSingleRotation(t, clu, []int{2, 3, 4, 5}, 3),
		newTestRotationSingleRotation(t, clu, []int{3, 4, 5, 6, 7, 8}, 5),
		newTestRotationSingleRotation(t, clu, []int{9, 4, 5, 6, 7, 8, 3}, 5),
		newTestRotationSingleRotation(t, clu, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 7),
	}

	t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotations[0].Committee, rotations[0].Quorum, rotations[0].Address)
	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), rotations[0].Committee, rotations[0].Quorum, rotations[0].Address)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID)

	chEnv := newChainEnv(t, clu, chain)

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair)

	for _, rotation := range rotations {
		t.Logf("Adding address %s of committee %v to allowed state controller addresses", rotation.Address, rotation.Committee)
		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation.Address).WithBaseTokens(1 * isc.Million)
		tx, err2 := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress.Name, *par)
		require.NoError(t, err2)
		_, err2 = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, waitTimeout)
		require.NoError(t, err2)
		require.NoError(t, chEnv.checkAllowedStateControllerAddressInAllNodes(rotation.Address))
	}

	chEnv.deployNativeIncCounterSC(0)

	keyPair, _, err := chEnv.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := chain.SCClient(nativeIncCounterSCHname, keyPair)

	for i, rotation := range rotations {
		t.Logf("Rotating to %v-th committee %v with quorum %v and address %s", i, rotation.Committee, rotation.Quorum, rotation.Address)

		_, err = myClient.PostNRequests(inccounter.FuncIncCounter.Name, numRequests)
		require.NoError(t, err)

		waitUntil(t, chEnv.counterEquals(int64(numRequests*(i+1))), chEnv.Clu.Config.AllNodes(), 30*time.Second)

		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, rotation.Address).WithBaseTokens(1 * isc.Million)
		tx, err := govClient.PostRequest(governance.FuncRotateStateController.Name, *par)
		require.NoError(t, err)
		require.NoError(t, chEnv.waitStateControllers(rotation.Address, waitTimeout))
		_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, waitTimeout)
		require.NoError(t, err)
	}
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
			e.t.Logf("Error received while waiting state controller change to %s in node %d", addr, nodeIndex)
			return false
		}
		return a.Equal(addr)
	})
	if err != nil {
		return err
	}
	if !result {
		return fmt.Errorf("timeout waiting state controller change to %s in node %d", addr, nodeIndex)
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
			return fmt.Errorf("state controller address %s is not allowed in node %d", addr, i)
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
	arr := collections.NewArray16ReadOnly(ret, governance.ParamAllowedStateControllerAddresses)
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
