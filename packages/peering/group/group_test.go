package group_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
)

func TestGroupProvider(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()

	nodeCount := 3
	netIDs, pubKeys, privKeys := testpeers.SetupKeys(uint16(nodeCount), suite)
	nodes := testpeers.SetupNet(netIDs, pubKeys, privKeys, testutil.NewPeeringNetReliable(), log)
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
	g, err := nodes[1].PeerGroup(netIDs)
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
