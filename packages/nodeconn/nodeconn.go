package nodeconn

import (
	"net"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/waspconn/chopper"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/netutil/buffconn"
)

type NodeConn struct {
	netID         string
	dial          DialFunc
	log           *logger.Logger
	mu            sync.Mutex
	bconn         *buffconn.BufferedConnection
	chSubscribe   chan *ledgerstate.AliasAddress
	chUnsubscribe chan *ledgerstate.AliasAddress
	msgChopper    *chopper.Chopper
	shutdown      chan bool
	Events        NodeConnEvents
}

type NodeConnEvents struct {
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
		dial:          dial,
		log:           log,
		chSubscribe:   make(chan *ledgerstate.AliasAddress),
		chUnsubscribe: make(chan *ledgerstate.AliasAddress),
		msgChopper:    chopper.NewChopper(),
		shutdown:      make(chan bool),
		Events: NodeConnEvents{
			MessageReceived: events.NewEvent(handleMessageReceived),
			Connected:       events.NewEvent(handleConnected),
		},
	}

	go n.subscriptionsLoop()
	go n.nodeConnectLoop()

	return n
}

func (n *NodeConn) Close() {
	go func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		if n.bconn != nil {
			n.log.Infof("Closing connection with node..")
			_ = n.bconn.Close()
			n.log.Infof("Closing connection with node.. Done")
		}
	}()
	close(n.shutdown)
	n.Events.MessageReceived.DetachAll()
}
