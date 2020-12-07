package tcp

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"net"

	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/chopper"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/payload"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/labstack/gommon/log"
)

// extension of BufferedConnection from hive.go
// BufferedConnection is a wrapper for net.Conn
// peeredConnection first handles handshake and then links
// with peer according to the handshake information
type peeredConnection struct {
	*buffconn.BufferedConnection
	peer        *peer
	net         *NetImpl
	handshakeOk bool
}

// creates new peered connection and attach event handlers for received data and closing
func newPeeredConnection(conn net.Conn, net *NetImpl, peer *peer) *peeredConnection {
	c := &peeredConnection{
		BufferedConnection: buffconn.NewBufferedConnection(conn, payload.MaxMessageSize),
		peer:               peer, // may be nil
		net:                net,
	}
	c.Events.ReceiveMessage.Attach(events.NewClosure(func(data []byte) {
		c.receiveData(data)
	}))
	c.Events.Close.Attach(events.NewClosure(func() {
		if c.peer != nil {
			c.peer.Lock()
			c.peer.peerconn = nil
			c.peer.handshakeOk = false
			c.peer.Unlock()
		}
		net.log.Debugw("closed buff connection", "conn", conn.RemoteAddr().String())
	}))
	return c
}

// receive data handler for peered connection
func (c *peeredConnection) receiveData(data []byte) {
	msg, err := decodeMessage(data)
	if err != nil {
		// gross violation of the protocol
		log.Errorf("!!!!! peeredConnection.receiveData.decodeMessage: %v", err)
		c.Close()
		return
	}
	if msg.MsgType == msgTypeMsgChunk {
		finalMsg, err := chopper.IncomingChunk(msg.MsgData, payload.MaxMessageSize-chunkMessageOverhead)
		if err != nil {
			log.Errorf("peeredConnection.receiveData: %v", err)
			return
		}
		if finalMsg != nil {
			c.receiveData(finalMsg)
		}
	}
	if c.peer != nil {
		// it is peered but maybe not handshaked yet (can only be outbound)
		if c.peer.handshakeOk {
			// it is handshake-ed
			c.net.events.Trigger(&peering.RecvEvent{
				From: c.peer,
				Msg:  msg,
			})
		} else {
			// expected handshake msg
			if msg.MsgType != msgTypeHandshake {
				log.Errorf("peeredConnection.receiveData: unexpected message during handshake 1")
				return
			}
			// not handshaked => do handshake
			c.processHandShakeOutbound(msg)
		}
	} else {
		// can only be inbound
		// expected handshake msg
		if msg.MsgType != msgTypeHandshake {
			log.Errorf("peeredConnection.receiveData: unexpected message during handshake 2")
			return
		}
		// not peered yet can be only inbound
		// peer up and do handshake
		c.processHandShakeInbound(msg)
	}
}

// receives handshake response from the outbound peer
// assumes the connection is already peered (i can be only for outbound peers)
func (c *peeredConnection) processHandShakeOutbound(msg *peering.PeerMessage) {
	id := string(msg.MsgData)
	log.Debugf("received handshake from outbound %s", id)
	if id != c.peer.peeringID() {
		log.Errorf(
			"closeConn the peer connection: wrong handshake message from outbound peer: expected %s got '%s'",
			c.peer.peeringID(), id,
		)
		if c.peer != nil {
			// may ne be peered yet
			c.peer.closeConn()
		}
	} else {
		log.Infof("CONNECTED WITH PEER %s (outbound)", id)
		c.peer.handshakeOk = true
	}
}

// receives handshake from the inbound peer
// links connection with the peer
// sends response back to finish the handshake
func (c *peeredConnection) processHandShakeInbound(msg *peering.PeerMessage) {
	peeringID := string(msg.MsgData)
	log.Debugf("received handshake from inbound id = %s", peeringID)

	c.net.peersMutex.RLock()
	peer, ok := c.net.peers[peeringID]
	c.net.peersMutex.RUnlock()

	if !ok || !peer.isInbound() {
		log.Debugf("inbound connection from unexpected peer id %s. Closing..", peeringID)
		_ = c.Close()
		return
	}
	c.peer = peer

	peer.Lock()
	peer.peerconn = c
	peer.handshakeOk = true
	peer.Unlock()

	log.Infof("CONNECTED WITH PEER %s (inbound)", peeringID)

	if err := peer.sendHandshake(); err != nil {
		log.Errorf("error while responding to handshake: %v. Closing connection", err)
		_ = c.Close()
	}
}
