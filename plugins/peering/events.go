package peering

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/coret"
)

var EventPeerMessageReceived = events.NewEvent(func(handler interface{}, params ...interface{}) {
	handler.(func(_ *PeerMessage))(params[0].(*PeerMessage))
})

type PeerMessage struct {
	ChainID     coret.ChainID
	SenderIndex uint16
	Timestamp   int64
	MsgType     byte
	MsgData     []byte
}
