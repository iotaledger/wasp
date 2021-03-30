package chainimpl

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type committeeObj struct {
	chainObj    chain.Chain
	netProvider peering.NetworkProvider
	peers       peering.GroupProvider
	size        uint16
	quorum      uint16
	ownIndex    uint16
	dkshare     *tcrypto.DKShare
	attachID    interface{}
}

func NewCommittee(chainObj chain.Chain, stateAddr ledgerstate.Address, netProvider peering.NetworkProvider, dksProvider tcrypto.RegistryProvider) (chain.Committee, error) {
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
	if util.ContainsDuplicates(cmtRec.Nodes) {
		return nil, xerrors.Errorf(
			"NewCommittee: committee record for %s contains duplicate node addresses: %+v",
			stateAddr, cmtRec.Nodes)
	}
	if dkshare.Index == nil || !iAmInTheCommittee(cmtRec.Nodes, dkshare.N, *dkshare.Index, netProvider) {
		return nil, xerrors.Errorf(
			"NewCommittee: chain record inconsistency. the own node %s is not in the committee for %s: %+v",
			netProvider.Self().NetID(), cmtRec.Address, cmtRec.Nodes,
		)
	}
	var peers peering.GroupProvider
	if peers, err = netProvider.Group(cmtRec.Nodes); err != nil {
		return nil, xerrors.Errorf(
			"node %s failed to setup committee communication with %+v, reason=%+v",
			netProvider.Self().NetID(), cmtRec.Nodes, err,
		)
	}
	return &committeeObj{
		chainObj:    chainObj,
		peers:       peers,
		netProvider: netProvider,
		size:        dkshare.N,
		quorum:      dkshare.T,
		ownIndex:    *dkshare.Index,
		dkshare:     dkshare,
	}, nil
}

func (c *committeeObj) Size() uint16 {
	return c.size
}

func (c *committeeObj) Quorum() uint16 {
	return c.quorum
}

func (c *committeeObj) OwnPeerIndex() uint16 {
	return c.ownIndex
}

func (c *committeeObj) DKShare() *tcrypto.DKShare {
	return c.dkshare
}

func (c *committeeObj) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if peer, ok := c.peers.OtherNodes()[targetPeerIndex]; ok {
		chainID := *c.chainObj.ID()
		peer.SendMsg(&peering.PeerMessage{
			ChainID:     chainID,
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
		ChainID:     *c.chainObj.ID(),
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

func (c *committeeObj) QuorumIsAlive() bool {
	count := uint16(0)
	for _, peer := range c.peers.OtherNodes() {
		if peer.IsAlive() {
			count++
		}
		if count+1 >= c.quorum {
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
	c.attachID = c.peers.Attach(c.chainObj.ID(), fun)
}

func (c *committeeObj) Close() {
	c.peers.Detach(c.attachID)
	c.peers.Close()
}

func (c *committeeObj) FeeDestination() coretypes.AgentID {
	return *coretypes.NewAgentID(c.chainObj.ID().AsAddress(), 0)
}

func (c *committeeObj) Chain() chain.Chain {
	return c.chainObj
}
