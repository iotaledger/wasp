package nodeconn

import (
	"net"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
)

type NodeConn struct {
	netID         string
	log           *logger.Logger
	chSendToNode  chan waspconn.Message
	chSubscribe   chan *ledgerstate.AliasAddress
	chUnsubscribe chan *ledgerstate.AliasAddress
	shutdown      chan bool
	Events        Events
	wgConnected   sync.WaitGroup
}

type Events struct {
	MessageReceived *events.Event
	Connected       *events.Event
}

type DialFunc func() (addr string, conn net.Conn, err error)

func handleMessageReceived(handler interface{}, params ...interface{}) {
	handler.(func(waspconn.Message))(params[0].(waspconn.Message))
}

func handleConnected(handler interface{}, params ...interface{}) {
	handler.(func())()
}

func New(netID string, log *logger.Logger, dial DialFunc) *NodeConn {
	n := &NodeConn{
		netID:         netID,
		log:           log,
		chSendToNode:  make(chan waspconn.Message),
		chSubscribe:   make(chan *ledgerstate.AliasAddress),
		chUnsubscribe: make(chan *ledgerstate.AliasAddress),
		shutdown:      make(chan bool),
		Events: Events{
			MessageReceived: events.NewEvent(handleMessageReceived),
			Connected:       events.NewEvent(handleConnected),
		},
	}

	go n.subscriptionsLoop()
	go n.nodeConnectLoop(dial)

	return n
}

func (n *NodeConn) Close() {
	close(n.shutdown)
	n.Events.MessageReceived.DetachAll()
	n.Events.Connected.DetachAll()
}
