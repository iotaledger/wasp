package group_test

import (
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
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
	var err0, err1, err2 error
	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019"}
	nodes := make([]peering.NetworkProvider, len(netIDs))
	nodes[0], err0 = udp.NewNetworkProvider(netIDs[0], 9017, key.NewKeyPair(suite), suite, log.Named("node0"))
	nodes[1], err1 = udp.NewNetworkProvider(netIDs[1], 9018, key.NewKeyPair(suite), suite, log.Named("node1"))
	nodes[2], err2 = udp.NewNetworkProvider(netIDs[2], 9019, key.NewKeyPair(suite), suite, log.Named("node2"))
	require.Nil(t, err0)
	require.Nil(t, err1)
	require.Nil(t, err2)
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
	g, err1 = nodes[1].Group(netIDs)
	require.Nil(t, err1)
	//
	// Broadcast a message and wait until it will be received on all the nodes.
	g.Broadcast(&peering.PeerMessage{ChainID: *coretypes.NewRandomChainID(), MsgType: 125}, true)
	<-doneCh0
	<-doneCh1
	<-doneCh2
	//
	// Done.
	g.Close()
}
