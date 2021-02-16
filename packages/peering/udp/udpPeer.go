// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/chopper"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
)

const (
	inactiveDeadline   = 1 * time.Minute
	inactivePingTime   = 30 * time.Second
	sendMsgSyncTimeout = 3 * time.Second

	maxChunkSize = 508 // Safe payload size for UDP.
)

type peer struct {
	remoteNetID   string
	remotePubKey  kyber.Point
	remoteUDPAddr *net.UDPAddr
	waitReady     *util.WaitChan
	accessLock    *sync.RWMutex
	lastMsgSent   time.Time
	lastMsgRecv   time.Time
	numUsers      int
	msgChopper    *chopper.Chopper
	net           *NetImpl
	log           *logger.Logger
}

func newPeerOnUserRequest(remoteNetID string, n *NetImpl) (*peer, error) {
	var err error
	var remoteUDPAddr *net.UDPAddr
	if remoteUDPAddr, err = net.ResolveUDPAddr("udp", remoteNetID); err != nil {
		return nil, err
	}
	var p *peer
	if p, err = newPeer(remoteNetID, remoteUDPAddr, n); err != nil {
		return nil, err
	}
	p.usePeer()
	return p, nil
}

func newPeerFromHandshake(handshake *handshakeMsg, remoteUDPAddr *net.UDPAddr, n *NetImpl) (*peer, error) {
	var err error
	var p *peer
	if p, err = newPeer(handshake.netID, remoteUDPAddr, n); err != nil {
		return nil, err
	}
	if oldUDPAddrStr, newUDPAddrStr := p.handleHandshake(handshake, remoteUDPAddr); oldUDPAddrStr != newUDPAddrStr {
		return nil, errors.New("inconsistent_udp_addr_on_create")
	}
	return p, nil
}

// That's internal, called from other constructors.
func newPeer(remoteNetID string, remoteUDPAddr *net.UDPAddr, n *NetImpl) (*peer, error) {
	var log = n.log.Named("peer:" + remoteNetID)
	p := &peer{
		remoteNetID:   remoteNetID,
		remotePubKey:  nil, // Will be retrieved on handshake.
		remoteUDPAddr: remoteUDPAddr,
		waitReady:     util.NewWaitChan(),
		accessLock:    &sync.RWMutex{},
		lastMsgSent:   time.Time{},
		lastMsgRecv:   time.Time{},
		numUsers:      0,
		msgChopper:    chopper.NewChopper(),
		net:           n,
		log:           log,
	}
	p.sendHandshake(true)
	return p, nil
}

func (p *peer) usePeer() {
	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.numUsers++
}

func (p *peer) handleHandshake(handshake *handshakeMsg, remoteUDPAddr *net.UDPAddr) (string, string) {
	p.accessLock.Lock()
	oldUDPAddrStr := p.remoteUDPAddr.String()
	newUDPAddrStr := remoteUDPAddr.String()
	if oldUDPAddrStr != newUDPAddrStr {
		p.log.Warnf("Remote UDPAddr has changed, old=%v, new=%v", oldUDPAddrStr, newUDPAddrStr)
		p.remoteUDPAddr = remoteUDPAddr
	}
	if p.remotePubKey == nil {
		// That's the first received handshake, pairing established.
		p.remotePubKey = handshake.pubKey
		p.waitReady.Done()
		p.log.Infof("Paired %v with %v", p.net.NetID(), p.remoteNetID)
	} else if p.remotePubKey != nil && p.remotePubKey.Equal(handshake.pubKey) {
		// It's just a ping.
	} else {
		// New PublicKey is used by the peer!
		if !p.remotePubKey.Equal(handshake.pubKey) {
			p.log.Warnf("Remote PubKey has changed, old=%v, new=%v", p.remotePubKey, handshake.pubKey)
		}
		p.remotePubKey = handshake.pubKey
	}
	p.lastMsgRecv = time.Now()
	p.accessLock.Unlock()
	if handshake.respond {
		// Respond to the handshake, if asked.
		p.sendHandshake(false)
	}
	return oldUDPAddrStr, newUDPAddrStr
}

func (p *peer) sendHandshake(respond bool) {
	var err error
	handshake := handshakeMsg{
		netID:   p.net.NetID(),
		pubKey:  p.net.PubKey(),
		respond: respond,
	}
	var msgDataBin []byte
	if msgDataBin, err = handshake.bytes(p.net.nodeKeyPair.Private, p.net.suite); err != nil {
		p.log.Errorf("Unable to encode outgoing handshake msg, reason=%v", err)
	}
	p.SendMsg(&peering.PeerMessage{
		Timestamp: time.Now().UnixNano(),
		MsgType:   peering.MsgTypeHandshake,
		MsgData:   msgDataBin,
	})
}

func (p *peer) noteReceived() {
	p.accessLock.Lock()
	p.lastMsgRecv = time.Now()
	p.accessLock.Unlock()
}

// Send pings, if needed. Other periodic actions can be added here.
func (p *peer) maintenanceCheck() {
	now := time.Now()
	old := now.Add(-inactivePingTime)
	p.accessLock.RLock()
	if p.numUsers > 0 && p.lastMsgRecv.Before(old) {
		p.accessLock.RUnlock()
		p.sendHandshake(true)
	} else {
		p.accessLock.RUnlock()
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
func (p *peer) PubKey() kyber.Point {
	_ = p.waitReady.WaitTimeout(sendMsgSyncTimeout)
	p.accessLock.RLock()
	defer p.accessLock.RUnlock()
	return p.remotePubKey
}

// SendMsg implements peering.PeerSender interface for the remote peers.
func (p *peer) SendMsg(msg *peering.PeerMessage) {
	var err error
	var msgChunks [][]byte
	if msg.IsUserMessage() {
		if !p.waitReady.WaitTimeout(sendMsgSyncTimeout) {
			// Just log a warning and try to send a message anyway.
			p.log.Warn("Sending a message despite the peering is not established yet, MsgType=%v", msg.MsgType)
		}
	}
	if msgChunks, err = msg.ChunkedBytes(maxChunkSize, p.msgChopper); err != nil {
		p.log.Warnf("Dropping outgoing message, unable to encode, reason=%v", err)
		return
	}
	for i := range msgChunks {
		var n int
		if n, err = p.net.myUDPConn.WriteTo(msgChunks[i], p.remoteUDPAddr); err != nil {
			p.log.Warnf("Dropping outgoing message, unable to send, reason=%v", err)
			return
		}
		if n != len(msgChunks[i]) {
			p.log.Warnf("Partial message sent, sent=%v, msgBin=%v", n, len(msgChunks[i]))
			return
		}
	}

	p.accessLock.Lock()
	defer p.accessLock.Unlock()
	p.lastMsgSent = time.Now()
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
	if p.waitReady.WaitTimeout(timeout) {
		return nil
	}
	return fmt.Errorf("timeout waiting for %v to become ready", p.remoteNetID)
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
