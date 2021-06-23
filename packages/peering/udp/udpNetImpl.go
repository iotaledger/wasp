// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package udp implements a UDP based peering.NetworkProvider.
package udp

import (
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/domain"
	"github.com/iotaledger/wasp/packages/peering/group"
)

const (
	maintenancePeriod    = 1 * time.Second
	recvQueueSize        = 1024 * 5
	recvBlockingDuration = 3 * time.Second
)

// NetImpl implements a peering.NetworkProvider interface.
type NetImpl struct {
	myNetID     string // NetID of this node.
	myUDPConn   *net.UDPConn
	port        int              // Port to use for peering.
	peers       map[string]*peer // By NetID
	peersByAddr map[string]*peer // By UDPAddr.String()
	peersLock   *sync.RWMutex
	recvEvents  *events.Event
	recvQueue   chan *peering.RecvEvent // A queue for received messages.
	nodeKeyPair ed25519.KeyPair
	trusted     peering.TrustedNetworkManager
	log         *logger.Logger
}

// NewNetworkProvider is a constructor for the TCP based
// peering network implementation.
func NewNetworkProvider(
	myNetID string,
	port int,
	nodeKeyPair ed25519.KeyPair,
	trusted peering.TrustedNetworkManager,
	log *logger.Logger,
) (*NetImpl, error) {
	var err error
	var myUDPConn *net.UDPConn
	if myUDPConn, err = net.ListenUDP("udp", &net.UDPAddr{Port: port}); err != nil {
		return nil, err
	}
	n := NetImpl{
		myNetID:     myNetID,
		myUDPConn:   myUDPConn,
		port:        port,
		peers:       make(map[string]*peer),
		peersByAddr: make(map[string]*peer),
		peersLock:   &sync.RWMutex{},
		recvEvents:  nil, // Initialized bellow.
		recvQueue:   make(chan *peering.RecvEvent, recvQueueSize),
		nodeKeyPair: nodeKeyPair,
		trusted:     trusted,
		log:         log,
	}
	n.recvEvents = events.NewEvent(n.eventHandler)
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
	queueRecvStopCh := make(chan bool)
	receiveStopCh := make(chan bool)
	maintenanceStopCh := make(chan bool)
	go n.queueRecvLoop(queueRecvStopCh)
	go n.receiveLoop(receiveStopCh)
	go n.maintenanceLoop(maintenanceStopCh)

	<-shutdownSignal
	close(maintenanceStopCh)
	close(receiveStopCh)
	close(queueRecvStopCh)
}

// Self implements peering.NetworkProvider.
func (n *NetImpl) Self() peering.PeerSender {
	return n
}

// Group creates peering.GroupProvider.
func (n *NetImpl) PeerGroup(peerNetIDs []string) (peering.GroupProvider, error) {
	var err error
	groupPeers := make([]peering.PeerSender, len(peerNetIDs))
	for i := range peerNetIDs {
		if groupPeers[i], err = n.usePeer(peerNetIDs[i]); err != nil {
			return nil, err
		}
	}
	return group.NewPeeringGroupProvider(n, groupPeers, n.log)
}

