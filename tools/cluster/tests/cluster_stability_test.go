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
	waitUntil(t, env.contractIsDeployed(incCounterSCName), env.clu.Config.AllNodes(), 50*time.Second, "contract to be deployed")

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

	from := 0
	to := from + breakCount - 1

	for i := from; i <= to; i++ {
		nodeList = append(nodeList, i)
	}

	e.From = from
	e.To = to
	e.SabotageList = nodeList
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
			err := e.chainEnv.clu.KillNode(nodeID)
			require.NoError(e.chainEnv.t, err)

			time.Sleep(inBetweenDelay)
		}

		wg.Done()
	}()

	return &wg
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

func TestSuccessfulIncCounterIncreaseWithoutInstability(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=1,numValidators=1,req=35", func(t *testing.T) {
		const numRequests = 35
		const numValidators = 1
		const clusterSize = 1

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.sendRequests(numRequests)
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=5,numValidators=4,req=35", func(t *testing.T) {
		const numRequests = 35
		const numValidators = 4
		const clusterSize = 5

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.sendRequests(numRequests)
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=8,numValidators=6,req=35", func(t *testing.T) {
		const numRequests = 35
		const numValidators = 6
		const clusterSize = 8

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.sendRequests(numRequests)
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})
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

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=5,numValidators=4,numBrokenNodes=2,req=35", func(t *testing.T) {
		const clusterSize = 5
		const numValidators = 4
		const numBrokenNodes = 2
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=7,numValidators=5,numBrokenNodes=2,req=35", func(t *testing.T) {
		const clusterSize = 7
		const numValidators = 5
		const numBrokenNodes = 2
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=8,numValidators=6,numBrokenNodes=3,req=35", func(t *testing.T) {
		const clusterSize = 8
		const numValidators = 6
		const numBrokenNodes = 3
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)

		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=9,numValidators=7,numBrokenNodes=3,req=35", func(t *testing.T) {
		const clusterSize = 9
		const numValidators = 7
		const numBrokenNodes = 3
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)

		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=9,numValidators=7,numBrokenNodes=4,req=35", func(t *testing.T) {
		const clusterSize = 9
		const numValidators = 7
		const numBrokenNodes = 4
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})
}

func TestFailsIncCounterIncreaseAsQuorumNotMet(t *testing.T) {
	t.Run("cluster=3,numValidators=2,numBrokenNodes=2,req=35", func(t *testing.T) {
		const clusterSize = 3
		const numValidators = 2
		const numBrokenNodes = 2
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterNotEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=5,numValidators=4,numBrokenNodes=3,req=35", func(t *testing.T) {
		const clusterSize = 5
		const numValidators = 4
		const numBrokenNodes = 3
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageValidators(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterNotEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=7,numValidators=4,numBrokenNodes=5,req=35", func(t *testing.T) {
		const clusterSize = 7
		const numValidators = 4
		const numBrokenNodes = 5
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageAll(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterNotEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})

	t.Run("cluster=9,numValidators=8,numBrokenNodes=7,req=35", func(t *testing.T) {
		const clusterSize = 9
		const numValidators = 8
		const numBrokenNodes = 7
		const numRequests = 35

		env := InitializeStabilityTest(t, numValidators, clusterSize)
		env.setSabotageAll(numBrokenNodes)

		wg := env.sabotageNodes(5*time.Second, 1*time.Second)
		env.sendRequests(numRequests)

		wg.Wait()
		// quorum is not met, incCounter should not equal numRequests
		waitUntil(t, env.chainEnv.counterNotEquals(numRequests), env.getActiveNodeList(), 120*time.Second, "contract to be deployed")
	})
}
