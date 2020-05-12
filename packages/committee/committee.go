package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/registry"
	"time"
)

type Committee interface {
	Address() *address.Address
	Color() *balance.Color
	Size() uint16
	OwnPeerIndex() uint16
	SetOperational()
	Dismiss()
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time)
	IsAlivePeer(peerIndex uint16) bool
	ReceiveMessage(msg interface{})
}

var New func(scdata *registry.SCData) (Committee, error)
