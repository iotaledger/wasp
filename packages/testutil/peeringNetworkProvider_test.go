// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/testutil"
)

func TestFakeNetwork(t *testing.T) {
	log := testutil.NewLogger(t)
	defer log.Sync()
	doneCh := make(chan bool)
	chain1 := coretypes.NewRandomChainID()
	chain2 := coretypes.NewRandomChainID()
	network := testutil.NewPeeringNetworkForLocs([]string{"a", "b", "c"}, 100, log)
	var netProviders []peering.NetworkProvider = network.NetworkProviders()
	//
	// Node "a" listens for chain1 messages.
	netProviders[0].Attach(&chain1, func(recv *peering.RecvEvent) {
		doneCh <- true
	})
	//
	// Node "b" sends some messages.
	var a, c peering.PeerSender
	a, _ = netProviders[1].PeerByNetID("a")
	c, _ = netProviders[1].PeerByNetID("c")
	a.SendMsg(&peering.PeerMessage{ChainID: chain1, MsgType: 1}) // Will be delivered.
	a.SendMsg(&peering.PeerMessage{ChainID: chain2, MsgType: 2}) // Will be dropped.
	c.SendMsg(&peering.PeerMessage{ChainID: chain1, MsgType: 3}) // Will be dropped.
	//
	// Wait for the result.
	select {
	case <-doneCh:
	case <-time.After(1 * time.Second):
		panic("timeout")
	}
}
