package peering

import (
	"github.com/iotaledger/wasp/packages/parameters"
	"sync"
)

func MyNetworkId() string {
	return parameters.GetString(parameters.PeeringMyNetId)
}

// adds new connection to the peer pool
// if it already exists, returns existing
// connection added to the pool is picked by loops which will try to establish connection
func UsePeer(remoteLocation string) *Peer {
	if !initialized.Load() {
		return nil
	}
	if remoteLocation == MyNetworkId() {
		return nil
	}
	peersMutex.Lock()
	defer peersMutex.Unlock()

	if qconn, ok := peers[peeringId(remoteLocation)]; ok {
		qconn.numUsers++
		return qconn
	}
	ret := &Peer{
		RWMutex:        &sync.RWMutex{},
		remoteLocation: remoteLocation,
		startOnce:      &sync.Once{},
		numUsers:       1,
	}
	peers[ret.PeeringId()] = ret
	log.Debugf("added new peer id %s inbound = %v", ret.PeeringId(), ret.isInbound())
	return ret
}

// decreases counter
func StopUsingPeer(peerId string) {
	if !initialized.Load() {
		return
	}
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
