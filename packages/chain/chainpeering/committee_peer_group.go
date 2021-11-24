// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainpeering

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/peering"
)

type committeePeerGroup struct {
	peeringID peering.PeeringID
	group     peering.GroupProvider
	attachIDs []interface{}
}

var _ chain.CommitteePeerGroup = &committeePeerGroup{}

func NewCommitteePeerGroup(peeringID peering.PeeringID, group peering.GroupProvider) chain.CommitteePeerGroup {
	return &committeePeerGroup{
		peeringID: peeringID,
		group:     group,
		attachIDs: make([]interface{}, 0),
	}
}

func (cpgT *committeePeerGroup) SendMsgByIndex(peerIdx uint16, msgReceiver, msgType byte, msgData []byte) error {
	if peer, ok := cpgT.group.OtherNodes()[peerIdx]; ok {
		peer.SendMsg(&peering.PeerMessageData{
			PeeringID:   cpgT.peeringID,
			MsgReceiver: msgReceiver,
			MsgType:     msgType,
			MsgData:     msgData,
		})
		return nil
	}
	return fmt.Errorf("SendMsg: wrong peer index")
}

func (cpgT *committeePeerGroup) SendMsgBroadcast(msgReceiver, msgType byte, msgData []byte, except ...uint16) {
	msg := &peering.PeerMessageData{
		PeeringID:   cpgT.peeringID,
		MsgReceiver: msgReceiver,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	cpgT.group.SendMsgBroadcast(msg, except...)
}

func (cpgT *committeePeerGroup) AttachToPeerMessages(peerMsgReceiver byte, fun func(peerMsg *peering.PeerMessageGroupIn)) {
	cpgT.attachIDs = append(cpgT.attachIDs, cpgT.group.Attach(&cpgT.peeringID, peerMsgReceiver, fun))
}

func (cpgT *committeePeerGroup) SelfIndex() uint16 {
	return cpgT.group.SelfIndex()
}

func (cpgT *committeePeerGroup) AllNodes(except ...uint16) map[uint16]peering.PeerSender {
	return cpgT.group.AllNodes(except...)
}

func (cpgT *committeePeerGroup) OtherNodes(except ...uint16) map[uint16]peering.PeerSender {
	return cpgT.group.AllNodes(except...)
}

func (cpgT *committeePeerGroup) Close() {
	for _, attachID := range cpgT.attachIDs {
		cpgT.group.Detach(attachID)
	}
	cpgT.group.Close()
}
