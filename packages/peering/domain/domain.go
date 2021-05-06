package domain

import (
	"crypto/rand"
	"sort"
	"sync"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
)

type domainImpl struct {
	netProvider peering.NetworkProvider
	nodes       map[string]peering.PeerSender
	permutation *util.Permutation16
	netIDs      []string
	log         *logger.Logger
	mutex       *sync.RWMutex
}

// NewPeeringDomain creates a collection. Ignores self
func NewPeeringDomain(netProvider peering.NetworkProvider, initialNodes []peering.PeerSender, log *logger.Logger) *domainImpl {
	ret := &domainImpl{
		netProvider: netProvider,
		nodes:       make(map[string]peering.PeerSender),
		permutation: util.NewPermutation16(uint16(len(initialNodes)), nil),
		netIDs:      make([]string, 0, len(initialNodes)),
		log:         log,
		mutex:       &sync.RWMutex{},
	}
	for _, sender := range initialNodes {
		ret.nodes[sender.NetID()] = sender
	}
	ret.reshufflePeers()
	return ret
}

func (d *domainImpl) SendMsgByNetID(netID string, msg *peering.PeerMessage) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	peer, ok := d.nodes[netID]
	if !ok {
		d.log.Warnf("SendMsgByNetID: wrong netID %s", netID)
		return
	}
	peer.SendMsg(msg)
}

func (d *domainImpl) SendMsgToRandomPeers(upToNumPeers uint16, msg *peering.PeerMessage) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if int(upToNumPeers) > len(d.nodes) {
		upToNumPeers = uint16(len(d.nodes))
	}
	for i := uint16(0); i < upToNumPeers; i++ {
		d.SendMsgByNetID(d.netIDs[d.permutation.Next()], msg)
	}
}

func (d *domainImpl) AddPeer(netID string) error {
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

func (d *domainImpl) RemovePeer(netID string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.nodes, netID)
	d.reshufflePeers()
}

func (d *domainImpl) ReshufflePeers(seedBytes ...[]byte) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.reshufflePeers(seedBytes...)
}

func (d *domainImpl) reshufflePeers(seedBytes ...[]byte) {
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

func (d *domainImpl) Attach(peeringID *peering.PeeringID, callback func(recv *peering.RecvEvent)) interface{} {
	return d.netProvider.Attach(peeringID, func(recv *peering.RecvEvent) {
		peer, ok := d.nodes[recv.From.NetID()]
		if ok && peer.NetID() != d.netProvider.Self().NetID() {
			recv.Msg.SenderNetID = peer.NetID()
			callback(recv)
			return
		}
		d.log.Warnf("dropping message MsgType=%v from %v, it does not belong to the peer domain.",
			recv.Msg.MsgType, recv.From.NetID())
	})
}

func (d *domainImpl) Detach(attachID interface{}) {
	d.netProvider.Detach(attachID)
}

func (d *domainImpl) Close() {
	for i := range d.nodes {
		d.nodes[i].Close()
	}
}
