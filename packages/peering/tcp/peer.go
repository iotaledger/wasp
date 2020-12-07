package tcp

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/chopper"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/payload"
	"github.com/iotaledger/hive.go/backoff"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"go.dedis.ch/kyber/v3"
	"go.uber.org/atomic"
)

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries)) // TODO: Global variables.

// peer represents point-to-point TCP connection between two nodes and another
// it is used as transport for message exchange
// Another end is always using the same connection
// the peer takes care about exchanging heartbeat messages.
// It keeps last several received heartbeats as "lad" data to be able to calculate how synced/unsynced
// clocks of peer are.
type peer struct {
	*sync.RWMutex
	isDismissed atomic.Bool       // to be GC-ed
	peerconn    *peeredConnection // nil means not connected
	handshakeOk bool

	remoteNetID string // network locations as taken from the SC data

	startOnce *sync.Once
	numUsers  int
	net       *NetImpl
	log       *logger.Logger
}

func newPeer(remoteNetID string, net *NetImpl) *peer {
	return &peer{
		RWMutex:     &sync.RWMutex{},
		remoteNetID: remoteNetID,
		startOnce:   &sync.Once{},
		numUsers:    1,
		net:         net,
		log:         net.log,
	}
}

// Location implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
func (p *peer) Location() string {
	return p.remoteNetID
}

// PubKey implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
func (p *peer) PubKey() kyber.Point {
	return nil // TODO: [KP] Get it on handshake.
}

// SendMsg implements peering.PeerSender interface for the remote peers.
func (p *peer) SendMsg(msg *peering.PeerMessage) {
	if err := p.doSendMsg(msg); err != nil {
		// Async sending, we should ignore the errors.
		p.log.Warnf("Failed to send a message, reason: %v", err)
	}
}

// IsAlive implements peering.PeerSender and peering.PeerStatusProvider interfaces for the remote peers.
// Return true if is alive and average latencyRingBuf in nanosec.
func (p *peer) IsAlive() bool {
	p.RLock()
	defer p.RUnlock()
	return p.peerconn != nil && p.handshakeOk
}

// IsInbound implements peering.PeerStatusProvider.
// It is used in the dashboard.
func (p *peer) IsInbound() bool {
	return p.net.isInbound(p.remoteNetID)
}

// IsInbound implements peering.PeerStatusProvider.
// It is used in the dashboard.
func (p *peer) NumUsers() int {
	p.RLock()
	defer p.RUnlock()
	return p.numUsers
}

// SendMsg implements peering.PeerSender interface for the remote peers.
func (p *peer) Close() {
	p.net.stopUsingPeer(p.remoteNetID)
}

func (p *peer) peeringID() string {
	return p.net.peeringID(p.remoteNetID)
}

func (p *peer) connStatus() (bool, bool) {
	p.RLock()
	defer p.RUnlock()
	if p.isDismissed.Load() {
		return false, false
	}
	return p.peerconn != nil, p.handshakeOk
}

func (p *peer) closeConn() {
	p.Lock()
	defer p.Unlock()

	if p.isDismissed.Load() {
		return
	}
	if p.peerconn != nil {
		_ = p.peerconn.Close()
	}
}

// dials outbound address and established connection
func (p *peer) runOutbound() {
	log := p.net.log
	if p.isDismissed.Load() {
		return
	}
	if p.IsInbound() {
		return
	}
	if p.peerconn != nil {
		panic("peer.peerconn != nil")
	}
	log.Debugf("runOutbound %s", p.remoteNetID)

	// always try to reconnect
	defer func() {
		go func() {
			time.Sleep(restartAfter)
			p.Lock()
			if !p.isDismissed.Load() {
				p.startOnce = &sync.Once{}
				log.Debugf("will run again: %s", p.peeringID())
			}
			p.Unlock()
		}()
	}()

	var conn net.Conn

	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		conn, err = net.DialTimeout("tcp", p.remoteNetID, dialTimeout)
		if err != nil {
			return fmt.Errorf("dial %s failed: %w", p.remoteNetID, err)
		}
		return nil
	}); err != nil {
		log.Warn(err)
		return
	}
	p.peerconn = newPeeredConnection(conn, p.net, p)
	if err := p.sendHandshake(); err != nil {
		log.Errorf("error during sendHandshake: %v", err)
		return
	}
	log.Debugf("starting reading outbound %s", p.remoteNetID)
	err := p.peerconn.Read()
	log.Debugw("stopped reading outbound. Closing", "remote", p.remoteNetID, "err", err)
	p.closeConn()
}

// sends handshake message. It contains myLocation
func (p *peer) sendHandshake() error {
	data := encodeMessage(&peering.PeerMessage{
		MsgType: msgTypeHandshake,
		MsgData: []byte(p.peeringID()),
	}, time.Now().UnixNano())
	_, err := p.peerconn.Write(data)
	p.net.log.Debugf("sendHandshake '%s' --> '%s', id = %s", p.net.myNetID, p.remoteNetID, p.peeringID())
	return err
}

func (p *peer) doSendMsg(msg *peering.PeerMessage) error {
	if msg.MsgType < peering.FirstCommitteeMsgCode {
		return errors.New("reserved message code")
	}
	ts := msg.Timestamp
	if ts == 0 {
		ts = time.Now().UnixNano()
	}
	data := encodeMessage(msg, ts)

	choppedData, chopped := chopper.ChopData(data, payload.MaxMessageSize-chunkMessageOverhead)

	p.RLock()
	defer p.RUnlock()

	if !chopped {
		return p.sendData(data)
	}
	return p.sendChunks(choppedData)
}

func (p *peer) sendChunks(chopped [][]byte) error {
	ts := time.Now().UnixNano()
	for _, piece := range chopped {
		d := encodeMessage(&peering.PeerMessage{
			MsgType: msgTypeMsgChunk,
			MsgData: piece,
		}, ts)
		if err := p.sendData(d); err != nil {
			return err
		}
	}
	return nil
}

// SendMsgToPeers sends same msg to all peers in the slice which are not nil
// with the same timestamp
// return number of successfully sent messages and timestamp
func SendMsgToPeers(msg *peering.PeerMessage, ts int64, peers ...*peer) uint16 { // TODO: [KP] Remove, unused.
	if msg.MsgType < peering.FirstCommitteeMsgCode {
		return 0
	}
	// timestamped here, once
	data := encodeMessage(msg, ts)
	choppedData, chopped := chopper.ChopData(data, payload.MaxMessageSize-chunkMessageOverhead)

	numSent := uint16(0)
	for _, peer := range peers {
		if peer == nil {
			continue
		}
		peer.RLock()
		if !chopped {
			if err := peer.sendData(data); err == nil {
				numSent++
			}
		} else {
			if err := peer.sendChunks(choppedData); err == nil {
				numSent++
			}
		}
		peer.RUnlock()
	}
	return numSent
}

func (p *peer) sendData(data []byte) error {
	if p.peerconn == nil {
		return fmt.Errorf("no connection with %s", p.remoteNetID)
	}
	num, err := p.peerconn.Write(data)
	if num != len(data) {
		return fmt.Errorf("not all bytes were written. err = %v", err)
	}
	return nil
}
