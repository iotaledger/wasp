/**
This test will test the recovery capabilities of the clusterized nodes, where some nodes can break/restart/lay dead at any time.
*/

package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/stretchr/testify/require"
)

type SabotageEnv struct {
	chainEnv      *chainEnv
	NumValidators int
	From          int
	To            int
	SabotageList  []int
}

func InitializeStabilityTest(t *testing.T, numValidators, clusterSize int) *SabotageEnv {
	progHash := inccounter.Contract.ProgramHash
	env := setupWithChain(t, clusterSize)
	_, _, err := env.clu.InitDKG(numValidators)

	require.NoError(t, err)

	_, _ = env.chain.DeployContract(incCounterSCName, progHash.String(), "testing with inccounter", nil)
	waitUntil(t, env.contractIsDeployed(incCounterSCName), env.clu.Config.AllNodes(), 50*time.Second, "incCounter matches expectation")

	return &SabotageEnv{
		chainEnv:      env,
		NumValidators: numValidators,
		From:          0,
		To:            0,
		SabotageList:  make([]int, 0),
	}
}

func (e *SabotageEnv) sendRequests(numRequests int) {
	for i := 0; i < numRequests; i++ {
		client := e.chainEnv.createNewClient()

		_, err := client.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(e.chainEnv.t, err)

		time.Sleep(time.Millisecond * 250)
	}
}

func (e *SabotageEnv) setSabotageValidators(breakCount int) {
	clusterSize := e.chainEnv.clu.Config.Wasp.NumNodes
	nodeList := []int{}

	from := clusterSize - e.NumValidators
	to := from + breakCount - 1

	for i := from; i <= to; i++ {
		nodeList = append(nodeList, i)
	}

	e.From = from
	e.To = to
	e.SabotageList = nodeList
}

func (e *SabotageEnv) setSabotageAll(breakCount int) {
	nodeList := []int{}

	from := 1
	to := from + breakCount - 1

	for i := from; i <= to; i++ {
		nodeList = append(nodeList, i)
	}

	e.From = from
	e.To = to
	e.SabotageList = nodeList
}

type SabotageOption int

const (
	SabotageByKill SabotageOption = iota
	// Important: Frozen nodes need to get killed/unfrozen manually after the test is done, otherwise they stay alive after the test has finished
	SabotageByFreeze SabotageOption = iota
)

func (e *SabotageEnv) sabotageNodes(sabotageOption SabotageOption, startDelay, inBetweenDelay time.Duration) *sync.WaitGroup {
	// Give the test time to start

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		e.chainEnv.t.Log("Sabotaging the following nodes:\n")
		e.chainEnv.t.Log(e.SabotageList)

		time.Sleep(startDelay)

		for _, nodeID := range e.SabotageList {
			e.chainEnv.t.Logf("Breaking node: %v (%s)", nodeID, time.Now())

			var err error

			if sabotageOption == SabotageByKill {
				err = e.chainEnv.clu.KillNode(nodeID)
			} else if sabotageOption == SabotageByFreeze {
				err = e.chainEnv.clu.FreezeNode(nodeID)
			}

			require.NoError(e.chainEnv.t, err)

			time.Sleep(inBetweenDelay)
		}

		wg.Done()
	}()

	return &wg
}

func (e *SabotageEnv) restartNodes() {
	for _, nodeID := range e.SabotageList {
		e.chainEnv.t.Logf("Restarting node %v", nodeID)
		err := e.chainEnv.clu.RestartNode(nodeID)

		require.NoError(e.chainEnv.t, err)
	}
}

func (e *SabotageEnv) unfreezeNodes() {
	for _, nodeID := range e.SabotageList {
		e.chainEnv.t.Logf("Unfreezing node %v", nodeID)
		err := e.chainEnv.clu.UnfreezeNode(nodeID)

		require.NoError(e.chainEnv.t, err)
	}
}

func (e *SabotageEnv) getActiveNodeList() []int {
	contains := func(x int) bool {
		for _, n := range e.SabotageList {
			if n == x {
				return true
			}
		}

		return false
	}

	activeNodeList := []int{}

	for _, n := range e.chainEnv.clu.Config.AllNodes() {
		if !contains(n) {
			activeNodeList = append(activeNodeList, n)
		}
	}

	return activeNodeList
}

func runTestSuccessfulIncCounterIncreaseWithoutInstability(t *testing.T, clusterSize, numValidators, numRequests int) {
	env := InitializeStabilityTest(t, numValidators, clusterSize)
	env.sendRequests(numRequests)
	waitUntil(t, env.chainEnv.counterEquals(int64(numRequests)), env.getActiveNodeList(), 120*time.Second, "incCounter matches expectation")
}

