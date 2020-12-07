package tcp_test

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/tcp"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	log := testutil.NewLogger(t)
	defer log.Sync()
	var err0, err1, err2 error
	doneCh := make(chan bool)
	chain1 := coretypes.NewRandomChainID()
	chain2 := coretypes.NewRandomChainID()
	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019"}
	nodes := make([]peering.NetworkProvider, len(netIDs))
	nodes[0], err0 = tcp.NewNetworkProvider(netIDs[0], 9017, log.Named("node0"))
	nodes[1], err1 = tcp.NewNetworkProvider(netIDs[1], 9018, log.Named("node1"))
	nodes[2], err2 = tcp.NewNetworkProvider(netIDs[2], 9019, log.Named("node2"))
	require.Nil(t, err0)
	require.Nil(t, err1)
	require.Nil(t, err2)
	for i := range nodes {
		go nodes[i].Run(make(<-chan struct{}))
	}

	<-time.After(time.Second) // TODO: Temporary.

	n0p2, _ := nodes[0].PeerByLocation(netIDs[2])
	n1p1, _ := nodes[1].PeerByLocation(netIDs[1])
	n2p0, _ := nodes[2].PeerByLocation(netIDs[0])

	<-time.After(time.Second) // TODO: Temporary.

	nodes[0].Attach(nil, func(recv *peering.RecvEvent) {
		doneCh <- true
	})

	n0p2.SendMsg(&peering.PeerMessage{ChainID: chain1, MsgType: 125})
	n1p1.SendMsg(&peering.PeerMessage{ChainID: chain1, MsgType: 125})
	n2p0.SendMsg(&peering.PeerMessage{ChainID: chain2, MsgType: 125})

	<-doneCh
}
