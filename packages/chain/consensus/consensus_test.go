// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain/consensus"
	"github.com/stretchr/testify/require"
)

const waitMempoolTimeout = 3 * time.Minute

func TestConsensusEnvMockedACS(t *testing.T) {
	t.Run("wait index mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		env.StartTimers()
		env.EventStateTransition()
		err := env.WaitStateIndex(4, 0)
		require.NoError(t, err)
	})
	t.Run("wait timer tick", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		env.StartTimers()
		env.EventStateTransition()
		env.WaitTimerTick(43)
	})
}

func TestConsensusPostRequestMockedACS(t *testing.T) {
	t.Run("post 1 mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1, true)
		err := env.WaitMempool(1, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 10 requests mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 10 requests post randomized mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 100 requests mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(100)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(100, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 100 requests randomized mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
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
		env, _ := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, quorum, 15*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1, true)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitStateIndex(quorum, 1)
		require.NoError(t, err)
	})
	t.Run("post 10 requests mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, quorum, 15*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 10 requests randomized mocked ACS", func(t *testing.T) {
		env, _ := consensus.NewMockedEnvWithMockedACS(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, quorum, 15*time.Second)
		require.NoError(t, err)
	})
}

//-------------------------------------------------

func TestConsensusEnv(t *testing.T) {
	t.Run("wait index", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		env.StartTimers()
		env.EventStateTransition()
		err := env.WaitStateIndex(4, 0)
		require.NoError(t, err)
	})
	t.Run("wait timer tick", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		env.StartTimers()
		env.EventStateTransition()
		env.WaitTimerTick(43)
	})
}

func TestConsensusPostRequest(t *testing.T) {
	t.Run("post 1", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1, true)
		err := env.WaitMempool(1, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests post randomized", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 100 requests", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(100)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(100, 3, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 100 requests randomized", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, 4, 3, false)
		defer env.Log.Sync()
		env.StartTimers()
		env.EventStateTransition()
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
		env, _ := consensus.NewMockedEnv(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1)
		err := env.WaitMempool(1, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(1, true)
		time.Sleep(500 * time.Millisecond)
		err := env.WaitMempool(1, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10)
		err := env.WaitMempool(10, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
	t.Run("post 10 requests randomized", func(t *testing.T) {
		env, _ := consensus.NewMockedEnv(t, numNodes, quorum, false)
		defer env.Log.Sync()

		env.StartTimers()
		env.EventStateTransition()
		env.PostDummyRequests(10, true)
		err := env.WaitMempool(10, quorum, waitMempoolTimeout)
		require.NoError(t, err)
	})
}
