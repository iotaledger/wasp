package peering

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/events"
)

var EventPeerMessageReceived *events.Event

func init() {
	EventPeerMessageReceived = events.NewEvent(func(handler interface{}, params ...interface{}) {
		handler.(func(_ *PeerMessage))(params[0].(*PeerMessage))
	})
}

type PeerMessage struct {
	Address     address.Address
	SenderIndex uint16
	Timestamp   int64
	MsgType     byte
	MsgData     []byte
}
