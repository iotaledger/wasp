// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/stretchr/testify/require"
)

const waitMempoolTimeout = 3 * time.Minute

func TestConsensusEnvMockedACS(t *testing.T) {
	t.Run("wait index mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		env.StartTimers()
		env.SetInitialConsensusState()
		err := env.WaitStateIndex(4, 0)
		require.NoError(t, err)
	})
	t.Run("wait timer tick", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		env.StartTimers()
		env.SetInitialConsensusState()
		err := env.WaitTimerTick(43)
		require.NoError(t, err)
	})
}

func TestConsensusPostRequestMockedACS(t *testing.T) {
	t.Run("post 1 mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1, true)
		err := env.WaitMempool(1, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 10 requests mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 10 requests post randomized mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 100 requests mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(100)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(100, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 100 requests randomized mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(100, true)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(100, 3, 5*time.Second)
		require.NoError(t, err)
	})
}

func TestConsensusMoreNodesMockedACS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	const numNodes = 22
	const quorum = (numNodes*2)/3 + 1

	t.Run("post 1 mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, quorum, 15*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1, true)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitStateIndex(quorum, 1)
		require.NoError(t, err)
	})
	t.Run("post 10 requests mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, quorum, 15*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 10 requests randomized mocked ACS", func(t *testing.T) {
		env := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, quorum, 15*time.Second)
		require.NoError(t, err)
	})
}

//-------------------------------------------------

func TestConsensusEnv(t *testing.T) {
	t.Run("wait index", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		env.StartTimers()
		env.SetInitialConsensusState()
		err := env.WaitStateIndex(4, 0)
		require.NoError(t, err)
	})
	t.Run("wait timer tick", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		env.StartTimers()
		env.SetInitialConsensusState()
		err := env.WaitTimerTick(43)
		require.NoError(t, err)
	})
}

func TestConsensusPostRequest(t *testing.T) {
	t.Run("post 1", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1, true)
		err := env.WaitMempool(1, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests post randomized", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 100 requests", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(100)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(100, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 100 requests randomized", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, 4, 3, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()
		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(100, true)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(100, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
}

func TestConsensusMoreNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	const numNodes = 22
	const quorum = (numNodes*2)/3 + 1

	t.Run("post 1", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(1, true)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(1, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests randomized", func(t *testing.T) {
		env := consensus.NewMockedEnv(t, numNodes, quorum, false)
		env.CreateNodes(consensus.NewConsensusTimers())
		defer env.Log.Sync()

		env.StartTimers()
		env.SetInitialConsensusState()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
}

func TestCruelWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	const numNodes = 10
	const quorum = (numNodes*2)/3 + 1
	env := consensus.NewMockedEnv(t, numNodes, quorum, false)
	timers := consensus.NewConsensusTimers()
	timers.BroadcastSignedResultRetry = 50 * time.Millisecond
	env.CreateNodes(timers)
	env.NetworkBehaviour.
		WithLosingChannel(nil, 80).
		WithRepeatingChannel(nil, 25).
		WithDelayingChannel(nil, 0*time.Millisecond, 200*time.Millisecond)
	env.StartTimers()
	env.SetInitialConsensusState()

	randFromIntervalFun := func(from int, till int) time.Duration {
		return time.Duration(from + rand.Intn(till-from))
	}
	var disconnectedNodes []string
	var mutex sync.Mutex
	go func() { // Connection cutter
		for {
			time.Sleep(randFromIntervalFun(1000, 3000) * time.Millisecond)
			mutex.Lock()
			// nodeName := env.Nodes[rand.Intn(numNodes)].NodeID
			// env.NetworkBehaviour.WithPeerDisconnected(&nodeName, nodeName)
			// env.Log.Debugf("Connection to node %v lost", nodeName)
			// disconnectedNodes = append(disconnectedNodes, nodeName)
			mutex.Unlock() // nolint:gocritic,staticcheck
		}
	}()

	go func() { // Connection restorer
		for {
			time.Sleep(randFromIntervalFun(500, 2000) * time.Millisecond)
			mutex.Lock()
			if len(disconnectedNodes) > 0 {
				env.NetworkBehaviour.RemoveHandler(disconnectedNodes[0])
				env.Log.Debugf("Connection to node %v restored", disconnectedNodes[0])
				disconnectedNodes[0] = ""
				disconnectedNodes = disconnectedNodes[1:]
			}
		}
	}()

	env.PostDummyRequests(1, true)
	err := env.WaitMempool(1, quorum, waitMempoolTimeout)
	require.NoError(t, err)

	env.PostDummyRequests(10, true)
	err = env.WaitMempool(11, quorum, waitMempoolTimeout)
	require.NoError(t, err)

	env.PostDummyRequests(100, true)
	err = env.WaitMempool(111, quorum, waitMempoolTimeout)
	require.NoError(t, err)
}
