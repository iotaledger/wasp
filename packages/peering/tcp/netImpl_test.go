// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcp_test

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/tcp"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	var err0, err1, err2 error
	doneCh := make(chan bool)
	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019"}
	nodeIdentities := []ed25519.KeyPair{
		ed25519.GenerateKeyPair(),
		ed25519.GenerateKeyPair(),
		ed25519.GenerateKeyPair(),
	}
	nodes := make([]peering.NetworkProvider, len(netIDs))
	nodes[0], err0 = tcp.NewNetworkProvider(netIDs[0], 9017, &nodeIdentities[0], log.Named("node0"))
	nodes[1], err1 = tcp.NewNetworkProvider(netIDs[1], 9018, &nodeIdentities[1], log.Named("node1"))
	nodes[2], err2 = tcp.NewNetworkProvider(netIDs[2], 9019, &nodeIdentities[2], log.Named("node2"))
	require.Nil(t, err0)
	require.Nil(t, err1)
	require.Nil(t, err2)
	for i := range nodes {
		go nodes[i].Run(make(<-chan struct{}))
	}

	<-time.After(time.Second) // TODO: [KP] Temporary.

	n0p2, _ := nodes[0].PeerByNetID(netIDs[2])
	n1p1, _ := nodes[1].PeerByNetID(netIDs[1])
	n2p0, _ := nodes[2].PeerByNetID(netIDs[0])

	<-time.After(time.Second) // TODO: [KP] Temporary.

	nodes[0].Attach(nil, func(recv *peering.RecvEvent) {
		doneCh <- true
	})

	chain1 := peering.RandomPeeringID()
	chain2 := peering.RandomPeeringID()
	n0p2.SendMsg(&peering.PeerMessage{PeeringID: chain1, MsgType: 125})
	n1p1.SendMsg(&peering.PeerMessage{PeeringID: chain1, MsgType: 125})
	n2p0.SendMsg(&peering.PeerMessage{PeeringID: chain2, MsgType: 125})

	<-doneCh
}
