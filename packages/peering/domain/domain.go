// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
)

type DomainImpl struct {
	netProvider peering.NetworkProvider
	nodes       map[cryptolib.PublicKeyKey]peering.PeerSender
	permutation *util.Permutation16
	permPubKeys []*cryptolib.PublicKey
	peeringID   peering.PeeringID
	attachIDs   []interface{}
	log         *logger.Logger
	mutex       *sync.RWMutex
}

var _ peering.PeerDomainProvider = &DomainImpl{}

// NewPeerDomain creates a collection. Ignores self
func NewPeerDomain(netProvider peering.NetworkProvider, peeringID peering.PeeringID, initialNodes []peering.PeerSender, log *logger.Logger) *DomainImpl {
	ret := &DomainImpl{
		netProvider: netProvider,
		nodes:       make(map[cryptolib.PublicKeyKey]peering.PeerSender),
		permutation: nil, // Will be set in ret.reshufflePeers().
		permPubKeys: nil, // Will be set in ret.reshufflePeers().
		peeringID:   peeringID,
		attachIDs:   make([]interface{}, 0),
		log:         log,
		mutex:       &sync.RWMutex{},
	}
	for _, sender := range initialNodes {
		ret.nodes[sender.PubKey().AsKey()] = sender
	}
	ret.initPermPubKeys()
	return ret
}

func (d *DomainImpl) SendMsgByPubKey(pubKey *cryptolib.PublicKey, msgReceiver, msgType byte, msgData []byte) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	peer, ok := d.nodes[pubKey.AsKey()]
	if !ok {
		d.log.Warnf("SendMsgByPubKey: PubKey %v is not in the domain", pubKey.AsString())
		return
	}
	peer.SendMsg(&peering.PeerMessageData{
		PeeringID:   d.peeringID,
		MsgReceiver: msgReceiver,
		MsgType:     msgType,
		MsgData:     msgData,
	})
}

func (d *DomainImpl) GetRandomOtherPeers(upToNumPeers int) []*cryptolib.PublicKey {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if upToNumPeers > len(d.permPubKeys) {
		upToNumPeers = len(d.permPubKeys)
	}
	ret := make([]*cryptolib.PublicKey, upToNumPeers)
	for i := range ret {
		ret[i] = d.permPubKeys[d.permutation.NextNoCycles()]
	}
	return ret
}

func (d *DomainImpl) UpdatePeers(newPeerPubKeys []*cryptolib.PublicKey) {
	d.mutex.RLock()
	oldPeers := make(map[cryptolib.PublicKeyKey]peering.PeerSender) // A copy, to avoid keeping the lock.
	for k, v := range d.nodes {
		oldPeers[k] = v
	}
	d.mutex.RUnlock()
	nodes := make(map[cryptolib.PublicKeyKey]peering.PeerSender) // Will collect the new set of nodes.
	changed := false
	//
	// Add new peers.
	for _, newPeerPubKey := range newPeerPubKeys {
		if _, isOldPeer := oldPeers[newPeerPubKey.AsKey()]; isOldPeer {
			continue // Old peers will be retained bellow.
		}
		newPeerSender, err := d.netProvider.PeerByPubKey(newPeerPubKey)
		if err != nil {
			d.log.Warnf("Domain peer skipped for now, pubKey=%v not found, reason: %v", newPeerPubKey.AsString(), err)
			continue
		}
		changed = true
		nodes[newPeerSender.PubKey().AsKey()] = newPeerSender
		d.log.Infof("Domain peer added, pubKey=%v, netID=%v", newPeerSender.PubKey().AsString(), newPeerSender.NetID())
	}
	//
	// Remove peers that are not needed anymore and retain others.
	for _, oldPeer := range oldPeers {
		oldPeerDropped := true
		if oldPeer.PubKey().Equals(d.netProvider.Self().PubKey()) {
			// We retain the current node in the domain all the time.
			nodes[oldPeer.PubKey().AsKey()] = oldPeer
			oldPeerDropped = false
		} else {
			for _, newPeerPubKey := range newPeerPubKeys {
				if oldPeer.PubKey().Equals(newPeerPubKey) {
					nodes[oldPeer.PubKey().AsKey()] = oldPeer
					oldPeerDropped = false
					break
				}
			}
		}
		if oldPeerDropped {
			changed = true
			d.log.Infof("Domain peer removed, pubKey=%v, netID=%v", oldPeer.PubKey().AsString(), oldPeer.NetID())
		}
	}
	if changed {
		d.mutex.Lock()
		d.nodes = nodes
		d.initPermPubKeys()
		d.mutex.Unlock()
	}
}

func (d *DomainImpl) ReshufflePeers() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.permutation.Shuffle()
}

func (d *DomainImpl) initPermPubKeys() {
	d.permPubKeys = make([]*cryptolib.PublicKey, 0, len(d.nodes))
	for _, sender := range d.nodes {
		if !sender.PubKey().Equals(d.netProvider.Self().PubKey()) { // Do not include self to the permutation.
			d.permPubKeys = append(d.permPubKeys, sender.PubKey())
		}
	}
	var err error
	d.permutation, err = util.NewPermutation16(uint16(len(d.permPubKeys)))
	if err != nil {
		d.log.Warnf("Error generating permutation: %v", err)
	}
}

func (d *DomainImpl) Attach(receiver byte, callback func(recv *peering.PeerMessageIn)) interface{} {
	attachID := d.netProvider.Attach(&d.peeringID, receiver, func(recv *peering.PeerMessageIn) {
		if recv.SenderPubKey.Equals(d.netProvider.Self().PubKey()) {
			d.log.Debugf("dropping message for receiver=%v MsgType=%v from %v: message from self.",
				recv.MsgReceiver, recv.MsgType, recv.SenderPubKey.AsString())
			return
		}
		_, ok := d.nodes[recv.SenderPubKey.AsKey()]
		if !ok {
			d.log.Warnf("dropping message for receiver=%v MsgType=%v from %v: it does not belong to the peer domain.",
				recv.MsgReceiver, recv.MsgType, recv.SenderPubKey.AsString())
			return
		}
		callback(recv)
	})
	d.attachIDs = append(d.attachIDs, attachID)
	return attachID
}

func (d *DomainImpl) PeerStatus() []peering.PeerStatusProvider {
	res := make([]peering.PeerStatusProvider, 0)
	for _, v := range d.nodes {
		res = append(res, v.Status())
	}
	return res
}

func (d *DomainImpl) Detach(attachID interface{}) {
	d.netProvider.Detach(attachID)
}

func (d *DomainImpl) Close() {
	for _, attachID := range d.attachIDs {
		d.Detach(attachID)
	}
	for i := range d.nodes {
		d.nodes[i].Close()
	}
}
