package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
)

func TestMissingRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}
	clu := newCluster(t, waspClusterOpts{nNodes: 4})
	cmt := []int{0, 1, 2, 3}
	threshold := uint16(4)
	addr, err := clu.RunDistributedKeyGeneration(cmt, threshold)
	require.NoError(t, err)

	chain, err := clu.DeployChain(clu.Config.AllNodes(), cmt, threshold, addr, false)
	require.NoError(t, err)

	chEnv := newChainEnv(t, clu, chain)

	userWallet, _, err := chEnv.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	// deposit funds before sending the off-ledger request
	chEnv.DepositFunds(iotaclient.DefaultGasBudget, userWallet)

	// send N requests to node 0
	const numRequests = 5
	storageContractAddr, transactions, err := chEnv.sendNRequests(newClusterTestEnv(t, chEnv, 0), int64(numRequests), 0, true)
	require.NoError(t, err)

	// verify N requests on all nodes
	err = chEnv.verifyNRequests(context.Background(), transactions, int64(numRequests), clu.Config.AllNodes(), storageContractAddr, nil)
	require.NoError(t, err)
}
