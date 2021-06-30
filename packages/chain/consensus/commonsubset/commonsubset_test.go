// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package commonsubset_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/chain/consensus/commonsubset"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Run("N=4/T=3", func(tt *testing.T) { testBasic(tt, 4, 3) })
	t.Run("N=7/T=5", func(tt *testing.T) { testBasic(tt, 7, 5) })
	t.Run("N=10/T=7", func(tt *testing.T) { testBasic(tt, 10, 7) })
}

func testBasic(t *testing.T, peerCount, threshold uint16) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	peeringID := peering.RandomPeeringID()
	peerNetIDs, peerIdentities := testpeers.SetupKeys(peerCount)
	networkProviders := testpeers.SetupNet(
		peerNetIDs, peerIdentities, testutil.NewPeeringNetReliable(),
		testlogger.WithLevel(log.Named("Network"), logger.LevelDebug, false),
	)
	t.Logf("Network created.")

	acsPeers := make([]*commonsubset.CommonSubset, peerCount)
	for i := range acsPeers {
		ii := i // Use a local copy in the callback.
		networkProviders[ii].Attach(&peeringID, func(recv *peering.RecvEvent) {
			if acsPeers[ii] != nil {
				require.True(t, acsPeers[ii].TryHandleMessage(recv))
			}
		})
	}
	for a := range acsPeers {
		group, err := networkProviders[a].PeerGroup(peerNetIDs)
		require.Nil(t, err)
		acsPeers[a], err = commonsubset.NewCommonSubset(0, 0, peeringID, group, threshold, newFakeCoin(), nil, log)
		require.Nil(t, err)
	}
	t.Logf("ACS Nodes created.")
	for a := range acsPeers {
		input := []byte(peerNetIDs[a])
		acsPeers[a].Input(input)
	}
	t.Logf("ACS Inputs sent.")

	for a := range acsPeers {
		out := <-acsPeers[a].OutputCh()
		t.Logf("ACS[%v] Output received: %+v", a, out)
	}
	t.Logf("ACS Nodes all decided.")

	for a := range acsPeers {
		acsPeers[a].Close()
	}
	t.Logf("ACS Nodes closed.")
}

func TestRandomized(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	var peerCount uint16 = 10
	var threshold uint16 = 7
	peeringID := peering.RandomPeeringID()
	peerNetIDs, peerIdentities := testpeers.SetupKeys(peerCount)
	netLogger := testlogger.WithLevel(log.Named("Network"), logger.LevelInfo, false)
	netBehavior := testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 100*time.Millisecond, netLogger)
	networkProviders := testpeers.SetupNet(peerNetIDs, peerIdentities, netBehavior, netLogger)
	t.Logf("Network created.")

	acsPeers := make([]*commonsubset.CommonSubset, peerCount)
	for i := range acsPeers {
		ii := i // Use a local copy in the callback.
		networkProviders[ii].Attach(&peeringID, func(recv *peering.RecvEvent) {
			if acsPeers[ii] != nil {
				require.True(t, acsPeers[ii].TryHandleMessage(recv))
			}
		})
	}
	for a := range acsPeers {
		group, err := networkProviders[a].PeerGroup(peerNetIDs)
		require.Nil(t, err)
		acsPeers[a], err = commonsubset.NewCommonSubset(0, 0, peeringID, group, threshold, newFakeCoin(), nil, log)
		require.Nil(t, err)
	}
	t.Logf("ACS Nodes created.")
	for a := range acsPeers {
		input := []byte(peerNetIDs[a])
		acsPeers[a].Input(input)
	}
	t.Logf("ACS Inputs sent.")

	//
	// Async wait here is for debugging only.
	output := make([]map[uint16][]byte, peerCount)
	outputWG := &sync.WaitGroup{}
	outputWG.Add(int(peerCount))
	for a := range acsPeers {
		aa := a
		go func() {
			outCh := acsPeers[aa].OutputCh()
			timerCh := time.After(15 * time.Second)
			for {
				select {
				case output[aa] = <-outCh:
					t.Logf("ACS[%v] Output received: %+v", aa, output[aa])
					outputWG.Done()
					return
				case <-timerCh:
					t.Logf("ACS[%v] Info: %+v", aa, acsPeers[aa])
					timerCh = time.After(15 * time.Second)
				}
			}
		}()
	}
	outputWG.Wait()
	t.Logf("ACS Nodes all decided.")
	for a := range acsPeers {
		acsPeers[a].Close()
	}
	t.Logf("ACS Nodes closed.")
	for a := range acsPeers {
		require.Equal(t, len(output[0]), len(output[a]))
		for i := range output[a] {
			require.Equal(t, 0, bytes.Compare(output[0][i], output[a][i]))
		}
	}
}

