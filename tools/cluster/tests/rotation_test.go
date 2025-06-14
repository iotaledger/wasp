package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/cluster"
)

func TestBasicRotation(t *testing.T) { // FIXME serious error
	t.Skipf("TODO: Rotation requires refactoring to work, skipped for now")
	/*
		env := setupNativeInccounterTest(t, 6, []int{0, 1, 2, 3})

		newCmtAddr, err := env.Clu.RunDKG([]int{2, 3, 4, 5}, 3)
		require.NoError(t, err)

		kp, _, err := env.Clu.NewKeyPairWithFunds()
		require.NoError(t, err)

		myClient := env.Chain.Client(kp)

		// check the chain works
		tx, err := myClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
			Transfer:  isc.NewAssets(10 + iotaclient.DefaultGasBudget),
			GasBudget: iotaclient.DefaultGasBudget,
		})

		require.NoError(t, err)
		_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), env.Chain.ChainID, tx, false, 20*time.Second)
		require.NoError(t, err)

		// change the committee to the new one
		govClient := env.Chain.Client(env.Chain.OriginatorKeyPair)

		tx, err = govClient.PostRequest(context.Background(),
			governance.FuncAddAllowedStateControllerAddress.Message(newCmtAddr), chainclient.PostRequestParams{
				GasBudget: iotaclient.DefaultGasBudget,
			})

		require.NoError(t, err)
		_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), env.Chain.ChainID, tx, false, 20*time.Second)
		require.NoError(t, err)

		tx, err = govClient.PostRequest(context.Background(), governance.FuncRotateStateController.Message(newCmtAddr),
			chainclient.PostRequestParams{
				GasBudget: 5 * iotaclient.DefaultGasBudget,
			})
		require.NoError(t, err)
		time.Sleep(3 * time.Second)
		_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), env.Chain.ChainID, tx, true, 20*time.Second)
		require.NoError(t, err)

		stateController, err := env.callGetStateController(0)
		require.NoError(t, err)
		require.True(t, stateController.Equals(newCmtAddr), "StateController, expected=%v, received=%v", newCmtAddr, stateController)

		// check the chain still works
		tx, err = myClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
			Transfer:  isc.NewAssets(10 + iotaclient.DefaultGasBudget),
			GasBudget: iotaclient.DefaultGasBudget,
		})
		require.NoError(t, err)
		_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), env.Chain.ChainID, tx, false, 20*time.Second)
		require.NoError(t, err)
	*/
}

