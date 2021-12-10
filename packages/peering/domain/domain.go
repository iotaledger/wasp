package domain

import (
	"crypto/rand"
	"sort"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
)

type DomainImpl struct {
	netProvider peering.NetworkProvider
	nodes       map[string]peering.PeerSender
	permutation *util.Permutation16
	netIDs      []string
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
		nodes:       make(map[string]peering.PeerSender),
		permutation: util.NewPermutation16(uint16(len(initialNodes)), nil),
		netIDs:      make([]string, 0, len(initialNodes)),
		peeringID:   peeringID,
		attachIDs:   make([]interface{}, 0),
		log:         log,
		mutex:       &sync.RWMutex{},
	}
	for _, sender := range initialNodes {
		ret.nodes[sender.NetID()] = sender
	}
	ret.reshufflePeers()
	return ret
}

func NewPeerDomainByNetIDs(netProvider peering.NetworkProvider, peeringID peering.PeeringID, peerNetIDs []string, log *logger.Logger) (*DomainImpl, error) {
	peers := make([]peering.PeerSender, 0, len(peerNetIDs))
	for _, nid := range peerNetIDs {
		if nid == netProvider.Self().NetID() {
			continue
		}
		peer, err := netProvider.PeerByNetID(nid)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}
	return NewPeerDomain(netProvider, peeringID, peers, log), nil
}

func (d *DomainImpl) SendMsgByNetID(netID string, msgReceiver, msgType byte, msgData []byte) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	peer, ok := d.nodes[netID]
	if !ok {
		d.log.Warnf("SendMsgByNetID: NetID %v is not in the domain", netID)
		return
	}
	peer.SendMsg(&peering.PeerMessageData{
		PeeringID:   d.peeringID,
		MsgReceiver: msgReceiver,
		MsgType:     msgType,
		MsgData:     msgData,
	})
}

func (d *DomainImpl) SendPeerMsgToRandomPeers(upToNumPeers int, msgReceiver, msgType byte, msgData []byte) {
	for _, netID := range d.GetRandomPeers(upToNumPeers) {
		d.SendMsgByNetID(netID, msgReceiver, msgType, msgData)
	}
}

func (d *DomainImpl) GetRandomPeers(upToNumPeers int) []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if upToNumPeers > len(d.netIDs) {
		upToNumPeers = len(d.netIDs)
	}
	ret := make([]string, upToNumPeers)
	for i := range ret {
		ret[i] = d.netIDs[d.permutation.Next()]
	}
	return ret
}

func (d *DomainImpl) AddPeer(netID string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.nodes[netID]; ok {
		return nil
	}
	if netID == d.netProvider.Self().NetID() {
		return nil
	}
	peer, err := d.netProvider.PeerByNetID(netID)
	if err != nil {
		return err
	}
	d.nodes[netID] = peer
	d.reshufflePeers()

	return nil
}

func (d *DomainImpl) RemovePeer(netID string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.nodes, netID)
	d.reshufflePeers()
}

func (d *DomainImpl) ReshufflePeers(seedBytes ...[]byte) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.reshufflePeers(seedBytes...)
}

func (d *DomainImpl) reshufflePeers(seedBytes ...[]byte) {
	d.netIDs = make([]string, 0, len(d.nodes))
	for netID := range d.nodes {
		d.netIDs = append(d.netIDs, netID)
	}
	sort.Strings(d.netIDs)
	var seedB []byte
	if len(seedBytes) == 0 {
		var b [8]byte
		seedB = b[:]
		_, _ = rand.Read(seedB)
	} else {
		seedB = seedBytes[0]
	}
	d.permutation.Shuffle(seedB)
}

func (d *DomainImpl) Attach(receiver byte, callback func(recv *peering.PeerMessageIn)) interface{} {
	attachID := d.netProvider.Attach(&d.peeringID, receiver, func(recv *peering.PeerMessageIn) {
		if recv.SenderNetID == d.netProvider.Self().NetID() {
			d.log.Warnf("dropping message for receiver=%v MsgType=%v from %v: message from self.",
				recv.MsgReceiver, recv.MsgType, recv.SenderNetID)
			return
		}
		_, ok := d.nodes[recv.SenderNetID]
		if !ok {
			d.log.Warnf("dropping message for receiver=%v MsgType=%v from %v: it does not belong to the peer domain.",
				recv.MsgReceiver, recv.MsgType, recv.SenderNetID)
			return
		}
		callback(recv)
	})
	d.attachIDs = append(d.attachIDs, attachID)
	return attachID
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
