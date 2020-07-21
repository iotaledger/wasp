package peering

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/chopper"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/payload"
	"github.com/iotaledger/hive.go/backoff"
	"go.uber.org/atomic"
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
	isDismissed atomic.Bool       // to be GC-ed
	peerconn    *peeredConnection // nil means not connected
	handshakeOk bool
	// network locations as taken from the SC data
	remoteLocation string

	startOnce *sync.Once
	numUsers  int
	// heartbeats and latencies
	lastHeartbeatReceived int64
	lastHeartbeatSent     int64
	latencyRingBuf        [numHeartbeatsToKeep]int64
	hbRingBufIdx          int
}

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

func isInbound(remoteLocation string) bool {
	if remoteLocation == MyNetworkId() {
		panic("remoteLocation == myLocation")
	}
	return remoteLocation < MyNetworkId()
}

func (peer *Peer) isInbound() bool {
	return isInbound(peer.remoteLocation)
}

func peeringId(remoteLocation string) string {
	if isInbound(remoteLocation) {
		return remoteLocation + "<" + MyNetworkId()
	} else {
		return MyNetworkId() + "<" + remoteLocation
	}
}

func (peer *Peer) PeeringId() string {
	return peeringId(peer.remoteLocation)
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

	// always try to reconnect
	defer func() {
		go func() {
			time.Sleep(restartAfter)
			peer.Lock()
			if !peer.isDismissed.Load() {
				peer.startOnce = &sync.Once{}
				log.Debugf("will run again: %s", peer.PeeringId())
			}
			peer.Unlock()
		}()
	}()

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
	err := peer.peerconn.Read()
	log.Debugw("stopped reading outbound. Closing", "remote", peer.remoteLocation, "err", err)
	peer.closeConn()
	//if ; err != nil {
	//	log.Warnw("")
	//	if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
	//		log.Warnw("Permanent error", "err", err)
	//	}
	//}
}

// sends handshake message. It contains myLocation
func (peer *Peer) sendHandshake() error {
	data, _ := encodeMessage(&PeerMessage{
		MsgType: MsgTypeHandshake,
		MsgData: []byte(peer.PeeringId()),
	})
	_, err := peer.peerconn.Write(data)
	log.Debugf("sendHandshake '%s' --> '%s', id = %s", MyNetworkId(), peer.remoteLocation, peer.PeeringId())
	return err
}

func (peer *Peer) SendMsg(msg *PeerMessage) error {
	//log.Debugw("SendMsg", "id", peer.PeeringId(), "msgType", msg.MsgType)

	if msg.MsgType < FirstCommitteeMsgCode {
		return errors.New("reserved message code")
	}
	data, ts := encodeMessage(msg)

	peer.lastHeartbeatSent = ts

	choppedData, chopped := chopper.ChopData(data, payload.MaxMessageSize-chunkMessageOverhead)

	peer.RLock()
	defer peer.RUnlock()

	if !chopped {
		return peer.sendData(data)
	}
	return peer.sendChunks(choppedData)
}

func (peer *Peer) sendChunks(chopped [][]byte) error {
	for _, piece := range chopped {
		d, _ := encodeMessage(&PeerMessage{
			MsgType: MsgTypeMsgChunk,
			MsgData: piece,
		})
		if err := peer.sendData(d); err != nil {
			return err
		}
	}
	return nil
}

// sends same msg to all peers in the slice which are not nil
// with the same timestamp
func SendMsgToPeers(msg *PeerMessage, peers ...*Peer) (uint16, int64) {
	if msg.MsgType < FirstCommitteeMsgCode {
		return 0, 0
	}
	// timestamped here, once
	data, ts := encodeMessage(msg)
	choppedData, chopped := chopper.ChopData(data, payload.MaxMessageSize-chunkMessageOverhead)

	ret := uint16(0)
	for _, peer := range peers {
		if peer == nil {
			continue
		}
		peer.RLock()
		if !chopped {
			peer.lastHeartbeatSent = ts
			if err := peer.sendData(data); err == nil {
				ret++
			}
		} else {
			if err := peer.sendChunks(choppedData); err == nil {
				ret++
			}
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
