package domain_test

import (
	"sync"
	"testing"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
)

func TestDomainProvider(t *testing.T) {
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
	d, err := nodes[1].PeerDomain(netIDs)
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

	nodeCount := 5
	netIDs, pubKeys, privKeys := testpeers.SetupKeys(uint16(nodeCount), suite)
	nodes := testpeers.SetupNet(netIDs, pubKeys, privKeys, testutil.NewPeeringNetReliable(), log)
	for i := range nodes {
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