func TestSuccessfulIncCounterIncreaseWithoutInstability(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=1,numValidators=1,req=35", func(t *testing.T) {
		const clusterSize = 1
		const numValidators = 1
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithoutInstability(t, clusterSize, numValidators, numRequests)
	})

	t.Run("cluster=5,numValidators=4,req=35", func(t *testing.T) {
		const clusterSize = 5
		const numValidators = 4
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithoutInstability(t, clusterSize, numValidators, numRequests)
	})

	t.Run("cluster=8,numValidators=6,req=35", func(t *testing.T) {
		const clusterSize = 8
		const numValidators = 6
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithoutInstability(t, clusterSize, numValidators, numRequests)
	})
}

func runTestSuccessfulIncCounterIncreaseWithMildInstability(t *testing.T, clusterSize, numValidators, numBrokenNodes, numRequests int) {
	env := InitializeStabilityTest(t, numValidators, clusterSize)
	env.setSabotageValidators(numBrokenNodes)

	wg := env.sabotageNodes(SabotageByKill, 5*time.Second, 1*time.Second)
	env.sendRequests(numRequests)

	wg.Wait()
	// quorum is not met, incCounter should not equal numRequests
	waitUntil(t, env.chainEnv.counterEquals(int64(numRequests)), env.getActiveNodeList(), 120*time.Second, "incCounter matches expectation")
}

