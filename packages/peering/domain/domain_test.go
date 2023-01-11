// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package domain_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

func TestDomainProvider(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	nodeCount := 3
	netIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	nodes, netCloser := testpeers.SetupNet(netIDs, nodeIdentities, testutil.NewPeeringNetReliable(log), log)
	nodePubKeys := testpeers.PublicKeys(nodeIdentities)
	for i := range nodes {
		go nodes[i].Run(context.Background())
	}

	//
	// Listen for messages on all the nodes.
	peeringID := peering.RandomPeeringID()
	receiver := byte(16)
	doneCh0 := make(chan bool)
	doneCh1 := make(chan bool)
	doneCh2 := make(chan bool)
	nodes[0].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		t.Log("0 received")
		doneCh0 <- true
	})
	nodes[1].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		t.Log("1 received")
		doneCh1 <- true
	})
	nodes[2].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		t.Log("2 received")
		doneCh2 <- true
	})
	//
	// Create a group on one of nodes.
	var d peering.PeerDomainProvider
	d, err := nodes[1].PeerDomain(peeringID, nodePubKeys)
	require.Nil(t, err)
	require.NotNil(t, d)

	d.SendMsgByPubKey(nodePubKeys[0], receiver, 125, []byte{})
	d.SendMsgByPubKey(nodePubKeys[2], receiver, 125, []byte{})
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
	nodePubKeys := testpeers.PublicKeys(nodeIdentities)
	for i := range nodes {
		go nodes[i].Run(context.Background())
	}
	peeringID := peering.RandomPeeringID()

	// Create a group on 2 of nodes.
	d1, err := nodes[1].PeerDomain(peeringID, nodePubKeys)
	require.NoError(t, err)
	require.NotNil(t, d1)

	d2, err := nodes[2].PeerDomain(peeringID, nodePubKeys)
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
			if nodePubKeys[1].Equals(recv.SenderPubKey) {
				r1++
			}
			if nodePubKeys[2].Equals(recv.SenderPubKey) {
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
		for _, pubKey := range d1.GetRandomOtherPeers(sendTo) {
			d1.SendMsgByPubKey(pubKey, receiver, 125, []byte{})
		}
		for _, pubKey := range d2.GetRandomOtherPeers(sendTo) {
			d2.SendMsgByPubKey(pubKey, receiver, 125, []byte{})
		}
		wg.Wait()
	}
	require.EqualValues(t, sendTo*5, r1)
	require.EqualValues(t, sendTo*5, r2)
	d1.Close()
	d2.Close()
	require.NoError(t, netCloser.Close())
}

func TestGetRandomOtherPeers(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	nodeCount := 8 // 7 excluding self
	peersToGet := 5
	iterationCount := 13
	netIDs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	nodes, netCloser := testpeers.SetupNet(netIDs, nodeIdentities, testutil.NewPeeringNetReliable(log), log)
	nodePubKeys := testpeers.PublicKeys(nodeIdentities)
	peeringID := peering.RandomPeeringID()

	domain, err := nodes[0].PeerDomain(peeringID, nodePubKeys)
	require.NoError(t, err)
	require.NotNil(t, domain)

	for i := 0; i < iterationCount; i++ {
		t.Logf("Iteration %v...", i)
		peers := domain.GetRandomOtherPeers(peersToGet)
		require.Equal(t, peersToGet, len(peers))
		for j := range peers {
			t.Logf("\tComparing peers %v, key %v...", j, peers[j].String())
			for k := range peers[j+1:] {
				kk := k + j + 1
				t.Logf("\t\t and %v, key %v", kk, peers[kk].String())
				require.False(t, peers[j].Equals(peers[kk]))
			}
		}
	}

	require.NoError(t, netCloser.Close())
}
