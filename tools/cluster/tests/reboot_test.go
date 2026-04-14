package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

// ensures a nodes resumes normal operation after rebooting all nodes with and without keeping the DB
func TestRebootAllNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	keepDBCases := []bool{false, true}

	for _, keepDB := range keepDBCases {
		t.Run(fmt.Sprintf("keepDB=%v", keepDB), func(t *testing.T) {
			allNodes := []int{0, 1, 2, 3}
			env := createTestWrapper(t, 4, allNodes)
			client, _ := env.NewRandomChainClient()

			env.DepositFunds(100_000_000, client.KeyPair.(*cryptolib.KeyPair)) // For Off-ledger requests to pass.

			// send N requests to node 0
			const numRequests = 1
			err := env.checkNRequests(newClusterTestEnv(t, env, 0), int64(numRequests), 0, env.Clu.Config.AllNodes(), 0)
			require.NoError(t, err)

			// restart the nodes
			err = env.Clu.RestartNodes(keepDB, 0, 1, 2, 3)
			require.NoError(t, err)

			time.Sleep(3 * time.Second)

			err = env.checkNRequests(newClusterTestEnv(t, env, 0), int64(numRequests), 0, env.Clu.Config.AllNodes(), 0)
			require.NoError(t, err)
		})
	}
}

// Test rebooting nodes during operation.
func TestRebootDuringTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}

	env := createTestWrapper(t, 4, []int{0, 1, 2, 3})
	restartDelay := 20 * time.Second
	restartCases := [][]int{
		{1, 2, 3},
		{0, 1, 2, 3},
	}
	postDelay := 200 * time.Millisecond
	postCount := 4 * int(restartDelay/postDelay) * len(restartCases) // To have enough posts for all restarts.

	keyPair, userAddr, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	client := env.Chain.Client(keyPair)

	// keep the nodes spammed with deposit requests
	go func() {
		depositAmount := coin.Value(10_000 + iotagraphql.DefaultGasBudget)
		for i := 0; i < postCount; i++ {
			_, err = client.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
				Transfer:  isc.NewAssets(depositAmount),
				GasBudget: iotagraphql.DefaultGasBudget,
			})
			fmt.Printf("=====> deposit request sent: %d\n", i)
			require.NoError(t, err)
			time.Sleep(postDelay)
		}
	}()

	lastBalance := coin.Value(0)

	restart := func(indexes ...int) {
		t.Logf("restart, indexes=%v", indexes)
		// restart the nodes
		err := env.Clu.RestartNodes(true, indexes...)
		require.NoError(t, err)
		time.Sleep(restartDelay)

		// after rebooting, the chain should resume processing requests without issues
		// verify that the balance is growing (deposits are being processed)
		agentID := isc.NewAddressAgentID(userAddr)
		balance := env.GetL2Balance(agentID, coin.BaseTokenType)
		fmt.Printf("=====> last balance: %d\n", lastBalance)
		fmt.Printf("=====> balance after restart: %d\n", balance)
		require.Greater(t, balance, lastBalance)
		lastBalance = balance
	}

	for _, restartIndexes := range restartCases {
		fmt.Printf("=====> restarting nodes: %v\n", restartIndexes)
		restart(restartIndexes...)
	}
}
