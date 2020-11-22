package testutil_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/plugins/peering"
)

func TestFakeNetwork(t *testing.T) {
	doneCh := make(chan bool)
	chain1 := coretypes.RandomChainID()
	chain2 := coretypes.RandomChainID()
	network := testutil.NewPeeringNetworkForLocs([]string{"a", "b", "c"}, 100)
	var netProviders []peering.NetworkProvider = network.NetworkProviders()
	//
	// Node "a" listens for chain1 messages.
	netProviders[0].Attach(chain1, func(from peering.PeerSender, msg *peering.PeerMessage) {
		doneCh <- true
	})
	//
	// Node "b" sends some messages.
	netProviders[1].SendByLocation("a", &peering.PeerMessage{ChainID: chain1, MsgType: 1}) // Will be delivered.
	netProviders[1].SendByLocation("a", &peering.PeerMessage{ChainID: chain2, MsgType: 2}) // Will be dropped.
	netProviders[1].SendByLocation("c", &peering.PeerMessage{ChainID: chain1, MsgType: 3}) // Will be dropped.
	//
	// Wait for the result.
	select {
	case <-doneCh:
	case <-time.After(1 * time.Second):
		panic("timeout")
	}
}
