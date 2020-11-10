package dkg_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

func TestSimple(t *testing.T) {
	//
	// Create a fake network and keys for the tests.
	var suite = edwards25519.NewBlakeSHA256Ed25519() //bn256.NewSuite()
	var peerLocs []string = []string{"a", "b", "c"}
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerLocs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerLocs))
	for i := range peerLocs {
		peerSecs[i] = suite.Scalar().Pick(suite.RandomStream())
		peerPubs[i] = suite.Point().Mul(peerSecs[i], nil)
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(peerLocs, peerPubs, peerSecs, 1000)
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	//
	// Initialize the DKG subsystem in each node.
	var dkgNodes []dkg.CoordNodeProvider = make([]dkg.CoordNodeProvider, len(peerLocs))
	for i := range peerLocs {
		dkgNodes[i] = dkg.InitNode(peerSecs[i], peerPubs[i], suite, networkProviders[i])
	}
	//
	// Initiate the key generation from some client node.
	var coordKey = suite.Scalar().Pick(suite.RandomStream())
	var coordPub = suite.Point().Mul(coordKey, nil)
	var coordNodeProvider dkg.CoordNodeProvider = testutil.NewDkgCoordNodeProvider(
		dkgNodes,
		time.Second,
	)
	c, err := dkg.GenerateDistributedKey(coordKey, coordPub, peerLocs, peerPubs, 1*time.Second, suite, coordNodeProvider)
	require.Nil(t, err)
	require.NotNil(t, c)
}
