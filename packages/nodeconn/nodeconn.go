package nodeconn

import (
	"net"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/waspconn/chopper"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"github.com/iotaledger/wasp/packages/parameters"
)

type NodeConn struct {
	netID             string
	dial              DialFunc
	log               *logger.Logger
	mu                sync.Mutex
	bconn             *buffconn.BufferedConnection
	subscriptions     map[[ledgerstate.AddressLength]byte]*ledgerstate.AliasAddress
	msgChopper        *chopper.Chopper
	subscriptionsSent bool
	shutdown          chan bool
	Events            NodeConnEvents
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

func New(netID string, log *logger.Logger, dial DialFunc, goshimerNodeAddressOpt ...string) *NodeConn {
	n := &NodeConn{
		netID:         netID,
		dial:          dial,
		log:           log,
		subscriptions: make(map[[ledgerstate.AddressLength]byte]*ledgerstate.AliasAddress),
		msgChopper:    chopper.NewChopper(),
		shutdown:      make(chan bool),
		Events: NodeConnEvents{
			MessageReceived: events.NewEvent(handleMessageReceived),
			Connected:       events.NewEvent(handleConnected),
		},
	}
	var goshimerNodeAddress string
	if len(goshimerNodeAddressOpt) > 0 {
		goshimerNodeAddress = goshimerNodeAddressOpt[0]
	} else {
		goshimerNodeAddress = parameters.GetString(parameters.NodeAddress)
	}
	go n.nodeConnect()
	go n.keepSendingSubscriptionIfNeeded(goshimerNodeAddress)
	go n.keepSendingSubscriptionForced(goshimerNodeAddress)
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

// checking if need to be sent every second
func (n *NodeConn) keepSendingSubscriptionIfNeeded(goshimerNodeAddress string) {
	for {
		select {
		case <-n.shutdown:
			return
		case <-time.After(1 * time.Second):
			n.sendSubscriptions(false, goshimerNodeAddress)
		}
	}
}

// will be sending subscriptions every minute to pull backlog
// needed in case node is not synced
func (n *NodeConn) keepSendingSubscriptionForced(goshimerNodeAddress string) {
	for {
		select {
		case <-n.shutdown:
			return
		case <-time.After(1 * time.Minute):
			n.sendSubscriptions(true, goshimerNodeAddress)
		}
	}
}
