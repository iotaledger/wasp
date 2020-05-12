package peering

import (
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/backoff"
	"github.com/iotaledger/wasp/packages/registry"
	"net"
	"sync"
	"time"
)

// represents point-to-point TCP connection between two qnodes and another
// it is used as transport for message exchange
// Another end is always using the same connection
// the Peer takes care about exchanging heartbeat messages.
// It keeps last several received heartbeats as "lad" data to be able to calculate how synced/unsynced
// clocks of peer are.
type Peer struct {
	*sync.RWMutex
	peerconn     *peeredConnection // nil means not connected
	handshakeOk  bool
	peerPortAddr *registry.PortAddr
	startOnce    *sync.Once
	numUsers     int
	// heartbeats and latencies
	lastHeartbeatReceived time.Time
	lastHeartbeatSent     time.Time
	latencyRingBuf        [numHeartbeatsToKeep]int64
	hbRingBufIdx          int
}

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

func (peer *Peer) isInbound() bool {
	return isInboundAddr(peer.peerPortAddr.String())
}

func (peer *Peer) connStatus() (bool, bool) {
	peer.RLock()
	defer peer.RUnlock()
	return peer.peerconn != nil, peer.handshakeOk
}

func (peer *Peer) closeConn() {
	peer.Lock()
	defer peer.Unlock()
	if peer.peerconn != nil {
		_ = peer.peerconn.Close()
	}
}

// dials outbound address and established connection
func (peer *Peer) runOutbound() {
	if peer.isInbound() {
		return
	}
	if peer.peerconn != nil {
		panic("peer.peerconn != nil")
	}
	log.Debugf("runOutbound %s", peer.peerPortAddr.String())

	defer peer.runAfter(restartAfter)

	var conn net.Conn
	addr := fmt.Sprintf("%s:%d", peer.peerPortAddr.Addr, peer.peerPortAddr.Port)
	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		conn, err = net.DialTimeout("tcp", addr, dialTimeout)
		if err != nil {
			return fmt.Errorf("dial %s failed: %w", addr, err)
		}
		return nil
	}); err != nil {
		log.Error(err)
		return
	}
	//manconn := network.NewManagedConnection(conn)
	peer.peerconn = newPeeredConnection(conn, peer)
	if err := peer.sendHandshake(); err != nil {
		log.Errorf("error during sendHandshake: %v", err)
		return
	}
	log.Debugf("starting reading outbound %s", peer.peerPortAddr.String())
	if err := peer.peerconn.Read(); err != nil {
		log.Error(err)
	}
	log.Debugf("stopped reading. Closing %s", peer.peerPortAddr.String())
	peer.closeConn()
}

// sends handshake message. It contains IP address of this end.
// The address is used by another end for peering
func (peer *Peer) sendHandshake() error {
	data, _ := encodeMessage(&PeerMessage{
		MsgType: MsgTypeHandshake,
		MsgData: []byte(OwnPortAddr().String()),
	})
	num, err := peer.peerconn.Write(data)
	log.Debugf("sendHandshake %d bytes to %s", num, peer.peerPortAddr.String())
	return err
}

func (peer *Peer) SendMsg(msg *PeerMessage) error {
	if msg.MsgType < FirstCommitteeMsgCode {
		return errors.New("reserved message code")
	}
	data, ts := encodeMessage(msg)
	peer.RLock()
	defer peer.RUnlock()

	peer.lastHeartbeatSent = ts
	return peer.sendData(data)
}

// sends same msg to all peers in the slice which are not nil
// with the same timestamp
func SendMsgToPeers(msg *PeerMessage, peers ...*Peer) (uint16, time.Time) {
	if msg.MsgType < FirstCommitteeMsgCode {
		return 0, time.Time{}
	}
	// timestamped here, once
	data, ts := encodeMessage(msg)
	ret := uint16(0)
	for _, peer := range peers {
		if peer == nil {
			continue
		}
		peer.RLock()
		peer.lastHeartbeatSent = ts
		if err := peer.sendData(data); err == nil {
			ret++
		}
		peer.RUnlock()
	}
	return ret, ts
}

func (peer *Peer) sendData(data []byte) error {
	if peer.peerconn == nil {
		return fmt.Errorf("error while sending data: connection with %s not established", peer.peerPortAddr.String())
	}
	num, err := peer.peerconn.Write(data)
	if num != len(data) {
		return fmt.Errorf("not all bytes written. err = %v", err)
	}
	go peer.scheduleNexHeartbeat()
	return nil
}
