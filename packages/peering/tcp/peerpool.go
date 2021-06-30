// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcp

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
)

//nolint:unused
func (n *NetImpl) iteratePeers(f func(p *peer)) {
	n.peersMutex.Lock()
	defer n.peersMutex.Unlock()

	for _, peer := range n.peers {
		if !peer.isDismissed.Load() {
			f(peer)
		}
	}
}

func (n *NetImpl) closeAll() {
	n.peersMutex.Lock()
	defer n.peersMutex.Unlock()

	for _, cconn := range n.peers {
		cconn.closeConn()
	}
}

// loop which maintains outbound peers connected (if possible)
func (n *NetImpl) connectOutboundLoop() {
	for {
		time.Sleep(100 * time.Millisecond)
		n.peersMutex.Lock()
		for _, c := range n.peers {
			c.startOnce.Do(func() {
				go c.runOutbound()
			})
		}
		n.peersMutex.Unlock()
	}
}

// loop which maintains inbound peers connected (when possible)
func (n *NetImpl) connectInboundLoop() {
	listenOn := fmt.Sprintf(":%d", n.port)
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		n.log.Errorf("tcp listen on %s failed: %v. Shutting down...", listenOn, err)
		gracefulshutdown.Shutdown() // TODO: Move it to the plugin.

		//log.Errorf("tcp listen on %s failed: %v. Restarting connectInboundLoop after 1 sec", listenOn, err)
		//go func() {
		//	time.Sleep(1 * time.Second)
		//	connectInboundLoop()
		//}()
		return
	}
	n.log.Infof("tcp listen inbound on %s", listenOn)
	for {
		conn, err := listener.Accept()
		if err != nil {
			n.log.Errorf("failed accepting a connection request: %v", err)
			continue
		}
		n.log.Debugf("accepted connection from %s", conn.RemoteAddr().String())

		// peer is not known yet
		bconn := newPeeredConnection(conn, n, nil)
		go func() {
			n.log.Debugf("starting reading inbound %s", conn.RemoteAddr().String())
			err := bconn.Read()
			n.log.Debugw("stopped reading inbound. Closing", "remote", conn.RemoteAddr(), "err", err)

			//if err := bconn.Read(); err != nil {
			//	if permanentBufConnReadingError(err) {
			//		n.log.Warnf("Permanent error reading inbound %s: %v", conn.RemoteAddr().String(), err)
			//	}
			//}
			_ = bconn.Close()
		}()
	}
}

// for testing
//nolint:unused
func (n *NetImpl) countConnectionsLoop() {
	var totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum int
	for {
		time.Sleep(2 * time.Second)
		totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum = 0, 0, 0, 0, 0, 0, 0
		n.peersMutex.Lock()
		for _, c := range n.peers {
			totalNum++
			isConn, isHandshaken := c.connStatus()
			if c.IsInbound() {
				inboundNum++
				if isConn {
					inConnectedNum++
				}
				if isHandshaken {
					inHSNum++
				}
			} else {
				outboundNum++
				if isConn {
					outConnectedNum++
				}
				if isHandshaken {
					outHSNum++
				}
			}
		}
		n.peersMutex.Unlock()
		n.log.Debugf("CONN STATUS: total conn: %d, in: %d, out: %d, inConnected: %d, outConnected: %d, inHS: %d, outHS: %d",
			totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum)
	}
}

//nolint:unused,deadcode
func permanentBufConnReadingError(err error) bool {
	if err == io.EOF {
		return false
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		return false
	}
	if strings.Contains(err.Error(), "invalid message header") {
		// someone with wrong protocol
		return false
	}
	return true
}