// cluster of 10 access nodes and two overlapping committees
func TestRotation(t *testing.T) {
	t.Skipf("TODO: Rotation requires refactoring to work, skipped for now")

	/*
		numRequests := 8

		clu := newCluster(t, waspClusterOpts{nNodes: 10})
		rotation1 := newTestRotationSingleRotation(t, clu, []int{0, 1, 2, 3}, 3)
		rotation2 := newTestRotationSingleRotation(t, clu, []int{2, 3, 4, 5}, 3)

		t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotation1.Committee, rotation1.Quorum, rotation1.Address)
		chain, err := clu.DeployChain(clu.Config.AllNodes(), rotation1.Committee, rotation1.Quorum, rotation1.Address)
		require.NoError(t, err)
		t.Logf("chainID: %s", chain.ChainID)

		chEnv := newChainEnv(t, clu, chain)

		require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 5*time.Second))

		keyPair, _, err := clu.NewKeyPairWithFunds()
		require.NoError(t, err)

		myClient := chain.Client(keyPair)

		_, err = myClient.PostMultipleRequests(context.Background(), inccounter.FuncIncCounter.Message(nil), numRequests)
		require.NoError(t, err)

		waitUntil(t, chEnv.counterEquals(int64(numRequests)), chEnv.Clu.Config.AllNodes(), 5*time.Second)

		govClient := chain.Client(chain.OriginatorKeyPair)

		t.Logf("Adding address %s of committee %v to allowed state controller addresses", rotation2.Address, rotation2.Committee)
		tx, err := govClient.PostRequest(context.Background(), governance.FuncAddAllowedStateControllerAddress.Message(rotation2.Address),
			*chainclient.NewPostRequestParams().WithBaseTokens(1 * isc.Million),
		)
		require.NoError(t, err)
		_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chEnv.Chain.ChainID, tx, false, 15*time.Second)
		require.NoError(t, err)
		require.NoError(t, chEnv.checkAllowedStateControllerAddressInAllNodes(rotation2.Address))
		require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 15*time.Second))

		t.Logf("Rotating to committee %v with quorum %v and address %s", rotation2.Committee, rotation2.Quorum, rotation2.Address)
		tx, err = govClient.PostRequest(context.Background(), governance.FuncRotateStateController.Message(rotation2.Address), chainclient.PostRequestParams{})
		require.NoError(t, err)
		require.NoError(t, chEnv.waitStateControllers(rotation2.Address, 15*time.Second))
		_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chEnv.Chain.ChainID, tx, false, 15*time.Second)
		require.NoError(t, err)

		_, err = myClient.PostMultipleRequests(context.Background(), inccounter.FuncIncCounter.Message(nil), numRequests)
		require.NoError(t, err)

		waitUntil(t, chEnv.counterEquals(int64(2*numRequests)), clu.Config.AllNodes(), 15*time.Second)
	*/
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
	t.Skip("TODO: Cluster tests currently disabled")

	/*
		numRequests := 16

		clu := newCluster(t, waspClusterOpts{nNodes: 10})
		rotation1 := newTestRotationSingleRotation(t, clu, []int{0}, 1)
		rotation2 := newTestRotationSingleRotation(t, clu, []int{1, 2, 3, 4}, 3)

		t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotation1.Committee, rotation1.Quorum, rotation1.Address)
		chain, err := clu.DeployChain(clu.Config.AllNodes(), rotation1.Committee, rotation1.Quorum, rotation1.Address)
		require.NoError(t, err)
		t.Logf("chainID: %s", chain.ChainID)

		chEnv := newChainEnv(t, clu, chain)
		require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 30*time.Second))
		incCounterResultChan := make(chan error)

		go func() {
			keyPair, _, err2 := clu.NewKeyPairWithFunds()
			if err2 != nil {
				incCounterResultChan <- fmt.Errorf("failed to create a key pair: %w", err2)
				return
			}
			myClient := chain.Client(keyPair)
			for i := 0; i < numRequests; i++ {
				t.Logf("Posting inccounter request number %v", i)
				_, err2 = myClient.PostRequest(context.Background(), inccounter.FuncIncCounter.Message(nil), chainclient.PostRequestParams{
					GasBudget: iotaclient.DefaultGasBudget,
				})
				if err2 != nil {
					incCounterResultChan <- fmt.Errorf("failed to post inccounter request number %v: %w", i, err2)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
			incCounterResultChan <- nil
		}()

		govClient := chain.Client(chain.OriginatorKeyPair)

		time.Sleep(500 * time.Millisecond)
		t.Logf("Adding address %s of committee %v to allowed state controller addresses", rotation2.Address, rotation2.Committee)
		tx, err := govClient.PostRequest(context.Background(), governance.FuncAddAllowedStateControllerAddress.Message(rotation2.Address),
			*chainclient.NewPostRequestParams().WithBaseTokens(1 * isc.Million),
		)
		require.NoError(t, err)
		_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chEnv.Chain.ChainID, tx, false, 30*time.Second)
		require.NoError(t, err)
		require.NoError(t, chEnv.checkAllowedStateControllerAddressInAllNodes(rotation2.Address))
		require.NoError(t, chEnv.waitStateControllers(rotation1.Address, 15*time.Second))

		time.Sleep(500 * time.Millisecond)
		t.Logf("Rotating to committee %v with quorum %v and address %s", rotation2.Committee, rotation2.Quorum, rotation2.Address)
		tx, err = govClient.PostRequest(context.Background(), governance.FuncRotateStateController.Message(rotation2.Address), chainclient.PostRequestParams{
			GasBudget: iotaclient.DefaultGasBudget,
		})
		require.NoError(t, err)
		require.NoError(t, chEnv.waitStateControllers(rotation2.Address, 30*time.Second))
		_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chEnv.Chain.ChainID, tx, false, 30*time.Second)
		require.NoError(t, err)

		select {
		case incCounterResult := <-incCounterResultChan:
			require.NoError(t, incCounterResult)
		case <-time.After(1 * time.Minute):
			t.Fatal("Timeout waiting incCounterResult")
		}

		waitUntil(t, chEnv.counterEquals(int64(numRequests)), chEnv.Clu.Config.AllNodes(), 30*time.Second)
	*/
}

