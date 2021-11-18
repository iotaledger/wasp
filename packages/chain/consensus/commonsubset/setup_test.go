// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package commonsubset

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type committeeMock struct {
	peeringID peering.PeeringID
	group     peering.GroupProvider
	t         *testing.T
}

var _ chain.Committee = &committeeMock{}

func NewCommitteeMock(peeringID peering.PeeringID, group peering.GroupProvider, t *testing.T) chain.Committee {
	return &committeeMock{
		peeringID: peeringID,
		group:     group,
		t:         t,
	}
}

func (cmT *committeeMock) Address() ledgerstate.Address {
	panic("Not implemented")
}

func (cmT *committeeMock) Size() uint16 {
	panic("Not implemented")
}

func (cmT *committeeMock) Quorum() uint16 {
	panic("Not implemented")
}

func (cmT *committeeMock) OwnPeerIndex() uint16 {
	panic("Not implemented")
}

func (cmT *committeeMock) DKShare() *tcrypto.DKShare {
	panic("Not implemented")
}

func (cmT *committeeMock) SendMsgByIndex(peerIdx uint16, msgReceiver, msgType byte, msgData []byte) error {
	if peer, ok := cmT.group.OtherNodes()[peerIdx]; ok {
		peer.SendMsg(&peering.PeerMessageNet{
			PeerMessageData: peering.PeerMessageData{
				PeeringID:   cmT.peeringID,
				Timestamp:   time.Now().UnixNano(),
				MsgReceiver: msgReceiver,
				MsgType:     msgType,
				MsgData:     msgData,
			},
		})
		return nil
	}
	return fmt.Errorf("SendMsg: wrong peer index")
}

func (cmT *committeeMock) SendMsgBroadcast(msgReceiver, msgType byte, msgData []byte, except ...uint16) {
	panic("Not implemented")
}

func (cmT *committeeMock) IsAlivePeer(peerIndex uint16) bool {
	panic("Not implemented")
}

func (cmT *committeeMock) QuorumIsAlive(quorum ...uint16) bool {
	panic("Not implemented")
}

func (cmT *committeeMock) PeerStatus() []*chain.PeerStatus {
	panic("Not implemented")
}

func (cmT *committeeMock) AttachToPeerMessages(peerMsgReceiver byte, fun func(peerMsg *peering.PeerMessageGroupIn)) {
	cmT.group.Attach(&cmT.peeringID, peerMsgReceiver, fun)
}

func (cmT *committeeMock) IsReady() bool {
	panic("Not implemented")
}

func (cmT *committeeMock) Close() {
	panic("Not implemented")
}

func (cmT *committeeMock) RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte)) {
	panic("Not implemented")
}

func (cmT *committeeMock) GetOtherValidatorsPeerIDs() []string {
	panic("Not implemented")
}

func (cmT *committeeMock) GetRandomValidators(upToN int) []string {
	panic("Not implemented")
}
