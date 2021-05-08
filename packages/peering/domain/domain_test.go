package domain_test

import (
	"sync"
	"testing"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestDomainProvider(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019"}
	nodes := make([]peering.NetworkProvider, len(netIDs))

	cfg0, err := peering.NewStaticPeerNetworkConfigProvider(netIDs[0], 9017, netIDs...)
	require.NoError(t, err)
	nodes[0], err = udp.NewNetworkProvider(cfg0, key.NewKeyPair(suite), suite, log.Named("node0"))
	require.NoError(t, err)
	cfg1, err := peering.NewStaticPeerNetworkConfigProvider(netIDs[1], 9018, netIDs...)
	require.NoError(t, err)
	nodes[1], err = udp.NewNetworkProvider(cfg1, key.NewKeyPair(suite), suite, log.Named("node1"))
	require.NoError(t, err)
	cfg2, err := peering.NewStaticPeerNetworkConfigProvider(netIDs[2], 9019, netIDs...)
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
		t.Logf("0 received")
		doneCh0 <- true
	})
	nodes[1].Attach(nil, func(recv *peering.RecvEvent) {
		t.Logf("1 received")
		doneCh1 <- true
	})
	nodes[2].Attach(nil, func(recv *peering.RecvEvent) {
		t.Logf("2 received")
		doneCh2 <- true
	})
	//
	// Create a group on one of nodes.
	var d peering.PeerDomainProvider
	d, err = nodes[1].PeerDomain(netIDs)
	require.Nil(t, err)
	require.NotNil(t, d)

	msg := &peering.PeerMessage{PeeringID: peering.RandomPeeringID(), MsgType: 125}
	d.SendMsgByNetID(netIDs[0], msg)
	d.SendMsgByNetID(netIDs[2], msg)
	<-doneCh0
	<-doneCh2
	//
	// Done.
	d.Close()
}

func TestRandom(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()

	netIDs := []string{"localhost:9000", "localhost:9001", "localhost:9002", "localhost:9003", "localhost:9004"}
	nodes := make([]peering.NetworkProvider, len(netIDs))
	var err error
	for i := range nodes {
		cfg, err := peering.NewStaticPeerNetworkConfigProvider(netIDs[i], 9000+i, netIDs...)
		require.NoError(t, err)
		nodes[i], err = udp.NewNetworkProvider(cfg, key.NewKeyPair(suite), suite, log.Named("node0"))
		require.NoError(t, err)
		go nodes[i].Run(make(<-chan struct{}))
	}

	// Create a group on 2 of nodes.
	d1, err := nodes[1].PeerDomain(netIDs)
	require.NoError(t, err)
	require.NotNil(t, d1)

	d2, err := nodes[2].PeerDomain(netIDs)
	require.NoError(t, err)
	require.NotNil(t, d1)

	//
	// Listen for messages on all the nodes.
	var wg sync.WaitGroup
	var r1, r2 int
	for i := range nodes {
		ii := i
		nodes[i].Attach(nil, func(recv *peering.RecvEvent) {
			t.Logf("%d received", ii)
			if netIDs[1] == recv.From.NetID() {
				r1++
			}
			if netIDs[2] == recv.From.NetID() {
				r2++
			}
			wg.Done()
		})
	}
	//
	const sendTo = 2
	for i := 0; i < 5; i++ {
		wg.Add(sendTo * 2)
		t.Log("----------------------------------")
		msg := &peering.PeerMessage{PeeringID: peering.RandomPeeringID(), MsgType: 125}
		d1.SendMsgToRandomPeers(sendTo, msg)
		d2.SendMsgToRandomPeers(sendTo, msg)
		wg.Wait()
	}
	require.EqualValues(t, sendTo*5, r1)
	require.EqualValues(t, sendTo*5, r2)
	d1.Close()
	d2.Close()
}
