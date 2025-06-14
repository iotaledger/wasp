/**
This test will test the recovery capabilities of the clusterized nodes, where some nodes can break/restart/lay dead at any time.
*/

package tests

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

type SabotageEnv struct {
	chainEnv      *ChainEnv
	NumValidators int
	SabotageList  []int
}

func initializeStabilityTest(t *testing.T, numValidators, clusterSize int) *SabotageEnv {
	env := SetupWithChain(t, waspClusterOpts{nNodes: clusterSize})
	_, _, err := env.Clu.InitDKG(numValidators)

	require.NoError(t, err)

	return &SabotageEnv{
		chainEnv:      env,
		NumValidators: numValidators,
		SabotageList:  make([]int, 0),
	}
}

func (e *SabotageEnv) sendRequests(numRequests int, messageDelay time.Duration) (*chainclient.Client, int) {
	client, _ := e.chainEnv.NewRandomChainClient()
	for i := 0; i < numRequests; i++ {
		_, err := client.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
			GasBudget:   iotaclient.DefaultGasBudget,
			Allowance:   isc.NewAssets(iotaclient.DefaultGasBudget),
			L2GasBudget: iotaclient.DefaultGasBudget,
			Transfer:    isc.NewAssets(iotaclient.DefaultGasBudget),
		})
		require.NoError(e.chainEnv.t, err)

		time.Sleep(messageDelay)
	}

	return client, (iotaclient.DefaultGasBudget - BaseTokensDepositFee) * numRequests
}

func (e *SabotageEnv) setSabotageValidators(breakCount int) {
	clusterSize := len(e.chainEnv.Clu.Config.Wasp)

	from := clusterSize - e.NumValidators
	to := from + breakCount - 1

	e.SabotageList = util.MakeRange(from, to)
}

func (e *SabotageEnv) setSabotageAll(breakCount int) {
	from := 1
	to := from + breakCount - 1

	e.SabotageList = util.MakeRange(from, to)
}

func (e *SabotageEnv) sabotageNodes(startDelay, inBetweenDelay time.Duration) *sync.WaitGroup {
	// Give the test time to start

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		e.chainEnv.t.Log("Sabotaging the following nodes:\n")
		e.chainEnv.t.Log(e.SabotageList)

		time.Sleep(startDelay)

		for _, nodeID := range e.SabotageList {
			e.chainEnv.t.Logf("Breaking node: %v (%s)", nodeID, time.Now())

			err := e.chainEnv.Clu.KillNodeProcess(nodeID, false)

			require.NoError(e.chainEnv.t, err)

			time.Sleep(inBetweenDelay)
		}

		wg.Done()
	}()

	return &wg
}

func (e *SabotageEnv) getActiveNodeList() []int {
	contains := func(x int) bool {
		return slices.Contains(e.SabotageList, x)
	}

	activeNodeList := []int{}

	for _, n := range e.chainEnv.Clu.Config.AllNodes() {
		if !contains(n) {
			activeNodeList = append(activeNodeList, n)
		}
	}

	return activeNodeList
}

func TestSuccessfulIncCounterIncreaseWithoutInstability(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	const clusterSize = 8
	const numValidators = 6
	const numRequests = 35

	env := initializeStabilityTest(t, numValidators, clusterSize)
	client, expectedBalance := env.sendRequests(numRequests, time.Millisecond*250)

	waitUntil(t, env.chainEnv.balanceEquals(isc.NewAddressAgentID(client.KeyPair.Address()), expectedBalance), env.chainEnv.Clu.Config.AllNodes(), 120*time.Second, "incCounter matches expectation")
}

func TestSuccessfulIncCounterIncreaseWithMildInstability(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	testutil.RunHeavy(t)

	const clusterSize = 10
	const numValidators = 9
	const numBrokenNodes = 2
	const numRequests = 35

	env := initializeStabilityTest(t, numValidators, clusterSize)
	env.setSabotageValidators(numBrokenNodes)

	wg := env.sabotageNodes(4*time.Second, 1*time.Second)
	client, expectedBalance := env.sendRequests(numRequests, time.Millisecond*250)

	wg.Wait()

	waitUntil(t, env.chainEnv.balanceEquals(isc.NewAddressAgentID(client.KeyPair.Address()), expectedBalance), env.getActiveNodeList(), 120*time.Second, "incCounter matches expectation")
}

func runTestFailsIncCounterIncreaseAsQuorumNotMet(t *testing.T, clusterSize, numValidators, numBrokenNodes, numRequests int) {
	env := initializeStabilityTest(t, numValidators, clusterSize)
	env.setSabotageAll(numBrokenNodes)

	wg := env.sabotageNodes(5*time.Second, 500*time.Millisecond)
	env.sendRequests(numRequests, time.Millisecond*250)

	wg.Wait()
	// quorum is not met, incCounter should not equal numRequests
	time.Sleep(time.Second * 25)
	// counter := env.chainEnv.getNativeContractCounter()
	// require.NotEqual(t, numRequests, int(counter))
}

func TestFailsIncCounterIncreaseAsQuorumNotMet(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=3,numValidators=2,numBrokenNodes=2,req=35", func(t *testing.T) {
		const clusterSize = 3
		const numValidators = 2
		const numBrokenNodes = 2
		const numRequests = 35

		runTestFailsIncCounterIncreaseAsQuorumNotMet(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=14,numValidators=12,numBrokenNodes=11,req=35", func(t *testing.T) {
		testutil.RunHeavy(t)
		const clusterSize = 14
		const numValidators = 12
		const numBrokenNodes = 11
		const numRequests = 35

		runTestFailsIncCounterIncreaseAsQuorumNotMet(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})
}
