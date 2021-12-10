package domain_test

import (
	"sync"
	"testing"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
)

func TestDomainProvider(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	nodeCount := 3
	netIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	nodes, netCloser := testpeers.SetupNet(netIDs, nodeIdentities, testutil.NewPeeringNetReliable(log), log)
	for i := range nodes {
		go nodes[i].Run(make(<-chan struct{}))
	}

	//
	// Listen for messages on all the nodes.
	peeringID := peering.RandomPeeringID()
	receiver := byte(16)
	doneCh0 := make(chan bool)
	doneCh1 := make(chan bool)
	doneCh2 := make(chan bool)
	nodes[0].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		t.Logf("0 received")
		doneCh0 <- true
	})
	nodes[1].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		t.Logf("1 received")
		doneCh1 <- true
	})
	nodes[2].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		t.Logf("2 received")
		doneCh2 <- true
	})
	//
	// Create a group on one of nodes.
	var d peering.PeerDomainProvider
	d, err := nodes[1].PeerDomain(peeringID, netIDs)
	require.Nil(t, err)
	require.NotNil(t, d)

	d.SendMsgByNetID(netIDs[0], receiver, 125, []byte{})
	d.SendMsgByNetID(netIDs[2], receiver, 125, []byte{})
	<-doneCh0
	<-doneCh2
	//
	// Done.
	d.Close()
	require.NoError(t, netCloser.Close())
}

func TestRandom(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	nodeCount := 5
	netIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	nodes, netCloser := testpeers.SetupNet(netIDs, nodeIdentities, testutil.NewPeeringNetReliable(log), log)
	for i := range nodes {
		go nodes[i].Run(make(<-chan struct{}))
	}
	peeringID := peering.RandomPeeringID()

	// Create a group on 2 of nodes.
	d1, err := nodes[1].PeerDomain(peeringID, netIDs)
	require.NoError(t, err)
	require.NotNil(t, d1)

	d2, err := nodes[2].PeerDomain(peeringID, netIDs)
	require.NoError(t, err)
	require.NotNil(t, d1)

	//
	// Listen for messages on all the nodes.
	var wg sync.WaitGroup
	var r1, r2 int
	receiver := byte(8)
	for i := range nodes {
		ii := i
		nodes[i].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
			t.Logf("%d received", ii)
			if netIDs[1] == recv.SenderNetID {
				r1++
			}
			if netIDs[2] == recv.SenderNetID {
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
		for _, netID := range d1.GetRandomPeers(sendTo) {
			d1.SendMsgByNetID(netID, receiver, 125, []byte{})
		}
		for _, netID := range d2.GetRandomPeers(sendTo) {
			d2.SendMsgByNetID(netID, receiver, 125, []byte{})
		}
		wg.Wait()
	}
	require.EqualValues(t, sendTo*5, r1)
	require.EqualValues(t, sendTo*5, r2)
	d1.Close()
	d2.Close()
	require.NoError(t, netCloser.Close())
}
