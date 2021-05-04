package committeeimpl

import (
	"fmt"
	"github.com/iotaledger/hive.go/logger"
	"go.uber.org/atomic"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"golang.org/x/xerrors"
)

type committeeObj struct {
	isReady     *atomic.Bool
	netProvider peering.NetworkProvider
	peers       peering.GroupProvider
	peeringID   peering.PeeringID
	size        uint16
	quorum      uint16
	ownIndex    uint16
	dkshare     *tcrypto.DKShare
	attachID    interface{}
	log         *logger.Logger
}

func NewCommittee(stateAddr ledgerstate.Address, netProvider peering.NetworkProvider, dksProvider tcrypto.RegistryProvider, log *logger.Logger) (chain.Committee, error) {
	cmtRec, err := registry.CommitteeRecordFromRegistry(stateAddr)
	if err != nil || cmtRec == nil {
		return nil, xerrors.Errorf(
			"NewCommittee: failed to lead committee record for address %s. err = %v", stateAddr, err)
	}
	dkshare, err := dksProvider.LoadDKShare(cmtRec.Address)
	if err != nil {
		return nil, xerrors.Errorf(
			"NewCommittee: failed loading DKShare for address %s: %v", stateAddr, err)
	}
	if containsDuplicates(cmtRec.Nodes) {
		return nil, xerrors.Errorf(
			"NewCommittee: committee record for %s contains duplicate node addresses: %+v",
			stateAddr, cmtRec.Nodes)
	}
	if dkshare.Index == nil || !iAmInTheCommittee(cmtRec.Nodes, dkshare.N, *dkshare.Index, netProvider) {
		return nil, xerrors.Errorf(
			"NewCommittee: chain record inconsistency. the own node %s is not in the committee for %s: %+v",
			netProvider.Self().NetID(), cmtRec.Address.Base58(), cmtRec.Nodes,
		)
	}
	var peers peering.GroupProvider
	if peers, err = netProvider.Group(cmtRec.Nodes); err != nil {
		return nil, xerrors.Errorf(
			"node %s failed to setup committee communication with %+v, reason=%+v",
			netProvider.Self().NetID(), cmtRec.Nodes, err,
		)
	}

	ret := &committeeObj{
		isReady:     atomic.NewBool(false),
		peers:       peers,
		peeringID:   stateAddr.Array(), // committee is made for specific state address
		netProvider: netProvider,
		size:        dkshare.N,
		quorum:      dkshare.T,
		ownIndex:    *dkshare.Index,
		dkshare:     dkshare,
		log:         log,
	}
	go ret.waitReady()

	return ret, nil
}

func (c *committeeObj) Size() uint16 {
	return c.size
}

func (c *committeeObj) Quorum() uint16 {
	return c.quorum
}

func (c *committeeObj) IsReady() bool {
	return c.isReady.Load()
}

func (c *committeeObj) OwnPeerIndex() uint16 {
	return c.ownIndex
}

func (c *committeeObj) DKShare() *tcrypto.DKShare {
	return c.dkshare
}

func (c *committeeObj) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if peer, ok := c.peers.OtherNodes()[targetPeerIndex]; ok {
		peer.SendMsg(&peering.PeerMessage{
			PeeringID:   c.peeringID,
			SenderIndex: c.ownIndex,
			MsgType:     msgType,
			MsgData:     msgData,
		})
		return nil
	}
	return fmt.Errorf("SendMsg: wrong peer index")
}

func (c *committeeObj) SendMsgToPeers(msgType byte, msgData []byte, ts int64) uint16 {
	msg := &peering.PeerMessage{
		PeeringID:   c.peeringID,
		SenderIndex: c.ownIndex,
		Timestamp:   ts,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	c.peers.Broadcast(msg, false)
	return uint16(len(c.peers.OtherNodes())) // TODO: [KP] Reconsider this, we cannot guaranty if they are actually sent.
}

func (c *committeeObj) IsAlivePeer(peerIndex uint16) bool {
	allNodes := c.peers.AllNodes()
	if int(peerIndex) >= len(allNodes) {
		return false
	}
	if peerIndex == c.ownIndex {
		return true
	}
	if allNodes[peerIndex] == nil {
		panic(xerrors.Errorf("c.peers[peerIndex] == nil. peerIndex: %d, ownIndex: %d", peerIndex, c.ownIndex))
	}
	return allNodes[peerIndex].IsAlive()
}

func (c *committeeObj) QuorumIsAlive(quorum ...uint16) bool {
	q := c.quorum
	if len(quorum) > 0 {
		q = quorum[0]
	}
	count := uint16(0)
	for _, peer := range c.peers.OtherNodes() {
		if peer.IsAlive() {
			count++
		}
		if count+1 >= q {
			return true
		}
	}
	return false
}

func (c *committeeObj) PeerStatus() []*chain.PeerStatus {
	ret := make([]*chain.PeerStatus, 0)
	for i, peer := range c.peers.AllNodes() {
		status := &chain.PeerStatus{
			Index:  int(i),
			IsSelf: peer == nil || peer.NetID() == c.netProvider.Self().NetID(),
		}
		if status.IsSelf {
			status.PeeringID = c.netProvider.Self().NetID()
			status.Connected = true
		} else {
			status.PeeringID = peer.NetID()
			status.Connected = peer.IsAlive()
		}
		ret = append(ret, status)
	}
	return ret
}

func (c *committeeObj) OnPeerMessage(fun func(recv *peering.RecvEvent)) {
	c.attachID = c.peers.Attach(&c.peeringID, fun)
}

func (c *committeeObj) Close() {
	c.isReady.Store(false)
	c.peers.Detach(c.attachID)
	c.peers.Close()
}

func (c *committeeObj) waitReady() {
	c.log.Infof("wait for at least quorum of comittee peers (%d) to connect before activating the committee", c.Quorum())
	for !c.QuorumIsAlive() {
		time.Sleep(500 * time.Millisecond)
	}
	c.log.Infof("committee is ready for addr %s", c.dkshare.Address.Base58())
	c.log.Debugf("peer status: %s", c.PeerStatus())
	c.isReady.Store(true)
}

func containsDuplicates(lst []string) bool {
	for i := range lst {
		for j := i + 1; j < len(lst); j++ {
			if lst[i] == lst[j] {
				return true
			}
		}
	}
	return false
}

// iAmInTheCommittee checks if NetIDs makes sense
func iAmInTheCommittee(committeeNodes []string, n, index uint16, netProvider peering.NetworkProvider) bool {
	if len(committeeNodes) != int(n) {
		return false
	}
	return committeeNodes[index] == netProvider.Self().NetID()
}
