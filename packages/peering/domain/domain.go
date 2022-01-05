// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"crypto/rand"
	"sync"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
)

type DomainImpl struct {
	netProvider peering.NetworkProvider
	nodes       map[ed25519.PublicKey]peering.PeerSender
	permutation *util.Permutation16
	permPubKeys []*ed25519.PublicKey
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
		nodes:       make(map[ed25519.PublicKey]peering.PeerSender),
		permutation: nil, // Will be set in ret.reshufflePeers().
		permPubKeys: nil, // Will be set in ret.reshufflePeers().
		peeringID:   peeringID,
		attachIDs:   make([]interface{}, 0),
		log:         log,
		mutex:       &sync.RWMutex{},
	}
	for _, sender := range initialNodes {
		ret.nodes[*sender.PubKey()] = sender
	}
	ret.reshufflePeers()
	return ret
}

func (d *DomainImpl) SendMsgByPubKey(pubKey *ed25519.PublicKey, msgReceiver, msgType byte, msgData []byte) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	peer, ok := d.nodes[*pubKey]
	if !ok {
		d.log.Warnf("SendMsgByPubKey: PubKey %v is not in the domain", pubKey.String())
		return
	}
	peer.SendMsg(&peering.PeerMessageData{
		PeeringID:   d.peeringID,
		MsgReceiver: msgReceiver,
		MsgType:     msgType,
		MsgData:     msgData,
	})
}

func (d *DomainImpl) GetRandomOtherPeers(upToNumPeers int) []*ed25519.PublicKey {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if upToNumPeers > len(d.permPubKeys) {
		upToNumPeers = len(d.permPubKeys)
	}
	ret := make([]*ed25519.PublicKey, upToNumPeers)
	for i := range ret {
		ret[i] = d.permPubKeys[d.permutation.Next()]
	}
	return ret
}

func (d *DomainImpl) UpdatePeers(newPeerPubKeys []*ed25519.PublicKey) {
	d.mutex.RLock()
	nodes := make(map[ed25519.PublicKey]peering.PeerSender)
	for k, v := range d.nodes {
		nodes[k] = v
	}
	d.mutex.RUnlock()
	changed := false
	//
	// Add new peers.
	for _, newPeerPubKey := range newPeerPubKeys {
		found := false
		for _, existingPeer := range nodes {
			if *existingPeer.PubKey() == *newPeerPubKey {
				found = true
				break
			}
		}
		if !found {
			newPeerSender, err := d.netProvider.PeerByPubKey(newPeerPubKey)
			if err != nil {
				// TODO: Maybe more control should be needed here. Distinguish the mandatory and the optional nodes?
				d.log.Warnf("Peer with pubKey=%v not found, will be ignored for now, reason: %v", newPeerPubKey.String(), err)
			} else {
				changed = true
				nodes[*newPeerSender.PubKey()] = newPeerSender
				d.log.Infof("Domain peer added, pubKey=%v, netID=%v", newPeerSender.PubKey().String(), newPeerSender.NetID())
			}
		}
	}
	//
	// Remove peers that are not needed anymore.
	for _, oldPeer := range nodes {
		found := false
		for _, newPeerPubKey := range newPeerPubKeys {
			if *oldPeer.PubKey() == *newPeerPubKey {
				found = true
				break
			}
		}
		if !found && (*oldPeer.PubKey() != *d.netProvider.Self().PubKey()) {
			changed = true
			delete(nodes, *oldPeer.PubKey())
			d.log.Infof("Domain peer removed, pubKey=%v, netID=%v", oldPeer.PubKey().String(), oldPeer.NetID())
		}
	}
	if changed {
		d.mutex.Lock()
		d.nodes = nodes
		d.reshufflePeers()
		d.mutex.Unlock()
	}
}

func (d *DomainImpl) ReshufflePeers(seedBytes ...[]byte) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.reshufflePeers(seedBytes...)
}

func (d *DomainImpl) reshufflePeers(seedBytes ...[]byte) {
	d.permPubKeys = make([]*ed25519.PublicKey, 0, len(d.nodes))
	for pubKey := range d.nodes {
		peerPubKey := pubKey
		if peerPubKey != *d.netProvider.Self().PubKey() { // Do not include self to the permutation.
			d.permPubKeys = append(d.permPubKeys, &peerPubKey)
		}
	}
	var seedB []byte
	if len(seedBytes) == 0 {
		var b [8]byte
		seedB = b[:]
		_, _ = rand.Read(seedB)
	} else {
		seedB = seedBytes[0]
	}
	d.permutation = util.NewPermutation16(uint16(len(d.permPubKeys)), seedB)
}

func (d *DomainImpl) Attach(receiver byte, callback func(recv *peering.PeerMessageIn)) interface{} {
	attachID := d.netProvider.Attach(&d.peeringID, receiver, func(recv *peering.PeerMessageIn) {
		if *recv.SenderPubKey == *d.netProvider.Self().PubKey() {
			d.log.Warnf("dropping message for receiver=%v MsgType=%v from %v: message from self.",
				recv.MsgReceiver, recv.MsgType, recv.SenderPubKey.String())
			return
		}
		_, ok := d.nodes[*recv.SenderPubKey]
		if !ok {
			d.log.Warnf("dropping message for receiver=%v MsgType=%v from %v: it does not belong to the peer domain.",
				recv.MsgReceiver, recv.MsgType, recv.SenderPubKey.String())
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
