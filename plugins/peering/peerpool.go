package peering

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/config"
	"net"
	"sync"
	"time"
)

var (
	// all qnode peers maintained by the node
	// map index is IP addr:port
	peers      = make(map[string]*Peer)
	peersMutex = &sync.RWMutex{}
)

func Init() {
}

// IP address and port of this node
func OwnPortAddr() *registry.PortAddr {
	return &registry.PortAddr{
		Port: config.Node.GetInt(CfgPeeringPort),
		Addr: "127.0.0.1", // TODO for testing only
	}
}

func FindOwnIndex(netLocations []*registry.PortAddr) (uint16, bool) {
	ownLoc := OwnPortAddr().String()
	for i, loc := range netLocations {
		if ownLoc == loc.String() {
			return uint16(i), true
		}
	}
	return 0, false
}

func closeAll() {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	for _, cconn := range peers {
		cconn.closeConn()
	}
}

// determines if the address of the peer inbound or outbound
// it is guaranteed that the peer on the other end will always get the opposite result
// It is determined by comparing address string with own address string
// panics if address equals to the own address
// This helps peers to know their role in the peer-to-peer connection
func isInboundAddr(addr string) bool {
	own := OwnPortAddr().String()
	if own == addr {
		// shouldn't come to this point due to checks before
		panic("can't be same PortAddr")
	}
	return addr < own
}

// adds new connection to the peer pool
// if it already exists, returns existing
// connection added to the pool is picked by loops which will try to establish connection
func UsePeer(portAddr *registry.PortAddr) *Peer {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	addr := portAddr.String()
	if qconn, ok := peers[addr]; ok {
		qconn.numUsers++
		return qconn
	}
	peers[addr] = &Peer{
		RWMutex:      &sync.RWMutex{},
		peerPortAddr: portAddr,
		startOnce:    &sync.Once{},
		numUsers:     1,
	}
	log.Debugf("added new peer %s inbound = %v", addr, peers[addr].isInbound())
	return peers[addr]
}

// decreases counter
func StopUsingPeer(portAddr *registry.PortAddr) {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	loc := portAddr.String()
	if peer, ok := peers[loc]; ok {
		peer.numUsers--
		if peer.numUsers == 0 {
			go func() {
				peersMutex.Lock()
				defer peersMutex.Unlock()

				delete(peers, loc)
				peer.closeConn()
			}()
		}
	}
}

// wait some time the rests peer to be connected by the loops
func (peer *Peer) runAfter(d time.Duration) {
	go func() {
		time.Sleep(d)
		peer.Lock()
		peer.startOnce = &sync.Once{}
		peer.Unlock()
		log.Debugf("will run %s again", peer.peerPortAddr.String())
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
		log.Errorf("tcp listen on %s failed: %v. Restarting connectInboundLoop after 1 sec", listenOn, err)
		go func() {
			time.Sleep(1 * time.Second)
			connectInboundLoop()
		}()
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
				log.Error(err)
			}
			_ = bconn.Close()
		}()
	}
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
