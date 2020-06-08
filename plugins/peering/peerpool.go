package peering

import (
	"fmt"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	peers      = make(map[string]*Peer)
	peersMutex = &sync.RWMutex{}
)

func closeAll() {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	for _, cconn := range peers {
		cconn.closeConn()
	}
}

// wait some time the rests peer to be connected by the loops
func (peer *Peer) runAfter(d time.Duration) {
	go func() {
		time.Sleep(d)
		peer.Lock()
		if !peer.isDismissed.Load() {
			peer.startOnce = &sync.Once{}
			log.Debugf("will run %s again", peer.PeeringId())
		}
		peer.Unlock()
	}()
}

// loop which maintains outbound peers connected (if possible)
func connectOutboundLoop() {
	for {
		time.Sleep(100 * time.Millisecond)
		peersMutex.Lock()
		for _, c := range peers {
			c.startOnce.Do(func() {
				go c.runOutbound()
			})
		}
		peersMutex.Unlock()
	}
}

// loop which maintains inbound peers connected (when possible)
func connectInboundLoop() {
	listenOn := fmt.Sprintf(":%d", config.Node.GetInt(CfgPeeringPort))
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		log.Errorf("tcp listen on %s failed: %v. Shutting down...", listenOn, err)
		gracefulshutdown.Shutdown()

		//log.Errorf("tcp listen on %s failed: %v. Restarting connectInboundLoop after 1 sec", listenOn, err)
		//go func() {
		//	time.Sleep(1 * time.Second)
		//	connectInboundLoop()
		//}()
		return
	}
	log.Infof("tcp listen inbound on %s", listenOn)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("failed accepting a connection request: %v", err)
			continue
		}
		log.Debugf("accepted connection from %s", conn.RemoteAddr().String())

		bconn := newPeeredConnection(conn, nil)
		go func() {
			log.Debugf("starting reading inbound %s", conn.RemoteAddr().String())
			if err := bconn.Read(); err != nil {
				if permanentBufconnReadingError(err) {
					log.Warnf("Permanent error reading inbound %s: %v", conn.RemoteAddr().String(), err)
				}
			}
			_ = bconn.Close()
		}()
	}
}

func permanentBufconnReadingError(err error) bool {
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

// for testing
func countConnectionsLoop() {
	var totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum int
	for {
		time.Sleep(2 * time.Second)
		totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum = 0, 0, 0, 0, 0, 0, 0
		peersMutex.Lock()
		for _, c := range peers {
			totalNum++
			isConn, isHandshaken := c.connStatus()
			if c.isInbound() {
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
		peersMutex.Unlock()
		log.Debugf("CONN STATUS: total conn: %d, in: %d, out: %d, inConnected: %d, outConnected: %d, inHS: %d, outHS: %d",
			totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum)
	}
}