type testRotationSingleRotation struct {
	Committee []int
	Quorum    uint16
	Address   *cryptolib.Address
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
	t.Skip("TODO: Cluster tests currently disabled")
	/*
		testutil.RunHeavy(t)
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		const numRequests = 2
		const waitTimeout = 260 * time.Second

		clu := newCluster(t, waspClusterOpts{nNodes: 10})
		rotations := []testRotationSingleRotation{
			newTestRotationSingleRotation(t, clu, []int{0, 1, 2, 3}, 3),
			newTestRotationSingleRotation(t, clu, []int{2, 3, 4, 5}, 3),
			newTestRotationSingleRotation(t, clu, []int{3, 4, 5, 6, 7, 8}, 5),
			newTestRotationSingleRotation(t, clu, []int{9, 4, 5, 6, 7, 8, 3}, 5),
			newTestRotationSingleRotation(t, clu, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 7),
		}

		t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotations[0].Committee, rotations[0].Quorum, rotations[0].Address)
		chain, err := clu.DeployChain(clu.Config.AllNodes(), rotations[0].Committee, rotations[0].Quorum, rotations[0].Address)
		require.NoError(t, err)
		t.Logf("chainID: %s", chain.ChainID)

		chEnv := newChainEnv(t, clu, chain)

		govClient := chain.Client(chain.OriginatorKeyPair)

		for _, rotation := range rotations {
			t.Logf("Adding address %s of committee %v to allowed state controller addresses", rotation.Address, rotation.Committee)
			tx, err2 := govClient.PostRequest(context.Background(), governance.FuncAddAllowedStateControllerAddress.Message(rotation.Address),
				*chainclient.NewPostRequestParams().WithBaseTokens(1 * isc.Million),
			)
			require.NoError(t, err2)
			_, err2 = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chEnv.Chain.ChainID, tx, false, waitTimeout)
			require.NoError(t, err2)
			require.NoError(t, chEnv.checkAllowedStateControllerAddressInAllNodes(rotation.Address))
		}

		keyPair, _, err := chEnv.Clu.NewKeyPairWithFunds()
		require.NoError(t, err)

		myClient := chain.Client(keyPair)

		for i, rotation := range rotations {
			t.Logf("Rotating to %v-th committee %v with quorum %v and address %s", i, rotation.Committee, rotation.Quorum, rotation.Address)

			_, err = myClient.PostMultipleRequests(context.Background(), inccounter.FuncIncCounter.Message(nil), numRequests)
			require.NoError(t, err)

			waitUntil(t, chEnv.counterEquals(int64(numRequests*(i+1))), chEnv.Clu.Config.AllNodes(), 30*time.Second)

			tx, err := govClient.PostRequest(context.Background(), governance.FuncRotateStateController.Message(rotation.Address), chainclient.PostRequestParams{
				GasBudget: iotaclient.DefaultGasBudget,
			})
			require.NoError(t, err)
			require.NoError(t, chEnv.waitStateControllers(rotation.Address, waitTimeout))
			_, err = chEnv.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), chEnv.Chain.ChainID, tx, false, waitTimeout)
			require.NoError(t, err)
		}
	*/
}
