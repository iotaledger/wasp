package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

// Fails in CI
// cluster of 10 access nodes and two overlapping committees with concurrent requests
func TestRotationOverlappingCommitteesWithConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	const numRequests = 8

	clu := newCluster(t, waspClusterOpts{nNodes: 10})
	rotation1 := newTestRotationSingleRotation(t, clu, []int{0, 1, 2, 3}, 3)

	t.Logf("Deploying chain by committee %v with quorum %v and address %s", rotation1.Committee, rotation1.Quorum, rotation1.Address)
	chain, err := clu.DeployChain(clu.Config.AllNodes(), rotation1.Committee, rotation1.Quorum, rotation1.Address, true)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID)

	contractRegistry, err := chain.ContractRegistry(0)
	require.NoError(t, err)
	require.True(t, len(contractRegistry) > 0)

	chEnv := newChainEnv(t, clu, chain)

	waitCommitteeStateAddress(t, clu, rotation1.Address.String(), 6*time.Second, 2)

	keyPair, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	myClient := chain.Client(keyPair)
	myClient.DepositFunds(100 * isc.Million)
	time.Sleep(2 * time.Second)

	storageContractAddr, transactions, err := chEnv.sendNRequests(newClusterTestEnv(t, chEnv, 0), int64(numRequests), 0, true)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// rotate
	rotation2 := newTestRotationSingleRotation(t, clu, []int{2, 3, 4, 5}, 3)
	newAddrStr := rotation2.Address.String()
	for i := 0; i < 4; i++ {
		c := clu.WaspClientFromHostName(clu.Config.APIHost(i))
		rotateRequest := c.ChainsAPI.RotateChain(context.Background()).RotateRequest(apiclient.RotateChainRequest{
			RotateToAddress: &newAddrStr,
		})
		_, err := rotateRequest.Execute()
		require.NoError(t, err)
		time.Sleep(1 * time.Second)
	}

	// we need to do somethind to trigger the rotation. So doing deposit
	_, err = myClient.DepositFunds(10 * isc.Million)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = chEnv.verifyNRequests(context.Background(), transactions, int64(numRequests), clu.AllNodes(), storageContractAddr, nil)
	require.NoError(t, err)

	// Wait until committee state address equals rotation2 address
	waitCommitteeStateAddress(t, clu, rotation2.Address.String(), 20*time.Second, 3)

	err = chEnv.checkNRequests(newClusterTestEnv(t, chEnv, 0), int64(numRequests), 0, clu.AllNodes(), 0)
	require.NoError(t, err)

	// check that state index in anchor equals to state index in storage
	newBlock, _, err := clu.WaspClientFromHostName(clu.Config.APIHost(0)).CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).Execute()
	require.NoError(t, err)

	chainObjId, err := iotago.ObjectIDFromHex(chain.ChainID.String())
	require.NoError(t, err)

	object, err := clu.L1Client().GetObject(context.Background(), *chainObjId)
	require.NoError(t, err)

	var fieldMap map[string]interface{}
	err = json.Unmarshal(object.Object.AsMoveObjectContent.Contents.Data, &fieldMap)
	require.NoError(t, err)

	require.Equal(t, int(fieldMap["state_index"].(float64)), int(newBlock.BlockIndex), "state index in anchor should equal to state index in storage")
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

// waitCommitteeStateAddress waits until the chain's committee reports the given state address
func waitCommitteeStateAddress(t *testing.T, clu *cluster.Cluster, wantAddress string, timeout time.Duration, nodeIndex int) {
	t.Helper()
	client := clu.WaspClientFromHostName(clu.Config.APIHost(nodeIndex))
	deadline := time.Now().Add(timeout)
	for {
		info, _, err := client.ChainsAPI.GetCommitteeInfo(context.Background()).Execute()
		require.NoError(t, err)
		if info.StateAddress == wantAddress {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for committee state address %s", wantAddress)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
