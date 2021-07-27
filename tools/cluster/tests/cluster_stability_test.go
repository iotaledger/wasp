/**
This test will test the recovery capabilities of the clusterized nodes, where some nodes can break/restart/lay dead at any time.
*/

package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
)

func initializeStabilityTest(t *testing.T, numValidators int, clusterSize int) *chainEnv {
	cmt := util.MakeRange(0, numValidators-1)
	quorum := uint16((2*len(cmt))/3 + 1)

	env := setupWithChain(t, clusterSize)
	_, err := env.clu.RunDKG(cmt, quorum)
	require.NoError(t, err)
	description := "testing with inccounter"
	progHash := inccounter.Contract.ProgramHash

	_, _ = env.chain.DeployContract(incCounterSCName, progHash.String(), description, nil)
	waitUntil(t, env.contractIsDeployed(incCounterSCName), env.clu.Config.AllNodes(), 50*time.Second, "contract to be deployed")

	return env
}

func sendRequests(t *testing.T, env *chainEnv, numRequests int) {
	for i := 0; i < numRequests; i++ {
		client := env.createNewClient()

		_, err := client.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)

		t.Logf("Posting request %v", i)
		time.Sleep(time.Millisecond * 50)
	}
}

func sabotageNetwork(t *testing.T, env *chainEnv, numValidators int, clusterSize int) {
	// Give the test time to start
	time.Sleep(time.Second * 15)
	breakPercentage := 0.90
	breakCount := float64(numValidators) * breakPercentage

	t.Logf("Break Percentage: %v, break count: %v", breakPercentage, breakCount)

	for i := numValidators - 1; i >= numValidators-1-int(breakCount); i-- {
		t.Logf("Breaking node: %v", i)
		err := env.clu.FreezeNode(i)
		require.NoError(t, err)

		time.Sleep(time.Second * 5)
	}

	time.Sleep(time.Second * 10)

	for i := numValidators - 1; i >= numValidators-1-int(breakCount); i-- {
		t.Logf("Breaking node: %v", i)
		err := env.clu.UnfreezeNode(i)
		require.NoError(t, err)

		time.Sleep(time.Second * 5)
	}
}

func TestOngoingFailureWithoutRecovery(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=6,N=4,req=8", func(t *testing.T) {
		const numRequests = 99999999
		const numValidators = 10
		const clusterSize = 15

		env := initializeStabilityTest(t, numValidators, clusterSize)
		go sabotageNetwork(t, env, numValidators, clusterSize)
		sendRequests(t, env, numRequests)
	})
}
