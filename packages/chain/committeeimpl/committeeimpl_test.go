package committeeimpl

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
)

func TestCommitteeBasic(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	nodeCount := 4
	netIDs, pubKeys, privKeys := testpeers.SetupKeys(uint16(nodeCount), suite)
	stateAddr, dksRegistries := testpeers.SetupDkg(t, uint16((len(netIDs)*2)/3+1), netIDs, pubKeys, privKeys, suite, log.Named("dkg"))
	nodes := testpeers.SetupNet(netIDs, pubKeys, privKeys, testutil.NewPeeringNetReliable(), log)
	net0 := nodes[0]
	reg := testchain.NewMockedCommitteeRegistry(netIDs)
	cfg0 := &committeeimplTestConfigProvider{
		ownNetID:  netIDs[0],
		neighbors: netIDs,
	}

	c, err := NewCommittee(stateAddr, nil, net0, cfg0, dksRegistries[0], reg, log)
	require.NoError(t, err)
	require.True(t, c.Address().Equals(stateAddr))
	require.EqualValues(t, 4, c.Size())
	require.EqualValues(t, 3, c.Quorum())

	time.Sleep(100 * time.Millisecond)
	require.True(t, c.IsReady())
	c.Close()
	require.False(t, c.IsReady())
}

var _ coretypes.PeerNetworkConfigProvider = &committeeimplTestConfigProvider{}

// TODO: should this object be obtained from peering.NetworkProvider?
// Or should coretypes.PeerNetworkConfigProvider methods methods be part of
// peering.NetworkProvider interface
type committeeimplTestConfigProvider struct {
	ownNetID  string
	neighbors []string
}

func (p *committeeimplTestConfigProvider) OwnNetID() string {
	return p.ownNetID
}

func (p *committeeimplTestConfigProvider) PeeringPort() int {
	return 0 // Anything
}

func (p *committeeimplTestConfigProvider) Neighbors() []string {
	return p.neighbors
}

func (p *committeeimplTestConfigProvider) String() string {
	return fmt.Sprintf("committeeimplPeerConfig( ownNetID: %s, neighbors: %+v )", p.OwnNetID(), p.Neighbors())
}
