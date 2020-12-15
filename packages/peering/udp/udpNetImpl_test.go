// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestUDPPeeringImpl(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testutil.NewLogger(t)
	defer log.Sync()
	var err0, err1, err2 error
	doneCh := make(chan bool)
	chain1 := coretypes.NewRandomChainID()
	chain2 := coretypes.NewRandomChainID()
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

	n0p2, err0 := nodes[0].PeerByNetID(netIDs[2])
	n1p1, err1 := nodes[1].PeerByNetID(netIDs[1])
	n2p0, err2 := nodes[2].PeerByNetID(netIDs[0])
	require.Nil(t, err0)
	require.Nil(t, err1)
	require.Nil(t, err2)

	nodes[0].Attach(nil, func(recv *peering.RecvEvent) {
		doneCh <- true
	})

	n0p2.SendMsg(&peering.PeerMessage{ChainID: chain1, MsgType: 125})
	n1p1.SendMsg(&peering.PeerMessage{ChainID: chain1, MsgType: 125})
	n2p0.SendMsg(&peering.PeerMessage{ChainID: chain2, MsgType: 125})

	<-doneCh
}
