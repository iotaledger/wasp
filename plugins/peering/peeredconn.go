package peering

import (
	"net"

	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/chopper"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/payload"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
)

// extension of BufferedConnection from hive.go
// BufferedConnection is a wrapper for net.Conn
// peeredConnection first handles handshake and then links
// with peer (peers) according to the handshake information
type peeredConnection struct {
	*buffconn.BufferedConnection
	peer        *Peer
	handshakeOk bool
}

// creates new peered connection and attach event handlers for received data and closing
func newPeeredConnection(conn net.Conn, peer *Peer) *peeredConnection {
	bconn := &peeredConnection{
		BufferedConnection: buffconn.NewBufferedConnection(conn, payload.MaxMessageSize),
		peer:               peer, // may be nil
	}
	bconn.Events.ReceiveMessage.Attach(events.NewClosure(func(data []byte) {
		bconn.receiveData(data)
	}))
	bconn.Events.Close.Attach(events.NewClosure(func() {
		if bconn.peer != nil {
			bconn.peer.Lock()
			bconn.peer.peerconn = nil
			bconn.peer.handshakeOk = false
			bconn.peer.Unlock()
			log.Debugw("closed buff connection", "conn", conn.RemoteAddr().String())
		}
	}))
	return bconn
}

// receive data handler for peered connection
func (bconn *peeredConnection) receiveData(data []byte) {
	msg, err := decodeMessage(data)
	if err != nil {
		log.Errorf("peeredConnection.receiveData: %v", err)
		if bconn.peer != nil {
			// may not be peered yet
			bconn.peer.closeConn()
		}
		return
	}
	if msg.MsgType == MsgTypeMsgChunk {
		finalMsg, err := chopper.IncomingChunk(msg.MsgData, payload.MaxMessageSize-chunkMessageOverhead)
		if err != nil {
			log.Errorf("peeredConnection.receiveData: %v", err)
			return
		}
		if finalMsg != nil {
			bconn.receiveData(finalMsg)
		}
	}
	if bconn.peer != nil {
		// it is peered but maybe not handshaked yet (can only be outbound)
		if bconn.peer.handshakeOk {
			// it is handshake-ed
			EventPeerMessageReceived.Trigger(msg)
		} else {
			// expected handshake msg
			if msg.MsgType != MsgTypeHandshake {
				log.Errorf("peeredConnection.receiveData: unexpected message during handshake")
				return
			}
			// not handshaked => do handshake
			bconn.processHandShakeOutbound(msg)
		}
	} else {
		// can only be inbound
		// expected handshake msg
		if msg.MsgType != MsgTypeHandshake {
			log.Errorf("peeredConnection.receiveData: unexpected message during handshake")
			return
		}
		// not peered yet can be only inbound
		// peer up and do handshake
		bconn.processHandShakeInbound(msg)
	}
}

// receives handshake response from the outbound peer
// assumes the connection is already peered (i can be only for outbound peers)
func (bconn *peeredConnection) processHandShakeOutbound(msg *PeerMessage) {
	id := string(msg.MsgData)
	log.Debugf("received handshake from outbound %s", id)
	if id != bconn.peer.PeeringId() {
		log.Error("closeConn the peer connection: wrong handshake message from outbound peer: expected %s got '%s'",
			bconn.peer.PeeringId(), id)
		if bconn.peer != nil {
			// may ne be peered yet
			bconn.peer.closeConn()
		}
	} else {
		log.Infof("CONNECTED WITH PEER %s (outbound)", id)
		bconn.peer.handshakeOk = true
	}
}

// receives handshake from the inbound peer
// links connection with the peer
// sends response back to finish the handshake
func (bconn *peeredConnection) processHandShakeInbound(msg *PeerMessage) {
	peeringId := string(msg.MsgData)
	log.Debugf("received handshake from inbound id = %s", peeringId)

	peersMutex.RLock()
	peer, ok := peers[peeringId]
	peersMutex.RUnlock()

	if !ok || !peer.isInbound() {
		log.Debugf("inbound connection from unexpected peer id %s. Closing..", peeringId)
		_ = bconn.Close()
		return
	}
	bconn.peer = peer

	peer.Lock()
	peer.peerconn = bconn
	peer.handshakeOk = true
	peer.Unlock()

	log.Infof("CONNECTED WITH PEER %s (inbound)", peeringId)

	if err := peer.sendHandshake(); err != nil {
		log.Error("error while responding to handshake: %v. Closing connection", err)
		_ = bconn.Close()
	}
}
