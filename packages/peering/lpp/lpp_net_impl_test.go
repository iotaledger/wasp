// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package lpp_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/lpp"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestLPPPeeringImpl(t *testing.T) {
	var err error
	log := testlogger.NewLogger(t)
	defer log.Shutdown()

	doneCh := make(chan bool)
	peeringURLs := []string{"localhost:9027", "localhost:9028", "localhost:9029"}
	nodes := make([]peering.NetworkProvider, len(peeringURLs))

	keys := make([]*cryptolib.KeyPair, len(peeringURLs))
	tnms := make([]peering.TrustedNetworkManager, len(peeringURLs))
	for i := range keys {
		keys[i] = cryptolib.NewKeyPair()
		tnms[i] = testutil.NewTrustedNetworkManager()
	}
	for _, tnm := range tnms {
		for i := range peeringURLs {
			_, err = tnm.TrustPeer(keys[i].GetPublicKey().String(), keys[i].GetPublicKey(), peeringURLs[i])
			require.NoError(t, err)
		}
	}
	nodes[0], _, err = lpp.NewNetworkProvider(peeringURLs[0], 9027, keys[0], tnms[0], peering.NewEmptyMetrics(), log.NewChildLogger("node0"))
	require.NoError(t, err)
	nodes[1], _, err = lpp.NewNetworkProvider(peeringURLs[1], 9028, keys[1], tnms[1], peering.NewEmptyMetrics(), log.NewChildLogger("node1"))
	require.NoError(t, err)
	nodes[2], _, err = lpp.NewNetworkProvider(peeringURLs[2], 9029, keys[2], tnms[2], peering.NewEmptyMetrics(), log.NewChildLogger("node2"))
	require.NoError(t, err)
	for i := range nodes {
		go nodes[i].Run(context.Background())
	}

	n0p2, err := nodes[0].PeerByPubKey(keys[2].GetPublicKey())
	require.NoError(t, err)
	n1p1, err := nodes[1].PeerByPubKey(keys[1].GetPublicKey())
	require.NoError(t, err)
	n2p0, err := nodes[2].PeerByPubKey(keys[0].GetPublicKey())
	require.NoError(t, err)

	chain1 := peering.RandomPeeringID()
	chain2 := peering.RandomPeeringID()
	receiver := byte(3)
	nodes[0].Attach(&chain2, receiver, func(recv *peering.PeerMessageIn) {
		doneCh <- true
	})

	n0p2.SendMsg(peering.NewPeerMessageData(chain1, receiver, 125, nil))
	n1p1.SendMsg(peering.NewPeerMessageData(chain1, receiver, 125, nil))
	n2p0.SendMsg(peering.NewPeerMessageData(chain2, receiver, 125, nil))

	<-doneCh
	time.Sleep(100 * time.Millisecond)
}
