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
	d, err1 = nodes[1].Domain(netIDs)
	require.Nil(t, err1)
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
		nodes[i], err = udp.NewNetworkProvider(netIDs[i], 9000+i, key.NewKeyPair(suite), suite, log.Named("node0"))
		require.NoError(t, err)
		go nodes[i].Run(make(<-chan struct{}))
	}

	// Create a group on 2 of nodes.
	d1, err := nodes[1].Domain(netIDs)
	require.NoError(t, err)
	require.NotNil(t, d1)

	d2, err := nodes[2].Domain(netIDs)
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
