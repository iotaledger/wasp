package peering

import (
	"time"
)

// message type is 1 byte
// from 0 until maxSpecMsgCode inclusive it is reserved for heartbeat and other message types
// these messages are processed by processHeartbeat method
// the rest are forwarded to SC operator

func (peer *Peer) initHeartbeats() {
	peer.lastHeartbeatSent = 0
	peer.lastHeartbeatReceived = 0
	peer.hbRingBufIdx = 0
	for i := range peer.latencyRingBuf {
		peer.latencyRingBuf[i] = 0
	}
}

func (peer *Peer) receiveHeartbeat(ts int64) {
	peer.Lock()
	peer.lastHeartbeatReceived = time.Now().UnixNano()
	//log.Debugw("receiveHeartbeat", "id", peer.PeeringId(), "time", peer.lastHeartbeatReceived)
	lagNano := peer.lastHeartbeatReceived - ts
	peer.latencyRingBuf[peer.hbRingBufIdx] = lagNano
	peer.hbRingBufIdx = (peer.hbRingBufIdx + 1) % numHeartbeatsToKeep
	peer.Unlock()

	//log.Debugf("heartbeat received from %s, lag %f milisec", peer.remoteLocation.String(), float64(lagNano/10000)/100)
}

func (peer *Peer) scheduleNexHeartbeat() {
	//log.Debugw("scheduleNexHeartbeat", "id", peer.PeeringId())

	if peerAlive, _ := peer.IsAlive(); !peerAlive {
		//log.Debugw("scheduleNexHeartbeat", "id", peer.PeeringId(), "isAlive", peerAlive)
		return
	}
	time.Sleep(heartbeatEvery)
	if peerAlive, _ := peer.IsAlive(); !peerAlive {
		//log.Debugw("scheduleNexHeartbeat", "id", peer.PeeringId(), "isAlive", peerAlive)
		return
	}

	peer.Lock()

	//log.Debugw("scheduleNexHeartbeat", "id", peer.PeeringId(), "since last", time.Since(peer.lastHeartbeatSent))

	if time.Since(time.Unix(0, peer.lastHeartbeatSent)) <= heartbeatEvery {
		//log.Debugw("scheduleNexHeartbeat: too early", "id", peer.PeeringId())
		peer.Unlock()
		return
	}
	var hbMsgData []byte
	hbMsgData, peer.lastHeartbeatSent = encodeMessage(nil)

	peer.Unlock()

	//log.Debugw("sending heartbeat", "to", peer.remoteLocation)
	err := peer.sendData(hbMsgData)
	if err != nil {
		log.Debugw("sending heartbeat error",
			"to", peer.remoteLocation,
			"err", err,
		)
	}
}

// return true if is alive and average latencyRingBuf in nanosec
func (peer *Peer) IsAlive() (bool, int64) {
	peer.RLock()
	defer peer.RUnlock()
	if peer.peerconn == nil || !peer.handshakeOk {
		return false, 0
	}

	if time.Since(time.Unix(0, peer.lastHeartbeatReceived)) > heartbeatEvery*isDeadAfterMissing {
		return false, 0
	}
	sum := int64(0)
	for _, l := range peer.latencyRingBuf {
		sum += l
	}
	return true, sum / numHeartbeatsToKeep
}
