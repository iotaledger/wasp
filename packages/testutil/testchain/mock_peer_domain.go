package testchain

import (
	"math/rand"
	"time"

	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
)

var _ peering.PeerDomainProvider = &MockedPeerDomainProvider{}

//---------------------------------------------
type MockedPeerDomainProvider struct {
	log              *logger.Logger
	ownNetID         string
	peerNetIDs       []string
	onSendMsgByNetID func(netID string, msg *peering.PeerMessage)
}

func NewMockedPeerDomain(ownNetID string, peerNetIDs []string, log *logger.Logger) *MockedPeerDomainProvider {
	log = log.Named("mockedDomain")
	if !util.AllDifferentStrings(peerNetIDs) {
		log.Panic("duplicate net IDs")
	}
	return &MockedPeerDomainProvider{
		log:              log,
		ownNetID:         ownNetID,
		peerNetIDs:       peerNetIDs,
		onSendMsgByNetID: func(netID string, msg *peering.PeerMessage) {},
	}
}

func (m *MockedPeerDomainProvider) SendMsgByNetID(netID string, msg *peering.PeerMessage) {
	if !util.StringInList(netID, m.peerNetIDs) {
		m.log.Warnf("%s is not among neighbors", netID)
		return
	}
	m.onSendMsgByNetID(netID, msg)
}

func (m *MockedPeerDomainProvider) SendMsgToRandomPeers(upToNumPeers uint16, msg *peering.PeerMessage) {
	netIDs := getRnd(m.peerNetIDs, upToNumPeers)
	for _, nid := range netIDs {
		m.SendMsgByNetID(nid, msg)
	}
}

func (m *MockedPeerDomainProvider) SendSimple(netID string, msgType byte, msgData []byte) {
	m.SendMsgByNetID(netID, &peering.PeerMessage{
		SenderNetID: m.ownNetID,
		Timestamp:   time.Now().UnixNano(),
		MsgType:     msgType,
		MsgData:     msgData,
	})
}

func (m *MockedPeerDomainProvider) SendMsgToRandomPeersSimple(upToNumPeers uint16, msgType byte, msgData []byte) {
	m.SendMsgToRandomPeers(upToNumPeers, &peering.PeerMessage{
		SenderNetID: m.ownNetID,
		Timestamp:   time.Now().UnixNano(),
		MsgType:     msgType,
		MsgData:     msgData,
	})
}

func (m *MockedPeerDomainProvider) ReshufflePeers(seedBytes ...[]byte) {
}

func (m *MockedPeerDomainProvider) Attach(peeringID *peering.PeeringID, callback func(recv *peering.RecvEvent)) interface{} {
	return nil
}

func (m *MockedPeerDomainProvider) Detach(attachID interface{}) {
}

func (m *MockedPeerDomainProvider) Close() {
}

func (m *MockedPeerDomainProvider) OnSend(fun func(netID string, msg *peering.PeerMessage)) {
	m.onSendMsgByNetID = fun
}

func getRnd(strs []string, n uint16) []string {
	if int(n) >= len(strs) {
		return strs
	}
	ret := make([]string, 0, n)
	for len(ret) < int(n) {
		i := rand.Intn(len(strs))
		if !util.StringInList(strs[i], ret) {
			ret = append(ret, strs[i])
		}
	}
	return ret
}
