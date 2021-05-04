package mock_chain

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"testing"
	"time"
)

type MockedCommittee struct {
	T         *testing.T
	peerGroup peering.GroupProvider
	peeringID peering.PeeringID
	OwnIndex  uint16
	onPeerMsg func(recv *peering.RecvEvent)
}

func NewMockedCommittee(t *testing.T, peerGroup peering.GroupProvider, peeringID peering.PeeringID, index uint16) *MockedCommittee {
	ret := &MockedCommittee{
		T:         t,
		peerGroup: peerGroup,
		peeringID: peeringID,
		OwnIndex:  index,
	}
	return ret
}

func (cmt *MockedCommittee) Size() uint16 {
	return uint16(len(cmt.peerGroup.AllNodes()))
}

func (cmt *MockedCommittee) Quorum() uint16 {
	return (cmt.Size()*2)/3 + 1
}

func (cmt *MockedCommittee) OwnPeerIndex() uint16 {
	return cmt.OwnIndex
}

func (cmt *MockedCommittee) DKShare() *tcrypto.DKShare {
	panic("implement me")
}

func (cmt *MockedCommittee) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	cmt.peerGroup.SendMsgByIndex(targetPeerIndex, &peering.PeerMessage{
		PeeringID:   cmt.peeringID,
		SenderIndex: cmt.OwnPeerIndex(),
		Timestamp:   time.Now().UnixNano(),
		MsgType:     msgType,
		MsgData:     msgData,
	})
	return nil
}

func (cmt *MockedCommittee) SendMsgToPeers(msgType byte, msgData []byte, ts int64) uint16 {
	cmt.peerGroup.Broadcast(&peering.PeerMessage{
		PeeringID:   cmt.peeringID,
		SenderIndex: cmt.OwnPeerIndex(),
		Timestamp:   time.Now().UnixNano(),
		MsgType:     msgType,
		MsgData:     msgData,
	}, false)
	return cmt.Size() - 1
}

func (cmt *MockedCommittee) IsAlivePeer(peerIndex uint16) bool {
	peers := cmt.PeerStatus()
	if int(peerIndex) >= len(peers) || peers[peerIndex] == nil {
		return false
	}
	return peers[peerIndex].Connected
}

func (cmt *MockedCommittee) QuorumIsAlive(quorum ...uint16) bool {
	counter := uint16(0)
	for i := uint16(0); i < cmt.Size(); i++ {
		if cmt.IsAlivePeer(i) {
			counter++
		}
	}
	return counter >= cmt.Quorum()
}

func (cmt *MockedCommittee) PeerStatus() []*chain.PeerStatus {
	ret := make([]*chain.PeerStatus, cmt.Size())
	for idx, peer := range cmt.peerGroup.AllNodes() {
		ret[idx] = &chain.PeerStatus{
			Index:     int(idx),
			PeeringID: cmt.peeringID.String(),
			IsSelf:    idx == cmt.OwnIndex,
			Connected: peer.IsAlive(),
		}
	}
	return ret
}

func (cmt *MockedCommittee) OnPeerMessage(fun func(recv *peering.RecvEvent)) {
	cmt.onPeerMsg = fun
}

func (cmt *MockedCommittee) IsReady() bool {
	return true
}

func (cmt *MockedCommittee) Close() {
}
