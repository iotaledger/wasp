package committeeimpl

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain/mock_chain"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestCommitteeBasic(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019", "localhost:9020"}

	reg := mock_chain.NewMockedRegistry(4, 3, 0, netIDs)
	cfg0, err := peering.NewStaticPeerNetworkConfigProvider(netIDs[0], 9017, netIDs...)
	require.NoError(t, err)
	net0, err := udp.NewNetworkProvider(cfg0, key.NewKeyPair(suite), suite, log.Named("net0"))
	require.NoError(t, err)

	keyPair := ed25519.GenerateKeyPair()
	stateAddr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	c, err := NewCommittee(stateAddr, net0, cfg0, reg, reg, log)
	require.NoError(t, err)
	require.True(t, c.Address().Equals(stateAddr))
	require.EqualValues(t, 4, c.Size())
	require.EqualValues(t, 3, c.Quorum())

	time.Sleep(100 * time.Millisecond)
	require.True(t, c.IsReady())
	c.Close()
	require.False(t, c.IsReady())
}
