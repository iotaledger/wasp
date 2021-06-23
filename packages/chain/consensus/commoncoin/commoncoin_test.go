// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package commoncoin_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
)

func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testCC(t, testutil.NewPeeringNetReliable(), log)
}

func TestUnreliableNet(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	netBehavior := testutil.NewPeeringNetUnreliable( // NOTE: Network parameters.
		80,                                         // Delivered %
		20,                                         // Duplicated %
		10*time.Millisecond, 1000*time.Millisecond, // Delays (from, till)
		testlogger.WithLevel(log.Named("UnreliableNet"), logger.LevelWarn, false),
	)
	testCC(t, netBehavior, log)
}

func testCC(t *testing.T, netBehavior testutil.PeeringNetBehavior, log *logger.Logger) {
	var peerCount uint16 = 10
	var threshold uint16 = 7
	suite := pairing.NewSuiteBn256()
	peeringID := peering.RandomPeeringID()
	peerNetIDs, peerIdentities := testpeers.SetupKeys(peerCount)
	address, nodeRegistries := testpeers.SetupDkgPregenerated(t, threshold, peerNetIDs, suite)
	networkProviders := testpeers.SetupNet(peerNetIDs, peerIdentities, netBehavior, log)
	ccNodes := setupCommonCoinNodes(peeringID, address, peerNetIDs, nodeRegistries, networkProviders, log)
	//
	// Check, if the common coin algorithm works.
	wg := sync.WaitGroup{}
	wg.Add(len(peerNetIDs))
	ccResults := make([][][]byte, 10) // [attempt][node]Coin
	for attempt := range ccResults {
		ccResults[attempt] = make([][]byte, len(peerNetIDs))
	}
	ccDLock := &sync.RWMutex{}
	ccDuration := make([]time.Duration, 0)
	for i := range peerNetIDs {
		ii := i
		go func() {
			var ccErr error
			cc := make([]byte, 0)
			for attempt := 0; attempt < 10; attempt++ {
				start := time.Now()
				cc, ccErr = ccNodes[ii].GetCoin(cc)
				require.Nil(t, ccErr)
				require.NotNil(t, cc)
				ccResults[attempt][ii] = cc
				ccDLock.Lock()
				ccDuration = append(ccDuration, time.Since(start))
				ccDLock.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	//
	// Print duration.
	ccDAwg := 0 * time.Millisecond
	for i := range ccDuration {
		ccDAwg += ccDuration[i]
	}
	ccDAwg = time.Duration((int64(ccDAwg/time.Nanosecond) / int64(len(ccDuration)))) * time.Nanosecond
	t.Logf("Average duration: %v", ccDAwg)
	//
	// Validate results.
	for attempt := range ccResults {
		for ii := range peerNetIDs {
			require.NotNil(t, ccResults[attempt][ii])
			require.True(t, bytes.Equal(ccResults[attempt][0], ccResults[attempt][ii]))
		}
	}
	for i := range peerNetIDs {
		require.Nil(t, ccNodes[i].Close())
	}
}

func setupCommonCoinNodes(
	peeringID peering.PeeringID,
	address ledgerstate.Address,
	peerNetIDs []string,
	nodeRegistries []coretypes.DKShareRegistryProvider,
	networkProviders []peering.NetworkProvider,
	log *logger.Logger,
) []commoncoin.Provider {
	var ccNodes []commoncoin.Provider = make([]commoncoin.Provider, len(peerNetIDs))
	for i := range peerNetIDs {
		peerDKShare, _ := nodeRegistries[i].LoadDKShare(address)
		peerNetGroup, _ := networkProviders[i].PeerGroup(peerNetIDs)
		ccNodes[i] = commoncoin.NewCommonCoinNode(
			nil, peerDKShare, peeringID, peerNetGroup,
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
		)
	}
	return ccNodes
}
