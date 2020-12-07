package tcp

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"errors"
	"sync"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/group"
	"go.dedis.ch/kyber/v3"
)

// NetImpl implements a peering.NetworkProvider interface.
type NetImpl struct {
	myNetID string // NetID of this node.
	port    int    // Port to use for peering.

	peers      map[string]*peer
	peersMutex *sync.RWMutex
	events     *events.Event

	log *logger.Logger
}

// NewNetworkProvider is a constructor for the TCP based
// peering network implementation.
func NewNetworkProvider(myNetID string, port int, log *logger.Logger) (*NetImpl, error) {
	if err := checkMyNetID(myNetID, port); err != nil {
		// can't continue because NetID parameter is not correct
		log.Panicf("checkMyNetworkID: '%v'. || Check the 'netid' parameter in config.json", err)
		return nil, err
	}
	n := NetImpl{
		myNetID:    myNetID,
		port:       port,
		peers:      make(map[string]*peer),
		peersMutex: &sync.RWMutex{},
		log:        log,
	}
	n.events = events.NewEvent(n.eventHandler)
	return &n, nil
}

// A handler suitable for events.NewEvent().
func (n *NetImpl) eventHandler(handler interface{}, params ...interface{}) {
	callback := handler.(func(_ *peering.RecvEvent))
	recvEvent := params[0].(*peering.RecvEvent)
	callback(recvEvent)
}

// Run starts listening and communicating with the network.
func (n *NetImpl) Run(shutdownSignal <-chan struct{}) {
	go n.connectOutboundLoop()
	go n.connectInboundLoop()

	<-shutdownSignal

	n.log.Info("Closing all connections with peers...")
	n.closeAll()
	n.log.Info("Closing all connections with peers... done")
}

// Self implements peering.NetworkProvider.
func (n *NetImpl) Self() peering.PeerSender {
	return n
}

// Group implements peering.NetworkProvider.
func (n *NetImpl) Group(peerNetIDs []string) (peering.GroupProvider, error) {
	var err error
	peers := make([]peering.PeerSender, len(peerNetIDs))
	for i := range peerNetIDs {
		if peers[i], err = n.PeerByLocation(peerNetIDs[i]); err != nil {
			return nil, err
		}
	}
	return group.NewPeeringGroupProvider(n, peers), nil
}

// Attach implements peering.NetworkProvider.
func (n *NetImpl) Attach(chainID *coretypes.ChainID, callback func(recv *peering.RecvEvent)) interface{} {
	closure := events.NewClosure(func(recv *peering.RecvEvent) {
		if chainID == nil || chainID.Equal(&recv.Msg.ChainID) {
			callback(recv)
		}
	})
	n.events.Attach(closure)
	return closure
}

// Detach implements peering.NetworkProvider.
func (n *NetImpl) Detach(attachID interface{}) {
	switch closure := attachID.(type) {
	case *events.Closure:
		n.events.Detach(closure)
	default:
		panic("invalid_attach_id")
	}
}

// PeerByLocation implements peering.NetworkProvider.
func (n *NetImpl) PeerByLocation(peerLoc string) (peering.PeerSender, error) {
	if p := n.usePeer(peerLoc); p != nil {
		return p, nil
	}
	return n, nil // Self
}

// PeerByPubKey implements peering.NetworkProvider.
// NOTE: For now, only known nodes can be looked up by PubKey.
func (n *NetImpl) PeerByPubKey(peerPub kyber.Point) (peering.PeerSender, error) {
	for i := range n.peers {
		pk := n.peers[i].PubKey()
		if pk != nil && pk.Equal(peerPub) {
			return n.PeerByLocation(n.peers[i].Location())
		}
	}
	return nil, errors.New("known peer not found by pubKey")
}

// Location implements peering.PeerSender for the Self() node.
func (n *NetImpl) Location() string {
	return n.myNetID
}

// PubKey implements peering.PeerSender for the Self() node.
func (n *NetImpl) PubKey() kyber.Point {
	return nil // TODO
}

// SendMsg implements peering.PeerSender for the Self() node.
func (n *NetImpl) SendMsg(msg *peering.PeerMessage) {
	// Don't go via the network, if sending a message to self.
	n.events.Trigger(&peering.RecvEvent{
		From: n.Self(),
		Msg:  msg,
	})
}

// IsAlive implements peering.PeerSender for the Self() node.
func (n *NetImpl) IsAlive() bool {
	return true // This node is alive.
}

// Close implements peering.PeerSender for the Self() node.
func (n *NetImpl) Close() {
	// We will con close the connection of the own node.
}

func (n *NetImpl) isInbound(remoteLocation string) bool {
	// if remoteLocation == n.myNetID {	// TODO: [KP] Do we need this?
	// 	panic("remoteNetID == myLocation")
	// }
	return remoteLocation < n.myNetID
}
func (n *NetImpl) peeringID(remoteLocation string) string {
	if n.isInbound(remoteLocation) {
		return remoteLocation + "<" + n.myNetID
	}
	return n.myNetID + "<" + remoteLocation
}

// usePeer adds new connection to the peer pool
// if it already exists, returns existing.
// Return nil for for own netID
// connection added to the pool is picked by loops which will try to establish connection
func (n *NetImpl) usePeer(netID string) *peer {
	if netID == n.myNetID {
		// nil for itself
		return nil
	}
	n.peersMutex.Lock()
	defer n.peersMutex.Unlock()

	if peer, ok := n.peers[n.peeringID(netID)]; ok {
		// existing peer
		peer.numUsers++
		return peer
	}
	// new peer
	ret := newPeer(netID, n)
	n.peers[ret.peeringID()] = ret
	n.log.Debugf("added new peer id %s inbound = %v", ret.peeringID(), ret.isInbound())
	return ret
}

// stopUsingPeer decreases counter.
func (n *NetImpl) stopUsingPeer(peerID string) {
	n.peersMutex.Lock()
	defer n.peersMutex.Unlock()

	if peer, ok := n.peers[peerID]; ok {
		peer.numUsers--
		if peer.numUsers == 0 {
			peer.isDismissed.Store(true)

			go func() {
				n.peersMutex.Lock()
				defer n.peersMutex.Unlock()

				delete(n.peers, peerID)
				peer.closeConn()
			}()
		}
	}
}
