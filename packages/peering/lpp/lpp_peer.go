// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package lpp

import (
	"errors"
	"sync"
	"time"

	libp2ppeer "github.com/libp2p/go-libp2p/core/peer"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

const (
	inactiveDeadline = 1 * time.Minute
	inactivePingTime = 30 * time.Second
	maxPeerMsgBuffer = 10000
	traceMessages    = false
)

type peer struct {
	name             string
	remotePeeringURL string
	remotePubKey     *cryptolib.PublicKey
	remoteLppID      libp2ppeer.ID
	accessLock       *sync.RWMutex
	sendPipe         pipe.Pipe[*peering.PeerMessageNet]
	recvPipe         pipe.Pipe[*peering.PeerMessageNet]
	lastMsgSent      time.Time
	lastMsgRecv      time.Time
	numUsers         int
	trusted          bool
	net              *netImpl
	log              log.Logger
}

var _ peering.PeerSender = &peer{}

func newPeer(name, peeringURL string, remotePubKey *cryptolib.PublicKey, remoteLppID libp2ppeer.ID, n *netImpl) *peer {
	log := n.log.NewChildLogger("peer:" + peeringURL)
	messagePriorityFun := func(msg *peering.PeerMessageNet) bool {
		return false
	}
	p := &peer{
		name:             name,
		remotePeeringURL: peeringURL,
		remotePubKey:     remotePubKey,
		remoteLppID:      remoteLppID,
		accessLock:       &sync.RWMutex{},
		sendPipe:         pipe.NewLimitPriorityHashInfinitePipe(messagePriorityFun, maxPeerMsgBuffer),
		recvPipe:         pipe.NewLimitPriorityHashInfinitePipe(messagePriorityFun, maxPeerMsgBuffer),
		lastMsgSent:      time.Time{},
		lastMsgRecv:      time.Time{},
		numUsers:         0,
		trusted:          true,
		net:              n,
		log:              log,
	}
	go p.sendLoop()
	go p.recvLoop()
	return p
}

func (p *peer) usePeer() {
	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.numUsers++
}

func (p *peer) noteReceived() {
	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.lastMsgRecv = time.Now()
}

// Send pings, if needed. Other periodic actions can be added here.
func (p *peer) maintenanceCheck() {
	now := time.Now()
	old := now.Add(-inactivePingTime)

	p.accessLock.RLock()
	numUsers := p.numUsers
	lastMsgOld := p.lastMsgRecv.Before(old)
	trusted := p.trusted
	p.accessLock.RUnlock()

	if numUsers > 0 && lastMsgOld {
		p.net.lppHeartbeatSend(p, true)
	}
	if numUsers == 0 && !trusted && lastMsgOld {
		p.net.delPeerWithoutLock(p)
		p.sendPipe.Close()
		p.recvPipe.Close()
	}
}

// PeeringURL implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
func (p *peer) Name() string {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.name
}

// PeeringURL implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
func (p *peer) PeeringURL() string {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remotePeeringURL
}

// PubKey implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
// This function tries to await for the public key to be resolves for some time, but with no guarantees.
func (p *peer) PubKey() *cryptolib.PublicKey {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remotePubKey
}

// SendMsg implements peering.PeerSender interface for the remote peers.
// The send operation is performed asynchronously.
// The async sending helped to cope with sporadic deadlocks.
func (p *peer) SendMsg(msg *peering.PeerMessageData) {
	//
	p.accessLock.RLock()
	msgNet := &peering.PeerMessageNet{PeerMessageData: msg}
	if !p.trusted {
		p.log.LogInfof("Dropping outgoing message, because it was meant to send to a distrusted peer.")
		p.accessLock.RUnlock()
		return
	}
	p.accessLock.RUnlock()
	p.sendPipe.In() <- msgNet
	p.net.metrics.SendEnqueued(len(msg.MsgData), p.sendPipe.Len())
}

func (p *peer) RecvMsg(msg *peering.PeerMessageNet) {
	if traceMessages {
		p.log.LogDebugf("Peer message received from peer %v, peeringID %v, receiver %v, type %v, length %v, first bytes %v",
			p.PeeringURL(), msg.PeeringID, msg.MsgReceiver, msg.MsgType, len(msg.MsgData), firstBytes(16, msg.MsgData))
	}
	p.noteReceived()
	p.recvPipe.In() <- msg
	p.net.metrics.RecvEnqueued(len(msg.MsgData), p.recvPipe.Len())
}

func (p *peer) sendLoop() {
	for msg := range p.sendPipe.Out() {
		p.net.metrics.SendDequeued(len(msg.MsgData), p.sendPipe.Len())
		p.sendMsgDirect(msg)
	}
}

func (p *peer) recvLoop() {
	for msg := range p.recvPipe.Out() {
		p.net.metrics.RecvDequeued(len(msg.MsgData), p.recvPipe.Len())
		p.net.triggerRecvEvents(p.PubKey(), msg)
	}
}

func (p *peer) sendMsgDirect(msg *peering.PeerMessageNet) {
	stream, err := p.net.lppHost.NewStream(p.net.ctx, p.remoteLppID, lppProtocolPeering)
	if err != nil {
		p.log.LogWarnf("Failed to send outgoing message, unable to allocate stream, reason=%v", err)
		return
	}
	defer stream.Close()
	//
	msgBytes := msg.Bytes() // Do not use msg signatures, we are using TLS.
	if err := writeFrame(stream, msgBytes); err != nil {
		p.log.LogWarnf("Failed to send outgoing message to %s, send failed with reason=%v", p.remotePeeringURL, err)
		return
	}
	p.accessLock.Lock()
	p.lastMsgSent = time.Now()
	p.accessLock.Unlock()
	if traceMessages {
		p.log.LogDebugf("Peer message sent to peer %v, peeringID %v, receiver %v, type %v, length %v, first bytes %v",
			p.PeeringURL(), msg.PeeringID, msg.MsgReceiver, msg.MsgType, len(msg.MsgData), firstBytes(16, msg.MsgData))
	}
}

func firstBytes(maxCount int, array []byte) []byte {
	if len(array) <= maxCount {
		return array
	}
	return array[:maxCount]
}

// IsAlive implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
// Return true if is alive and average latencyRingBuf in nanosec.
func (p *peer) IsAlive() bool {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remotePubKey != nil && p.lastMsgRecv.After(time.Now().Add(-inactiveDeadline))
}

// Await implements peering.PeerSender interface for the remote peers.
func (p *peer) Await(timeout time.Duration) error {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	if p.trusted {
		return nil
	}
	return errors.New("peer not trusted")
}

// IsInbound implements peering.PeerStatusProvider.
// It is used in the dashboard.
func (p *peer) IsInbound() bool {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remotePeeringURL < p.net.myPeeringURL
}

// NumUsers implements peering.PeerStatusProvider.
// It is used in the dashboard.
func (p *peer) NumUsers() int {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.numUsers
}

// Status implements peering.PeerSender interface for the remote peers.
func (p *peer) Status() peering.PeerStatusProvider {
	return p
}

// SendMsg implements peering.PeerSender interface for the remote peers.
func (p *peer) Close() {
	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.numUsers--
}

func (p *peer) trust(trusted bool) {
	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.trusted = trusted
}

func (p *peer) setPeeringURL(url string) {
	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.remotePeeringURL = url
}
