// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package lpp

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
)

const (
	inactiveDeadline = 1 * time.Minute
	inactivePingTime = 30 * time.Second
)

type peer struct {
	remoteNetID    string
	remotePubKey   *ed25519.PublicKey
	remoteLppID    libp2ppeer.ID
	accessLock     *sync.RWMutex
	sendCh         chan *peering.PeerMessage
	sendChOverflow atomic.Uint32
	lastMsgSent    time.Time
	lastMsgRecv    time.Time
	numUsers       int
	trusted        bool
	net            *netImpl
	log            *logger.Logger
}

func newPeer(remoteNetID string, remotePubKey *ed25519.PublicKey, remoteLppID libp2ppeer.ID, n *netImpl) *peer {
	log := n.log.Named("peer:" + remoteNetID)
	p := &peer{
		remoteNetID:  remoteNetID,
		remotePubKey: remotePubKey,
		remoteLppID:  remoteLppID,
		accessLock:   &sync.RWMutex{},
		sendCh:       make(chan *peering.PeerMessage, 100),
		lastMsgSent:  time.Time{},
		lastMsgRecv:  time.Time{},
		numUsers:     0,
		trusted:      true,
		net:          n,
		log:          log,
	}
	go p.sendLoop()
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
		p.net.delPeer(p)
		close(p.sendCh)
	}
}

// NetID implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
func (p *peer) NetID() string {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remoteNetID
}

// PubKey implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
// This function tries to await for the public key to be resolves for some time, but with no guarantees.
func (p *peer) PubKey() *ed25519.PublicKey {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remotePubKey
}

// SendMsg implements peering.PeerSender interface for the remote peers.
// The send operation is performed asynchronously.
// The async sending helped to cope with sporadic deadlocks.
func (p *peer) SendMsg(msg *peering.PeerMessage) {
	//
	p.accessLock.RLock()
	if !p.trusted {
		p.log.Infof("Dropping outgoing message, because it was meant to send to a distrusted peer.")
		p.accessLock.RUnlock()
		return
	}
	p.accessLock.RUnlock()
	catch := func() {
		err := recover()
		if err == "send on closed channel" {
			p.log.Warnf("Failed to send message, reason=%v", err)
			return
		}
		if err != nil {
			p.log.Errorf("Failed to send message, reason=%v", err)
			panic(err)
		}
	}
	//
	defer catch()
	select {
	case p.sendCh <- msg:
		return
	default:
		overflow := p.sendChOverflow.Inc()
		p.log.Warnf("Send channel to %v overflown by %v", p.remoteNetID, overflow)
		go func() {
			defer catch()
			p.sendCh <- msg
			if overflow := p.sendChOverflow.Dec(); overflow == 0 {
				p.log.Warnf("Send channel to %v is not overflown anymore.", p.remoteNetID)
			}
		}()
	}
}

func (p *peer) sendLoop() {
	for msg := range p.sendCh {
		p.sendMsgDirect(msg)
	}
}

func (p *peer) sendMsgDirect(msg *peering.PeerMessage) {
	stream, err := p.net.lppHost.NewStream(p.net.ctx, p.remoteLppID, lppProtocolPeering)
	if err != nil {
		p.log.Warnf("Failed to send outgoing message, unable to allocate stream, reason=%v", err)
		return
	}
	defer stream.Close()
	//
	msgBytes, err := msg.Bytes(p.net.nodeKeyPair)
	if err != nil {
		p.log.Warnf("Failed to send outgoing message, unable to serialize, reason=%v", err)
		return
	}
	if err := writeFrame(stream, msgBytes); err != nil {
		p.log.Warnf("Failed to send outgoing message to , send failed with reason=%v", p.remoteNetID, err)
		return
	}
	p.accessLock.Lock()
	p.lastMsgSent = time.Now()
	p.accessLock.Unlock()
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
	return xerrors.New("Peer not trusted.")
}

// IsInbound implements peering.PeerStatusProvider.
// It is used in the dashboard.
func (p *peer) IsInbound() bool {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remoteNetID < p.net.myNetID
}

// IsInbound implements peering.PeerStatusProvider.
// It is used in the dashboard.
func (p *peer) NumUsers() int {
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.numUsers
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

func (p *peer) setNetID(netID string) {
	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.remoteNetID = netID
}