func TestSuccessfulIncCounterIncreaseWithMildInstability(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=3,numValidators=2,numBrokenNodes=1,req=35", func(t *testing.T) {
		const clusterSize = 3
		const numValidators = 2
		const numBrokenNodes = 1
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithMildInstability(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=5,numValidators=4,numBrokenNodes=2,req=35", func(t *testing.T) {
		const clusterSize = 5
		const numValidators = 4
		const numBrokenNodes = 2
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithMildInstability(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=7,numValidators=5,numBrokenNodes=2,req=35", func(t *testing.T) {
		const clusterSize = 7
		const numValidators = 5
		const numBrokenNodes = 2
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithMildInstability(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=8,numValidators=6,numBrokenNodes=3,req=35", func(t *testing.T) {
		const clusterSize = 8
		const numValidators = 6
		const numBrokenNodes = 3
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithMildInstability(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=9,numValidators=7,numBrokenNodes=3,req=35", func(t *testing.T) {
		const clusterSize = 9
		const numValidators = 7
		const numBrokenNodes = 3
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithMildInstability(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=9,numValidators=7,numBrokenNodes=4,req=35", func(t *testing.T) {
		const clusterSize = 9
		const numValidators = 7
		const numBrokenNodes = 4
		const numRequests = 35

		runTestSuccessfulIncCounterIncreaseWithMildInstability(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})
}

func runTestFailsIncCounterIncreaseAsQuorumNotMet(t *testing.T, clusterSize, numValidators, numBrokenNodes, numRequests int) {
	env := InitializeStabilityTest(t, numValidators, clusterSize)
	env.setSabotageAll(numBrokenNodes)

	wg := env.sabotageNodes(SabotageByKill, 5*time.Second, 1*time.Second)
	env.sendRequests(numRequests)

	wg.Wait()
	// quorum is not met, incCounter should not equal numRequests
	time.Sleep(time.Second * 25)
	counter := env.chainEnv.getCounter(incCounterSCHname)
	require.NotEqual(t, numRequests, int(counter))
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

	t.Run("cluster=5,numValidators=4,numBrokenNodes=3,req=35", func(t *testing.T) {
		const clusterSize = 5
		const numValidators = 4
		const numBrokenNodes = 3
		const numRequests = 35

		runTestFailsIncCounterIncreaseAsQuorumNotMet(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=7,numValidators=4,numBrokenNodes=5,req=35", func(t *testing.T) {
		const clusterSize = 7
		const numValidators = 4
		const numBrokenNodes = 5
		const numRequests = 35

		runTestFailsIncCounterIncreaseAsQuorumNotMet(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=9,numValidators=8,numBrokenNodes=7,req=35", func(t *testing.T) {
		const clusterSize = 9
		const numValidators = 8
		const numBrokenNodes = 7
		const numRequests = 35

		runTestFailsIncCounterIncreaseAsQuorumNotMet(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})
}

func runTestSuccessfulConsenseusWithReconnectingNodesAndNoQuorum(t *testing.T, clusterSize, numValidators, numBrokenNodes, numRequests int) {
	env := InitializeStabilityTest(t, numValidators, clusterSize)

	t.Cleanup(func() {
		// This hook is just a safety measure to unfreeze all nodes when an error happens. Otherwise they stay in a zombie mode after the tests ended.
		if env != nil {
			env.unfreezeNodes()
		}
	})

	env.setSabotageValidators(numBrokenNodes)

	wg := env.sabotageNodes(SabotageByFreeze, 5*time.Second, 1*time.Second)
	env.sendRequests(numRequests)
	wg.Wait()

	// quorum is not met, work stops, incCounter should not equal numRequests
	t.Log("Waiting for network to settle, no incCounter increases should be logged now")
	time.Sleep(time.Second * 25)
	counter := env.chainEnv.getCounter(incCounterSCHname)
	require.NotEqual(env.chainEnv.t, numRequests, int(counter))

	// unfreeze nodes, after bootstrapping it is expected to reach a full quorum leading to an equal incCounter
	env.unfreezeNodes()

	waitUntil(t, env.chainEnv.counterEquals(int64(numRequests)), env.getActiveNodeList(), 60*time.Second, "incCounter matches expectation")
}

func runTestSuccessfulConsenseusWithReconnectingNodesAndHighQuorum(t *testing.T, clusterSize, numValidators, numBrokenNodes, numRequests int) {
	env := InitializeStabilityTest(t, numValidators, clusterSize)

	t.Cleanup(func() {
		// This hook is just a safety measure to unfreeze all nodes when an error happens. Otherwise they stay in a zombie mode after the tests ended.
		if env != nil {
			env.unfreezeNodes()
		}
	})

	env.setSabotageValidators(numBrokenNodes)

	wg := env.sabotageNodes(SabotageByFreeze, 5*time.Second, 1*time.Second)
	env.sendRequests(numRequests)
	wg.Wait()

	// quorum is not met, work stops, incCounter should not equal numRequests
	t.Log("Waiting to see if nodes keep working, incCounter increases should be logged")
	time.Sleep(time.Second * 25)
	// unfreeze nodes, after bootstrapping it is expected to reach a full quorum leading to an equal incCounter
	waitUntil(t, env.chainEnv.counterEquals(int64(numRequests)), env.getActiveNodeList(), 60*time.Second, "incCounter matches expectation")
	env.unfreezeNodes()
}

func TestSuccessfulConsenseusWithReconnectingNodes(t *testing.T) {
	/**
	* incCounter requests get sent, after reaching a matching counter value, nodes get shut down, new requests get send in parallel.
	*	If killed nodes are below the quorum level, the incCounter count should never reach numRequests, otherwise the opposite is expected
	 */

	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=5,numValidators=4,numBrokenNodes=4,req=35,quorum=NO", func(t *testing.T) {
		const clusterSize = 5
		const numValidators = 4
		const numBrokenNodes = 4
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndNoQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=7,numValidators=5,numBrokenNodes=5,req=35,quorum=NO", func(t *testing.T) {
		const clusterSize = 7
		const numValidators = 5
		const numBrokenNodes = 5
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndNoQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=12,numValidators=10,numBrokenNodes=9,req=35,quorum=NO", func(t *testing.T) {
		const clusterSize = 12
		const numValidators = 10
		const numBrokenNodes = 9
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndNoQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=12,numValidators=10,numBrokenNodes=9,req=35,quorum=NO", func(t *testing.T) {
		const clusterSize = 15
		const numValidators = 13
		const numBrokenNodes = 12
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndNoQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=4,numValidators=3,numBrokenNodes=1,req=35,quorum=YES", func(t *testing.T) {
		const clusterSize = 4
		const numValidators = 3
		const numBrokenNodes = 1
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndHighQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=6,numValidators=4,numBrokenNodes=2,req=35,quorum=YES", func(t *testing.T) {
		const clusterSize = 6
		const numValidators = 4
		const numBrokenNodes = 2
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndHighQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=8,numValidators=7,numBrokenNodes=3,req=35,quorum=YES", func(t *testing.T) {
		const clusterSize = 8
		const numValidators = 7
		const numBrokenNodes = 3
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndHighQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})

	t.Run("cluster=12,numValidators=10,numBrokenNodes=4,req=35,quorum=YES", func(t *testing.T) {
		const clusterSize = 12
		const numValidators = 10
		const numBrokenNodes = 4
		const numRequests = 35

		runTestSuccessfulConsenseusWithReconnectingNodesAndHighQuorum(t, clusterSize, numValidators, numBrokenNodes, numRequests)
	})
}
