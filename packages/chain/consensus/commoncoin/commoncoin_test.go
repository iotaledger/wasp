package commoncoin_test

import (
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/testutil/testpeers"

	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"go.dedis.ch/kyber/v3/pairing"
)

func TestBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log := testlogger.NewLogger(t)
	defer log.Sync()
	var peerCount uint16 = 10
	var threshold uint16 = 7
	suite := pairing.NewSuiteBn256()
	peeringID := peering.RandomPeeringID()
	peerNetIDs, peerPubs, peerSecs := testpeers.SetupKeys(peerCount, suite)
	address, nodeRegistries := testpeers.SetupDkg(t, threshold, peerNetIDs, peerPubs, peerSecs, suite, log)
	networkProviders := testpeers.SetupNet(peerNetIDs, peerPubs, peerSecs, testutil.NewPeeringNetReliable(), log)
	ccNodes := setupCommonCoinNodes(peeringID, address, peerNetIDs, nodeRegistries, networkProviders, log)
	//
	// Check, if the common coin algorithm works.
	wg := sync.WaitGroup{}
	wg.Add(len(peerNetIDs))
	for i := range peerNetIDs {
		ii := i
		go func() {
			log.Infof("CC[0,%v], Asking for a common coin.", ii)
			var ccErr error
			cc := make([]byte, 0)
			for attempt := 0; attempt < 10; attempt++ {
				start := time.Now()
				cc, ccErr = ccNodes[ii].GetCoin(cc)
				require.Nil(t, ccErr)
				require.NotNil(t, cc)
				log.Infof("CC[%v,%v]: %+v in %v", attempt, ii, base58.Encode(cc), time.Since(start))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := range peerNetIDs {
		require.Nil(t, ccNodes[i].Close())
	}
}

func TestUnreliableNet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log := testlogger.NewLogger(t)
	defer log.Sync()
	var peerCount uint16 = 10
	var threshold uint16 = 7
	suite := pairing.NewSuiteBn256()
	peeringID := peering.RandomPeeringID()
	netBehavior := testutil.NewPeeringNetUnreliable( // NOTE: Network parameters.
		80,                                         // Delivered %
		20,                                         // Duplicated %
		10*time.Millisecond, 1000*time.Millisecond, // Delays (from, till)
		testlogger.WithLevel(log.Named("UnreliableNet"), logger.LevelDebug, false),
	)
	peerNetIDs, peerPubs, peerSecs := testpeers.SetupKeys(peerCount, suite)
	address, nodeRegistries := testpeers.SetupDkg(t, threshold, peerNetIDs, peerPubs, peerSecs, suite, log)
	networkProviders := testpeers.SetupNet(peerNetIDs, peerPubs, peerSecs, netBehavior, log)
	ccNodes := setupCommonCoinNodes(peeringID, address, peerNetIDs, nodeRegistries, networkProviders, log)
	//
	// Check, if the common coin algorithm works.
	wg := sync.WaitGroup{}
	wg.Add(len(peerNetIDs))
	for i := range peerNetIDs {
		ii := i
		go func() {
			log.Infof("CC[0,%v], Asking for a common coin.", ii)
			var ccErr error
			cc := make([]byte, 0)
			for attempt := 0; attempt < 10; attempt++ {
				start := time.Now()
				cc, ccErr = ccNodes[ii].GetCoin(cc)
				require.Nil(t, ccErr)
				require.NotNil(t, cc)
				log.Infof("CC[%v,%v]: %+v in %v", attempt, ii, base58.Encode(cc), time.Since(start))
			}
			wg.Done()
		}()
	}
	wg.Wait()
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
