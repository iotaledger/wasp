package peering

import (
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/backoff"
	"go.uber.org/atomic"
	"io"
	"net"
	"strings"
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
	isDismissed atomic.Bool       // to be GC-ed
	peerconn    *peeredConnection // nil means not connected
	handshakeOk bool
	// network locations as taken from the SC data
	remoteLocation string
	myLocation     string

	startOnce *sync.Once
	numUsers  int
	// heartbeats and latencies
	lastHeartbeatReceived time.Time
	lastHeartbeatSent     time.Time
	latencyRingBuf        [numHeartbeatsToKeep]int64
	hbRingBufIdx          int
}

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

func isInbound(remoteLocation, myLocation string) bool {
	if remoteLocation == myLocation {
		panic("remoteLocation == myLocation")
	}
	return remoteLocation < myLocation
}

func (peer *Peer) isInbound() bool {
	return isInbound(peer.remoteLocation, peer.myLocation)
}

func peeringId(remoteLocation, myLocation string) string {
	if isInbound(remoteLocation, myLocation) {
		return remoteLocation + "<" + myLocation
	} else {
		return myLocation + "<" + remoteLocation
	}
}

func (peer *Peer) PeeringId() string {
	return peeringId(peer.remoteLocation, peer.myLocation)
}

func (peer *Peer) connStatus() (bool, bool) {
	peer.RLock()
	defer peer.RUnlock()
	if peer.isDismissed.Load() {
		return false, false
	}
	return peer.peerconn != nil, peer.handshakeOk
}

func (peer *Peer) closeConn() {
	peer.Lock()
	defer peer.Unlock()

	if peer.isDismissed.Load() {
		return
	}
	if peer.peerconn != nil {
		_ = peer.peerconn.Close()
	}
}

// dials outbound address and established connection
func (peer *Peer) runOutbound() {
	if peer.isDismissed.Load() {
		return
	}
	if peer.isInbound() {
		return
	}
	if peer.peerconn != nil {
		panic("peer.peerconn != nil")
	}
	log.Debugf("runOutbound %s", peer.remoteLocation)

	defer peer.runAfter(restartAfter)

	var conn net.Conn

	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		conn, err = net.DialTimeout("tcp", peer.remoteLocation, dialTimeout)
		if err != nil {
			return fmt.Errorf("dial %s failed: %w", peer.remoteLocation, err)
		}
		return nil
	}); err != nil {
		log.Warn(err)
		return
	}
	peer.peerconn = newPeeredConnection(conn, peer)
	if err := peer.sendHandshake(); err != nil {
		log.Errorf("error during sendHandshake: %v", err)
		return
	}
	log.Debugf("starting reading outbound %s", peer.remoteLocation)
	if err := peer.peerconn.Read(); err != nil {
		if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Warnw("Permanent error", "err", err)
		}
	}
	log.Debugf("stopped reading. Closing %s", peer.remoteLocation)
	peer.closeConn()
}

// sends handshake message. It contains myLocation
func (peer *Peer) sendHandshake() error {
	data, _ := encodeMessage(&PeerMessage{
		MsgType: MsgTypeHandshake,
		MsgData: []byte(peer.PeeringId()),
	})
	_, err := peer.peerconn.Write(data)
	log.Debugf("sendHandshake '%s' --> '%s', id = %s", peer.myLocation, peer.remoteLocation, peer.PeeringId())
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
		return fmt.Errorf("no connection with %s", peer.remoteLocation)
	}
	num, err := peer.peerconn.Write(data)
	if num != len(data) {
		return fmt.Errorf("not all bytes were written. err = %v", err)
	}
	go peer.scheduleNexHeartbeat()
	return nil
}
