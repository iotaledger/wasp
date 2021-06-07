package testpeers

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry_pkg"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func SetupKeys(peerCount uint16, suite *pairing.SuiteBn256) ([]string, []kyber.Point, []kyber.Scalar) {
	peerNetIDs := make([]string, peerCount)
	peerPubs := make([]kyber.Point, len(peerNetIDs))
	peerSecs := make([]kyber.Scalar, len(peerNetIDs))
	for i := range peerNetIDs {
		peerPair := key.NewKeyPair(suite)
		peerNetIDs[i] = fmt.Sprintf("P%02d", i)
		peerSecs[i] = peerPair.Private
		peerPubs[i] = peerPair.Public
	}
	return peerNetIDs, peerPubs, peerSecs
}

func SetupDkg(
	t *testing.T,
	threshold uint16,
	peerNetIDs []string,
	peerPubs []kyber.Point,
	peerSecs []kyber.Scalar,
	suite *pairing.SuiteBn256,
	log *logger.Logger,
) (ledgerstate.Address, []registry_pkg.DKShareRegistryProvider) {
	timeout := 100 * time.Second
	networkProviders := SetupNet(peerNetIDs, peerPubs, peerSecs, testutil.NewPeeringNetReliable(), log)
	//
	// Initialize the DKG subsystem in each node.
	dkgNodes := make([]*dkg.Node, len(peerNetIDs))
	registries := make([]registry_pkg.DKShareRegistryProvider, len(peerNetIDs))
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

func SetupNet(
	peerNetIDs []string,
	peerPubs []kyber.Point,
	peerSecs []kyber.Scalar,
	behavior testutil.PeeringNetBehavior,
	log *logger.Logger,
) []peering.NetworkProvider {
	peeringNetwork := testutil.NewPeeringNetwork(
		peerNetIDs, peerPubs, peerSecs, 10000, behavior,
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	return peeringNetwork.NetworkProviders()
}
