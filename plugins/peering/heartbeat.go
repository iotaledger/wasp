package peering

import (
	"time"
)

// message type is 1 byte
// from 0 until maxSpecMsgCode inclusive it is reserved for heartbeat and other message types
// these messages are processed by processHeartbeat method
// the rest are forwarded to SC operator

func (peer *Peer) initHeartbeats() {
	peer.lastHeartbeatSent = time.Time{}
	peer.lastHeartbeatReceived = time.Time{}
	peer.hbRingBufIdx = 0
	for i := range peer.latencyRingBuf {
		peer.latencyRingBuf[i] = 0
	}
}

func (peer *Peer) receiveHeartbeat(ts int64) {
	peer.Lock()
	peer.lastHeartbeatReceived = time.Now()
	lagNano := peer.lastHeartbeatReceived.UnixNano() - ts
	peer.latencyRingBuf[peer.hbRingBufIdx] = lagNano
	peer.hbRingBufIdx = (peer.hbRingBufIdx + 1) % numHeartbeatsToKeep
	peer.Unlock()

	//log.Debugf("heartbeat received from %s, lag %f milisec", peer.remoteLocation.String(), float64(lagNano/10000)/100)
}

func (peer *Peer) scheduleNexHeartbeat() {
	if peerAlive, _ := peer.IsAlive(); !peerAlive {
		//log.Debugf("stopped sending heartbeat: peer %s is dead. peering id %s", peer.remoteLocation, peer.PeeringId())
		return
	}
	time.Sleep(heartbeatEvery)
	if peerAlive, _ := peer.IsAlive(); !peerAlive {
		//log.Debugf("stopped sending heartbeat: peer %s is dead. peering id %s", peer.remoteLocation, peer.PeeringId())
		return
	}

	peer.Lock()

	if time.Since(peer.lastHeartbeatSent) < heartbeatEvery {
		// was recently sent. exit
		peer.Unlock()
		return
	}
	var hbMsgData []byte
	hbMsgData, peer.lastHeartbeatSent = encodeMessage(nil)

	peer.Unlock()

	_ = peer.sendData(hbMsgData)
}

// return true if is alive and average latencyRingBuf in nanosec
func (peer *Peer) IsAlive() (bool, int64) {
	peer.RLock()
	defer peer.RUnlock()
	if peer.peerconn == nil || !peer.handshakeOk {
		return false, 0
	}

	if time.Since(peer.lastHeartbeatReceived) > heartbeatEvery*isDeadAfterMissing {
		return false, 0
	}
	sum := int64(0)
	for _, l := range peer.latencyRingBuf {
		sum += l
	}
	return true, sum / numHeartbeatsToKeep
}
