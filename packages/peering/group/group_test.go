package group_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/testutil/testlogger"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestGroupProvider(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()

	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019"}
	nodes := make([]peering.NetworkProvider, len(netIDs))

	cfg0, err := peering.NewPeerNetworkConfig(netIDs[0], 9017, netIDs...)
	require.NoError(t, err)
	nodes[0], err = udp.NewNetworkProvider(cfg0, key.NewKeyPair(suite), suite, log.Named("node0"))
	require.NoError(t, err)
	cfg1, err := peering.NewPeerNetworkConfig(netIDs[1], 9018, netIDs...)
	require.NoError(t, err)
	nodes[1], err = udp.NewNetworkProvider(cfg1, key.NewKeyPair(suite), suite, log.Named("node1"))
	require.NoError(t, err)
	cfg2, err := peering.NewPeerNetworkConfig(netIDs[2], 9019, netIDs...)
	require.NoError(t, err)
	nodes[2], err = udp.NewNetworkProvider(cfg2, key.NewKeyPair(suite), suite, log.Named("node2"))
	require.NoError(t, err)
	for i := range nodes {
		go nodes[i].Run(make(<-chan struct{}))
	}

	//
	// Listen for messages on all the nodes.
	doneCh0 := make(chan bool)
	doneCh1 := make(chan bool)
	doneCh2 := make(chan bool)
	nodes[0].Attach(nil, func(recv *peering.RecvEvent) {
		doneCh0 <- true
	})
	nodes[1].Attach(nil, func(recv *peering.RecvEvent) {
		doneCh1 <- true
	})
	nodes[2].Attach(nil, func(recv *peering.RecvEvent) {
		doneCh2 <- true
	})
	//
	// Create a group on one of nodes.
	var g peering.GroupProvider
	g, err = nodes[1].PeerGroup(netIDs)
	require.Nil(t, err)
	//
	// Broadcast a message and wait until it will be received on all the nodes.
	g.Broadcast(&peering.PeerMessage{PeeringID: peering.RandomPeeringID(), MsgType: 125}, true)
	<-doneCh0
	<-doneCh1
	<-doneCh2
	//
	// Done.
	g.Close()
}
