package peering

import "sync"

// adds new connection to the peer pool
// if it already exists, returns existing
// connection added to the pool is picked by loops which will try to establish connection
func UsePeer(remoteLocation, myLocation string) *Peer {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	if qconn, ok := peers[peeringId(remoteLocation, myLocation)]; ok {
		qconn.numUsers++
		return qconn
	}
	ret := &Peer{
		RWMutex:        &sync.RWMutex{},
		remoteLocation: remoteLocation,
		myLocation:     myLocation,
		startOnce:      &sync.Once{},
		numUsers:       1,
	}
	peers[ret.PeeringId()] = ret
	log.Debugf("added new peer id %s inbound = %v", ret.PeeringId(), ret.isInbound())
	return ret
}

// decreases counter
func StopUsingPeer(peerId string) {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	if peer, ok := peers[peerId]; ok {
		peer.numUsers--
		if peer.numUsers == 0 {
			peer.isDismissed.Store(true)

			go func() {
				peersMutex.Lock()
				defer peersMutex.Unlock()

				delete(peers, peerId)
				peer.closeConn()
			}()
		}
	}
}
