package commoncoin_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	var peerCount uint16 = 10
	var threshold uint16 = 7
	var suite = pairing.NewSuiteBn256()
	var peeringID = peering.RandomPeeringID()
	peerNetIDs, peerPubs, peerSecs := setupKeys(peerCount, suite)
	address, nodeRegistries := setupDkg(t, threshold, peerNetIDs, peerPubs, peerSecs, suite, log)
	networkProviders := setupNet(t, peerNetIDs, peerPubs, peerSecs, testutil.NewPeeringNetReliable(), log)
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
	var suite = pairing.NewSuiteBn256()
	var peeringID = peering.RandomPeeringID()
	netBehavior := testutil.NewPeeringNetUnreliable( // NOTE: Network parameters.
		80,                                         // Delivered %
		20,                                         // Duplicated %
		10*time.Millisecond, 1000*time.Millisecond, // Delays (from, till)
		testlogger.WithLevel(log.Named("UnreliableNet"), logger.LevelDebug, false),
	)
	peerNetIDs, peerPubs, peerSecs := setupKeys(peerCount, suite)
	address, nodeRegistries := setupDkg(t, threshold, peerNetIDs, peerPubs, peerSecs, suite, log)
	networkProviders := setupNet(t, peerNetIDs, peerPubs, peerSecs, netBehavior, log)
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
	nodeRegistries []tcrypto.RegistryProvider,
	networkProviders []peering.NetworkProvider,
	log *logger.Logger,
) []commoncoin.Provider {
	var ccNodes []commoncoin.Provider = make([]commoncoin.Provider, len(peerNetIDs))
	for i := range peerNetIDs {
		peerDKShare, _ := nodeRegistries[i].LoadDKShare(address)
		peerNetGroup, _ := networkProviders[i].Group(peerNetIDs)
		ccNodes[i] = commoncoin.NewCommonCoinNode(
			nil, peerDKShare, peeringID, peerNetGroup,
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelDebug, false),
		)
	}
	return ccNodes
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
func setupDkg(
	t *testing.T,
	threshold uint16,
	peerNetIDs []string,
	peerPubs []kyber.Point,
	peerSecs []kyber.Scalar,
	suite *pairing.SuiteBn256,
	log *logger.Logger,
) (ledgerstate.Address, []tcrypto.RegistryProvider) {
	var timeout = 100 * time.Second
	var networkProviders = setupNet(t, peerNetIDs, peerPubs, peerSecs, testutil.NewPeeringNetReliable(), log)
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []*dkg.Node = make([]*dkg.Node, len(peerNetIDs))
	var registries []tcrypto.RegistryProvider = make([]tcrypto.RegistryProvider, len(peerNetIDs))
	for i := range peerNetIDs {
		registries[i] = testutil.NewDkgRegistryProvider(suite)
		dkgNodes[i] = dkg.NewNode(
			peerSecs[i], peerPubs[i], suite, networkProviders[i], registries[i],
			testlogger.WithLevel(log.With("NetID", peerNetIDs[i]), logger.LevelError, false),
		)
	}
	//
	// Initiate the key generation from some client node.
	dkShare, err := dkgNodes[0].GenerateDistributedKey(
		peerNetIDs,
		peerPubs,
		threshold,
		1*time.Second,
		2*time.Second,
		timeout,
	)
	require.Nil(t, err)
	require.NotNil(t, dkShare.Address)
	require.NotNil(t, dkShare.SharedPublic)
	return dkShare.Address, registries
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