func TestCoordinator(t *testing.T) {
	t.Run("N=1/T=1", func(tt *testing.T) { testCoordinator(tt, 1, 1) })
	t.Run("N=4/T=3", func(tt *testing.T) { testCoordinator(tt, 4, 3) })
	t.Run("N=7/T=5", func(tt *testing.T) { testCoordinator(tt, 7, 5) })
	t.Run("N=10/T=7", func(tt *testing.T) { testCoordinator(tt, 10, 7) })
}

func testCoordinator(t *testing.T, peerCount, threshold uint16) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	peeringID := peering.RandomPeeringID()
	peerNetIDs, peerIdentities := testpeers.SetupKeys(peerCount)
	networkProviders := testpeers.SetupNet(
		peerNetIDs, peerIdentities, testutil.NewPeeringNetReliable(),
		testlogger.WithLevel(log.Named("Network"), logger.LevelDebug, false),
	)
	t.Logf("Network created.")

	acsCoords := make([]*commonsubset.CommonSubsetCoordinator, peerCount)
	for i := range acsCoords {
		ii := i // Use a local copy in the callback.
		group, err := networkProviders[i].PeerGroup(peerNetIDs)
		require.Nil(t, err)
		acsCoords[i] = commonsubset.NewCommonSubsetCoordinator(peeringID, networkProviders[i], group, threshold, newFakeCoin(), log)
		networkProviders[ii].Attach(&peeringID, func(recv *peering.RecvEvent) {
			if acsCoords[ii] != nil {
				require.True(t, acsCoords[ii].TryHandleMessage(recv))
			}
		})
	}
	t.Logf("ACS Nodes created.")

	sessionID := uint64(21695645984168)
	results := make([][][]byte, peerCount)
	resultsWG := &sync.WaitGroup{}
	resultsWG.Add(int(peerCount))
	for i := range acsCoords {
		ii := i
		input := []byte(peerNetIDs[i])
		acsCoords[i].RunACSConsensus(input, sessionID, 1, func(sid uint64, res [][]byte) {
			results[ii] = res
			resultsWG.Done()
		})
	}
	resultsWG.Wait()
	t.Logf("ACS Nodes all decided.")
	for i := range results {
		for j := range results[i] {
			require.True(t, bytes.Equal(results[i][j], results[0][j]))
		}
	}
	for i := range acsCoords {
		acsCoords[i].Close()
	}
	t.Logf("ACS Nodes closed.")
}

// fakeCoin is a trivial incorrect implementation of the common coin interface.
// It is used here just to avoid dependencies to particular crypto libraries.
type fakeCoin struct{}

func newFakeCoin() commoncoin.Provider {
	return &fakeCoin{}
}

func (fc *fakeCoin) GetCoin(seed []byte) ([]byte, error) {
	return seed, nil
}

func (fc *fakeCoin) FlipCoin(epoch uint32) bool {
	return (epoch/2)%2 == 0
}

func (fc *fakeCoin) Close() error {
	return nil
}

func (fc *fakeCoin) TryHandleMessage(recv *peering.RecvEvent) bool {
	return false
}