// Domain creates peering.PeerDomainProvider.
func (n *NetImpl) PeerDomain(peerNetIDs []string) (peering.PeerDomainProvider, error) {
	peers := make([]peering.PeerSender, 0, len(peerNetIDs))
	for _, nid := range peerNetIDs {
		if nid == n.Self().NetID() {
			continue
		}
		p, err := n.usePeer(nid)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return domain.NewPeerDomain(n, peers, n.log), nil
}

// Attach implements peering.NetworkProvider.
func (n *NetImpl) Attach(peeringID *peering.PeeringID, callback func(recv *peering.RecvEvent)) interface{} {
	closure := events.NewClosure(func(recv *peering.RecvEvent) {
		if peeringID == nil || *peeringID == recv.Msg.PeeringID {
			callback(recv)
		}
	})
	n.recvEvents.Attach(closure)
	return closure
}

// Detach implements peering.NetworkProvider.
func (n *NetImpl) Detach(attachID interface{}) {
	switch closure := attachID.(type) {
	case *events.Closure:
		n.recvEvents.Detach(closure)
	default:
		panic("invalid_attach_id")
	}
}

// PeerByNetID implements peering.NetworkProvider.
func (n *NetImpl) PeerByNetID(peerNetID string) (peering.PeerSender, error) {
	return n.usePeer(peerNetID)
}

// PeerByPubKey implements peering.NetworkProvider.
// NOTE: For now, only known nodes can be looked up by PubKey.
func (n *NetImpl) PeerByPubKey(peerPub *ed25519.PublicKey) (peering.PeerSender, error) {
	n.peersLock.RLock()
	defer n.peersLock.RUnlock()
	for i := range n.peers {
		pk := n.peers[i].PubKey()
		if pk != nil && *pk == *peerPub { // Compared as binaries.
			return n.PeerByNetID(n.peers[i].NetID())
		}
	}
	return nil, errors.New("known peer not found by pubKey")
}

// PeerStatus implements peering.NetworkProvider.
func (n *NetImpl) PeerStatus() []peering.PeerStatusProvider {
	n.peersLock.RLock()
	defer n.peersLock.RUnlock()
	peerStatus := make([]peering.PeerStatusProvider, 0)
	for i := range n.peers {
		peerStatus = append(peerStatus, n.peers[i])
	}
	return peerStatus
}

// NetID implements peering.PeerSender for the Self() node.
func (n *NetImpl) NetID() string {
	return n.myNetID
}

// PubKey implements peering.PeerSender for the Self() node.
func (n *NetImpl) PubKey() *ed25519.PublicKey {
	return &n.nodeKeyPair.PublicKey
}

// SendMsg implements peering.PeerSender for the Self() node.
func (n *NetImpl) SendMsg(msg *peering.PeerMessage) {
	// Don't go via the network, if sending a message to self.
	n.recvQueue <- &peering.RecvEvent{
		From: n.Self(),
		Msg:  msg,
	}
}

// IsAlive implements peering.PeerSender for the Self() node.
func (n *NetImpl) IsAlive() bool {
	return true // This node is alive.
}

// Await implements peering.PeerSender for the Self() node.
func (n *NetImpl) Await(timeout time.Duration) error {
	return nil // This node is alive immediately.
}

// Close implements peering.PeerSender for the Self() node.
func (n *NetImpl) Close() {
	// We will con close the connection of the own node.
}

// IsTrustedPeer implements the peering.TrustedNetworkManager interface.
func (n *NetImpl) IsTrustedPeer(pubKey ed25519.PublicKey) error {
	return n.trusted.IsTrustedPeer(pubKey)
}

// TrustPeer implements the peering.TrustedNetworkManager interface.
// It delegates everything to other implementation and updates the connections accordingly.
func (n *NetImpl) TrustPeer(pubKey ed25519.PublicKey, netID string) (*peering.TrustedPeer, error) {
	// TODO: Implement
	return n.trusted.TrustPeer(pubKey, netID)
}

// DistrustPeer implements the peering.TrustedNetworkManager interface.
// It delegates everything to other implementation and updates the connections accordingly.
func (n *NetImpl) DistrustPeer(pubKey ed25519.PublicKey) (*peering.TrustedPeer, error) {
	// TODO: Implement
	return n.trusted.DistrustPeer(pubKey)
}

// TrustedPeers implements the peering.TrustedNetworkManager interface.
func (n *NetImpl) TrustedPeers() ([]*peering.TrustedPeer, error) {
	return n.trusted.TrustedPeers()
}

func (n *NetImpl) usePeer(remoteNetID string) (peering.PeerSender, error) {
	var err error
	if remoteNetID == n.myNetID {
		return n, nil
	}
	n.peersLock.Lock()
	defer n.peersLock.Unlock()
	if p, ok := n.peers[remoteNetID]; ok {
		p.usePeer()
		return p, nil
	}
	var p *peer
	if p, err = newPeerOnUserRequest(remoteNetID, n); err != nil {
		return nil, err
	}
	n.peers[p.NetID()] = p
	n.peersByAddr[p.remoteUDPAddr.String()] = p
	return p, nil
}

func (n *NetImpl) queueRecvLoop(stopCh chan bool) {
	for {
		select {
		case <-stopCh:
			return
		case recvEvent, ok := <-n.recvQueue:
			if ok {
				n.recvEvents.Trigger(recvEvent)
			}
		}
	}
}

func (n *NetImpl) receiveLoop(stopCh chan bool) {
	var err error
	var buf = make([]byte, 2024)
	for {
		select { // Terminate the loop, if such request has been made.
		case <-stopCh:
			return
		default:
		}
		var peerUDPAddr *net.UDPAddr
		var recvDeadline = time.Now().Add(recvBlockingDuration)
		n.myUDPConn.SetReadDeadline(recvDeadline)
		if _, peerUDPAddr, err = n.myUDPConn.ReadFromUDP(buf); err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				// We need to limit the blocking to make graceful stop possible.
				continue
			}
			n.log.Warnf("Error while reading from UDP socket, reason=%v", err)
			continue
		}
		var peerMsg *peering.PeerMessage
		if peerMsg, err = peering.NewPeerMessageFromBytes(buf); err != nil {
			n.log.Warnf("Error while decoding a UDP message, reason=%v", err)
			continue
		}
		switch peerMsg.MsgType {
		case peering.MsgTypeReserved:
			// Nothing
		case peering.MsgTypeHandshake:
			var h *handshakeMsg
			if h, err = handshakeMsgFromBytes(peerMsg.MsgData); err != nil {
				n.log.Warnf("Error while decoding a UDP handshake, reason=%v", err)
				continue
			}
			// if err = n.trusted.IsTrustedPeer(h.pubKey); err != nil { // TODO: XXX: KP
			// 	n.log.Warnf("Dropping handshakeMsg from %v with pubKey=%v, error=%v", peerUDPAddr, h.pubKey, err)
			// }
			n.peersLock.Lock()
			if p, ok := n.peers[h.netID]; ok {
				if oldUDPAddrStr, newUDPAddrStr := p.handleHandshake(h, peerUDPAddr); oldUDPAddrStr != newUDPAddrStr {
					// Update the index to find the peer later on.
					n.peersByAddr[newUDPAddrStr] = p
					delete(n.peersByAddr, oldUDPAddrStr)
				}
			} else {
				if p, err = newPeerFromHandshake(h, peerUDPAddr, n); err != nil {
					n.log.Warnf("Error while creating a peer based on UDP handshake, reason=%v", err)
					n.peersLock.Unlock()
					continue
				}
				n.peers[p.NetID()] = p
				n.peersByAddr[p.remoteUDPAddr.String()] = p
			}
			n.peersLock.Unlock()
		case peering.MsgTypeMsgChunk:
			remoteUDPAddrStr := peerUDPAddr.String()
			n.peersLock.RLock()
			if p, ok := n.peersByAddr[remoteUDPAddrStr]; ok {
				n.peersLock.RUnlock()
				var reconstructedMsg *peering.PeerMessage
				if reconstructedMsg, err = peering.NewPeerMessageFromChunks(peerMsg.MsgData, maxChunkSize, p.msgChopper); err != nil {
					n.log.Warnf("Error while decoding chunked message, reason=%v", err)
					continue
				}
				if reconstructedMsg != nil {
					n.receiveUserMsg(reconstructedMsg, peerUDPAddr)
				}
			} else {
				n.peersLock.RUnlock()
				n.log.Warnf("Dropping received message from unknown peer=%v", remoteUDPAddrStr)
				continue
			}
		default:
			n.receiveUserMsg(peerMsg, peerUDPAddr)
		}
	}
}

func (n *NetImpl) receiveUserMsg(msg *peering.PeerMessage, peerUDPAddr *net.UDPAddr) {
	if !msg.IsUserMessage() {
		n.log.Warnf("Dropping received message, unexpected MsgType=%v", msg.MsgType)
		return
	}
	remoteUDPAddrStr := peerUDPAddr.String()
	n.peersLock.RLock()
	defer n.peersLock.RUnlock()

	if p, ok := n.peersByAddr[remoteUDPAddrStr]; ok {
		p.noteReceived()
		msg.SenderNetID = p.NetID()
		n.recvQueue <- &peering.RecvEvent{
			From: p,
			Msg:  msg,
		}
		return
	}
	n.log.Warnf("Dropping received message from unknown peer=%v", remoteUDPAddrStr)
}

func (n *NetImpl) maintenanceLoop(stopCh chan bool) {
	for {
		select {
		case <-time.After(maintenancePeriod):
			n.peersLock.RLock()
			for _, p := range n.peers {
				p.maintenanceCheck()
			}
			n.peersLock.RUnlock()
		case <-stopCh:
			return
		}
	}
}
