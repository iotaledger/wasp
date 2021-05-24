// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package commonsubset_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/chain/consensus/commonsubset"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	var peerCount uint16 = 4 // 10
	var threshold uint16 = 3 // 7
	var suite = pairing.NewSuiteBn256()
	var peeringID = peering.RandomPeeringID()
	peerNetIDs, peerPubs, peerSecs := setupKeys(peerCount, suite)
	networkProviders := setupNet(
		t, peerNetIDs, peerPubs, peerSecs, testutil.NewPeeringNetReliable(),
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
		acsPeers[a], err = commonsubset.NewCommonSubset(0, peeringID, networkProviders[a], group, threshold, newFakeCoin(), nil, log)
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
	var suite = pairing.NewSuiteBn256()
	var peeringID = peering.RandomPeeringID()
	peerNetIDs, peerPubs, peerSecs := setupKeys(peerCount, suite)
	netLogger := testlogger.WithLevel(log.Named("Network"), logger.LevelInfo, false)
	netBehavior := testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 100*time.Millisecond, netLogger)
	networkProviders := setupNet(t, peerNetIDs, peerPubs, peerSecs, netBehavior, netLogger)
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
		acsPeers[a], err = commonsubset.NewCommonSubset(0, peeringID, networkProviders[a], group, threshold, newFakeCoin(), nil, log)
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
	var output []map[uint16][]byte = make([]map[uint16][]byte, peerCount)
	var outputWG = &sync.WaitGroup{}
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

func setupKeys(peerCount uint16, suite *pairing.SuiteBn256) ([]string, []kyber.Point, []kyber.Scalar) {
	var peerNetIDs []string = make([]string, peerCount)
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerNetIDs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerNetIDs))
	for i := range peerNetIDs {
		peerPair := key.NewKeyPair(suite)
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	return peerNetIDs, peerPubs, peerSecs
}

// A helper for testcases.
func setupNet(
	t *testing.T,
	peerNetIDs []string,
	peerPubs []kyber.Point,
	peerSecs []kyber.Scalar,
	behavior testutil.PeeringNetBehavior,
	log *logger.Logger,
) []peering.NetworkProvider {
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000, behavior,
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	return peeringNetwork.NetworkProviders()
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
