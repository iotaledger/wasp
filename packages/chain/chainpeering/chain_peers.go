package chainpeering

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/peering"
)

type ChainPeersImpl struct {
	peeringID peering.PeeringID
	peers     peering.PeerDomainProvider
	attachIDs []interface{}
}

var _ chain.ChainPeers = &ChainPeersImpl{}

func NewChainPeers(peeringID peering.PeeringID, peers peering.PeerDomainProvider) chain.ChainPeers {
	return &ChainPeersImpl{
		peeringID: peeringID,
		peers:     peers,
		attachIDs: make([]interface{}, 0),
	}
}

func (cpiT *ChainPeersImpl) AttachToPeerMessages(receiver byte, fun func(recv *peering.PeerMessageIn)) {
	cpiT.attachIDs = append(cpiT.attachIDs, cpiT.peers.Attach(&cpiT.peeringID, receiver, fun))
}

func (cpiT *ChainPeersImpl) SendPeerMsgByNetID(netID string, msgReceiver, msgType byte, msgData []byte) {
	cpiT.peers.SendMsgByNetID(netID, &peering.PeerMessageData{
		PeeringID:   cpiT.peeringID,
		MsgReceiver: msgReceiver,
		MsgType:     msgType,
		MsgData:     msgData,
	})
}

func (cpiT *ChainPeersImpl) SendPeerMsgToRandomPeers(upToNumPeers int, msgReceiver, msgType byte, msgData []byte) {
	for _, netID := range cpiT.GetRandomPeers(upToNumPeers) {
		cpiT.SendPeerMsgByNetID(netID, msgReceiver, msgType, msgData)
	}
}

func (cpiT *ChainPeersImpl) GetRandomPeers(upToNumPeers int) []string {
	return cpiT.peers.GetRandomPeers(upToNumPeers)
}

func (cpiT *ChainPeersImpl) Close() {
	for _, attachID := range cpiT.attachIDs {
		cpiT.peers.Detach(attachID)
	}
}
