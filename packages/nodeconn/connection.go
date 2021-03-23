// package for maintaining connection with the main node
// on the node WaspConn plugin is handling yhe connection
package nodeconn

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/tangle"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/waspconn/chopper"
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

func (n *NodeConn) nodeConnectLoop(dial DialFunc) {
	msgChopper := chopper.NewChopper()
	n.wgConnected.Add(1)
	for {
		retry := n.connect(dial, msgChopper)
		if !retry {
			return
		}
		n.log.Infof("disconnected from node - will retry reconnecting after %v", retryAfter)
		select {
		case <-n.shutdown:
			return
		case <-time.After(retryAfter):
		}
	}
}

// dials outbound address and established connection
func (n *NodeConn) connect(dial DialFunc, msgChopper *chopper.Chopper) bool {
	var addr string
	var conn net.Conn
	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		addr, conn, err = dial()
		if err != nil {
			return fmt.Errorf("can't connect with the node: %v", err)
		}
		return nil
	}); err != nil {
		n.log.Warn(err)
		// retry
		return true
	}

	bconn := buffconn.NewBufferedConnection(conn, tangle.MaxMessageSize)
	defer bconn.Close()
	n.Events.Connected.Trigger()

	n.wgConnected.Done()
	defer n.wgConnected.Add(1)

	n.log.Debugf("established connection with node at %s", addr)

	dataReceived := make(chan []byte)
	dataReceivedClosure := events.NewClosure(func(data []byte) {
		// data slice is from internal buffconn buffer
		d := make([]byte, len(data))
		copy(d, data)
		dataReceived <- d
	})
	bconn.Events.ReceiveMessage.Attach(dataReceivedClosure)
	defer bconn.Events.ReceiveMessage.Detach(dataReceivedClosure)

	connectionClosed := make(chan bool)
	connectionClosedClosure := events.NewClosure(func() {
		n.log.Errorf("lost connection with %s", addr)
		close(connectionClosed)
	})
	bconn.Events.Close.Attach(connectionClosedClosure)
	defer bconn.Events.Close.Detach(connectionClosedClosure)

	go n.SendWaspIdToNode()

	// read loop
	go func() {
		if err := bconn.Read(); err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
				n.log.Warnw("Permanent error", "err", err)
			}
		}
	}()

	// r/w loop
	for {
		select {
		case d := <-dataReceived:
			if err := n.decodeReceivedMessage(d, msgChopper); err != nil {
				n.log.Errorf("decoding message from node: %v", err)
			}
		case msg := <-n.chSendToNode:
			if err := n.sendToNode(msg, bconn, msgChopper); err != nil {
				n.log.Errorf("sending message to node: %v", err)
			}
		case <-n.shutdown:
			return false
		case <-connectionClosed:
			return true // retry
		}
	}
}

func (n *NodeConn) decodeReceivedMessage(data []byte, msgChopper *chopper.Chopper) error {
	n.log.Debugf("received %d bytes from node (md5 %x)", len(data), md5.Sum(data))
	msg, err := waspconn.DecodeMsg(data, true)
	if err != nil {
		return fmt.Errorf("waspconn.DecodeMsg: %w", err)
	}

	switch msg := msg.(type) {
	case *waspconn.WaspMsgChunk:
		finalData, err := msgChopper.IncomingChunk(msg.Data, tangle.MaxMessageSize, waspconn.ChunkMessageHeaderSize)
		if err != nil {
			return fmt.Errorf("receiving msgchunk: %w", err)
		}
		if finalData != nil {
			return n.decodeReceivedMessage(finalData, msgChopper)
		}

	default:
		n.log.Debugf("received message from node: %T", msg)
		n.Events.MessageReceived.Trigger(msg)
	}
	return nil
}

func (n *NodeConn) WaitForConnection(timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		n.wgConnected.Wait()
	}()
	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (n *NodeConn) SendToNode(msg waspconn.Message) {
	n.chSendToNode <- msg
}

func (n *NodeConn) sendToNode(msg waspconn.Message, bconn *buffconn.BufferedConnection, msgChopper *chopper.Chopper) error {
	n.log.Debugf("sending message to node: %T", msg)
	data := waspconn.EncodeMsg(msg)
	choppedData, chopped, err := msgChopper.ChopData(data, tangle.MaxMessageSize, waspconn.ChunkMessageHeaderSize)
	if err != nil {
		return err
	}
	if !chopped {
		_, err = bconn.Write(data)
		return err
	}
	for _, piece := range choppedData {
		if _, err = bconn.Write(waspconn.EncodeMsg(&waspconn.WaspMsgChunk{Data: piece})); err != nil {
			return err
		}
	}
	return nil
}
