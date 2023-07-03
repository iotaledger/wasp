package peering_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestPeerMessageSerialization(t *testing.T) {
	msg := &peering.PeerMessageNet{
		PeerMessageData: peering.NewPeerMessageData(
			peering.RandomPeeringID(),
			byte(10),
			peering.FirstUserMsgCode+17,
			[]byte{1, 2, 3, 4, 5}),
	}
	rwutil.BytesTest(t, msg, peering.PeerMessageNetFromBytes)
}
