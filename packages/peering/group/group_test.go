// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package group_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/testutil"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"
)

func TestGroupProvider(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()

	nodeCount := 4
	peeringURLs, nodeIdentities := testpeers.SetupKeys(uint16(nodeCount))
	nodes, netCloser := testpeers.SetupNet(peeringURLs, nodeIdentities, testutil.NewPeeringNetReliable(log), log)
	for i := range nodes {
		go nodes[i].Run(context.Background())
	}

	//
	// Listen for messages on all the nodes.
	peeringID := peering.RandomPeeringID()
	receiver := byte(4)
	doneCh1 := make(chan bool)
	doneCh2 := make(chan bool)
	doneCh3 := make(chan bool)
	nodes[1].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		doneCh1 <- true
	})
	nodes[2].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		doneCh2 <- true
	})
	nodes[3].Attach(&peeringID, receiver, func(recv *peering.PeerMessageIn) {
		doneCh3 <- true
	})
	//
	// Create a group on one of nodes.
	var g peering.GroupProvider
	g, err := nodes[0].PeerGroup(peeringID, testpeers.PublicKeys(nodeIdentities))
	require.NoError(t, err)
	//
	// Broadcast a message and wait until it will be received on all the nodes.
	g.SendMsgBroadcast(receiver, 125, []byte{})
	<-doneCh1
	<-doneCh2
	<-doneCh3
	//
	// Done.
	g.Close()
	require.NoError(t, netCloser.Close())
}
