// package for maintaining connection with the main node
// on the node WaspConn plugin is handling yhe connection
package nodeconn

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/tangle"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/hive.go/backoff"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
)

const (
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
	retryAfter   = 8 * time.Second
)

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

// dials outbound address and established connection
func (n *NodeConn) nodeConnectLoop() {
	var addr string
	var conn net.Conn
	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		addr, conn, err = n.dial()
		if err != nil {
			return fmt.Errorf("can't connect with the node: %v", err)
		}
		return nil
	}); err != nil {
		n.log.Warn(err)

		n.retryNodeConnect()
		return
	}

	n.mu.Lock()
	n.bconn = buffconn.NewBufferedConnection(conn, tangle.MaxMessageSize)
	n.mu.Unlock()
	n.Events.Connected.Trigger()

	n.log.Debugf("established connection with node at %s", addr)

	dataReceivedClosure := events.NewClosure(func(data []byte) {
		n.msgDataToEvent(data)
	})

	n.bconn.Events.ReceiveMessage.Attach(dataReceivedClosure)
	n.bconn.Events.Close.Attach(events.NewClosure(func() {
		n.log.Errorf("lost connection with %s", addr)
		go func() {
			n.mu.Lock()
			bconnOld := n.bconn
			defer bconnOld.Events.ReceiveMessage.Detach(dataReceivedClosure)
			n.bconn = nil
			n.mu.Unlock()
		}()
	}))

	if err := n.SendWaspIdToNode(); err == nil {
		n.log.Debugf("sent own wasp id to node: %s", n.netID)
	} else {
		n.log.Errorf("failed to send wasp id to node: %v", err)
	}

	// read loop
	if err := n.bconn.Read(); err != nil {
		if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
			n.log.Warnw("Permanent error", "err", err)
		}
	}
	// try to reconnect after some time
	n.log.Debugf("disconnected from node. Will try to reconnect after %v", retryAfter)

	n.retryNodeConnect()
}

func (n *NodeConn) WaitForConnection(timeout time.Duration) bool {
	if n.IsConnected() {
		return true
	}

	ok := make(chan bool)
	cl := events.NewClosure(func() {
		close(ok)
	})
	n.Events.Connected.Attach(cl)
	defer n.Events.Connected.Detach(cl)

	select {
	case <-ok:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (n *NodeConn) IsConnected() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.bconn != nil
}

func (n *NodeConn) retryNodeConnect() {
	n.log.Infof("will retry connecting to the node after %v", retryAfter)
	time.Sleep(retryAfter)
	go n.nodeConnectLoop()
}

func (n *NodeConn) sendToNode(msg waspconn.Message) error {
	data := waspconn.EncodeMsg(msg)
	choppedData, chopped, err := n.msgChopper.ChopData(data, tangle.MaxMessageSize, waspconn.ChunkMessageHeaderSize)
	if err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.bconn == nil {
		return fmt.Errorf("sendToNode: not connected to node")
	}
	if !chopped {
		_, err = n.bconn.Write(data)
	} else {
		for _, piece := range choppedData {
			d := waspconn.EncodeMsg(&waspconn.WaspMsgChunk{
				Data: piece,
			})
			_, err = n.bconn.Write(d)
			if err != nil {
				break
			}
		}
	}
	return err
}
