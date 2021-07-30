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

func initializeStabilityTest(t *testing.T, numValidators, clusterSize int) *chainEnv {
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

		time.Sleep(time.Millisecond * 15)
	}
}

type SabotageModel struct {
	From int
	To   int
	List []int
}

func createValidatorNodeSabotage(breakPercentage float64, numValidators, clusterSize int) SabotageModel {
	breakCount := float64(numValidators) * breakPercentage
	nodeList := []int{}

	from := clusterSize - numValidators
	to := clusterSize - numValidators + int(breakCount)

	for i := from; i <= to; i++ {
		nodeList = append(nodeList, i)
	}

	return SabotageModel{
		From: from,
		To:   to,
		List: nodeList,
	}
}

func sabotageNodes(t *testing.T, env *chainEnv, sabotageConfiguration SabotageModel) {
	// Give the test time to start
	t.Log("Sabotaging the following nodes:\n")
	t.Log(sabotageConfiguration.List)
	time.Sleep(time.Second * 15)
	for i := sabotageConfiguration.From; i <= sabotageConfiguration.To; i++ {
		t.Logf("Breaking node: %v", i)
		err := env.clu.KillNode(i)
		require.NoError(t, err)

		// time.Sleep(time.Second * 5)
	}
}

func getActiveNodeList(allNodes, brokenNodes []int) []int {
	contains := func(x int) bool {
		for _, n := range brokenNodes {
			if n == x {
				return true
			}
		}

		return false
	}

	activeNodeList := []int{}

	for _, n := range allNodes {
		if !contains(n) {
			activeNodeList = append(activeNodeList, n)
		}
	}

	return activeNodeList
}

func TestOngoingFailureWithoutRecovery(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=6,N=4,req=8", func(t *testing.T) {
		const numRequests = 35
		const numValidators = 10
		const clusterSize = 15

		sabotageConfiguration := createValidatorNodeSabotage(0.2, numValidators, clusterSize)
		env := initializeStabilityTest(t, numValidators, clusterSize)

		go sabotageNodes(t, env, sabotageConfiguration)

		sendRequests(t, env, numRequests)

		waitUntil(t, env.counterEquals(numRequests), getActiveNodeList(env.clu.Config.AllNodes(), sabotageConfiguration.List), 120*time.Second, "contract to be deployed")
	})
}
