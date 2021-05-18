// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/testutil/testlogger"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestUDPPeeringImpl(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()

	doneCh := make(chan bool)
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

	n0p2, err := nodes[0].PeerByNetID(netIDs[2])
	require.NoError(t, err)
	n1p1, err := nodes[1].PeerByNetID(netIDs[1])
	require.NoError(t, err)
	n2p0, err := nodes[2].PeerByNetID(netIDs[0])
	require.NoError(t, err)

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
