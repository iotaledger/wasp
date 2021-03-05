package nodeconn

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn/chopper"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/netutil/buffconn"
)

type NodeConn struct {
	netID                string
	log                  *logger.Logger
	bconn                *buffconn.BufferedConnection
	bconnMutex           sync.Mutex
	subscriptions        map[ledgerstate.Address]ledgerstate.Color
	msgChopper           *chopper.Chopper
	subscriptionsSent    bool
	shutdown             chan bool
	EventMessageReceived *events.Event
}

func param1Caller(handler interface{}, params ...interface{}) {
	handler.(func(interface{}))(params[0])
}

func New(netID string, log *logger.Logger) *NodeConn {
	n := &NodeConn{
		netID:                netID,
		log:                  log,
		subscriptions:        make(map[ledgerstate.Address]ledgerstate.Color),
		msgChopper:           chopper.NewChopper(),
		shutdown:             make(chan bool),
		EventMessageReceived: events.NewEvent(param1Caller),
	}
	go n.nodeConnect()
	go n.keepSendingSubscriptionIfNeeded()
	go n.keepSendingSubscriptionForced()
	return n
}

func (n *NodeConn) Close() {
	go func() {
		n.bconnMutex.Lock()
		defer n.bconnMutex.Unlock()
		if n.bconn != nil {
			n.log.Infof("Closing connection with node..")
			_ = n.bconn.Close()
			n.log.Infof("Closing connection with node.. Done")
		}
	}()
	close(n.shutdown)
	n.EventMessageReceived.DetachAll()
}

// checking if need to be sent every second
func (n *NodeConn) keepSendingSubscriptionIfNeeded() {
	for {
		select {
		case <-n.shutdown:
			return
		case <-time.After(1 * time.Second):
			n.sendSubscriptions(false)
		}
	}
}

// will be sending subscriptions every minute to pull backlog
// needed in case node is not synced
func (n *NodeConn) keepSendingSubscriptionForced() {
	for {
		select {
		case <-n.shutdown:
			return
		case <-time.After(1 * time.Minute):
			n.sendSubscriptions(true)
		}
	}
}
