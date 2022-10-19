// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package group implements a generic peering.GroupProvider.
package group

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
)

const NotInGroup uint16 = 0xFFFF

// groupImpl implements peering.GroupProvider
type groupImpl struct {
	netProvider peering.NetworkProvider
	nodes       []peering.PeerSender
	other       map[uint16]peering.PeerSender
	selfIndex   uint16
	peeringID   peering.PeeringID
	attachIDs   []interface{}
	log         *logger.Logger
}

var _ peering.GroupProvider = &groupImpl{}

// NewPeeringGroupProvider creates a generic peering group.
// That should be used as a helper for peering implementations.
func NewPeeringGroupProvider(netProvider peering.NetworkProvider, peeringID peering.PeeringID, nodes []peering.PeerSender, log *logger.Logger) (peering.GroupProvider, error) {
	other := make(map[uint16]peering.PeerSender)
	selfFound := false
	selfIndex := uint16(0)
	for i := range nodes {
		if nodes[i].NetID() == netProvider.Self().NetID() {
			selfIndex = uint16(i)
			selfFound = true
		} else {
			other[uint16(i)] = nodes[i]
		}
	}
	if !selfFound {
		return nil, xerrors.Errorf("group must involve the current node")
	}
	return &groupImpl{
		netProvider: netProvider,
		nodes:       nodes,
		other:       other,
		selfIndex:   selfIndex,
		peeringID:   peeringID,
		attachIDs:   make([]interface{}, 0),
		log:         log,
	}, nil
}

// PeerIndex implements peering.GroupProvider.
func (g *groupImpl) SelfIndex() uint16 {
	return g.selfIndex
}

// PeerIndex implements peering.GroupProvider.
func (g *groupImpl) PeerIndex(peer peering.PeerSender) (uint16, error) {
	return g.PeerIndexByPubKey(peer.PubKey())
}

// PeerIndexByNetID implements peering.GroupProvider.
func (g *groupImpl) PeerIndexByPubKey(peerPubKey *cryptolib.PublicKey) (uint16, error) {
	for i := range g.nodes {
		if g.nodes[i].PubKey().Equals(peerPubKey) {
			return uint16(i), nil
		}
	}
	return NotInGroup, errors.New("peer not found by pubKey")
}

func (g *groupImpl) PubKeyByIndex(index uint16) (*cryptolib.PublicKey, error) {
	if index < uint16(len(g.nodes)) {
		return g.nodes[index].PubKey(), nil
	}
	return nil, errors.New("peer index out of scope")
}

// SendMsgByIndex implements peering.GroupProvider.
func (g *groupImpl) SendMsgByIndex(peerIdx uint16, msgReceiver, msgType byte, msgData []byte) {
	g.nodes[peerIdx].SendMsg(&peering.PeerMessageData{
		PeeringID:   g.peeringID,
		MsgReceiver: msgReceiver,
		MsgType:     msgType,
		MsgData:     msgData,
	})
}

// Broadcast implements peering.GroupProvider.
func (g *groupImpl) SendMsgBroadcast(msgReceiver, msgType byte, msgData []byte, except ...uint16) {
	for i := range g.OtherNodes(except...) {
		g.SendMsgByIndex(i, msgReceiver, msgType, msgData)
	}
}

// ExchangeRound sends a message to the specified set of peers and waits for acks.
// Resends the messages if acks are not received for some time.
//

//nolint: gocyclo
func (g *groupImpl) ExchangeRound(
	peers map[uint16]peering.PeerSender,
	recvCh chan *peering.PeerMessageIn,
	retryTimeout time.Duration,
	giveUpTimeout time.Duration,
	sendCB func(peerIdx uint16, peer peering.PeerSender),
	recvCB func(recv *peering.PeerMessageGroupIn) (bool, error),
) error {
	acks := make(map[uint16]bool)
	errs := make(map[uint16]error)
	retryCh := time.After(retryTimeout)
	giveUpCh := time.After(giveUpTimeout)
	for i := range peers {
		acks[i] = false
		sendCB(i, peers[i])
	}
	haveAllAcks := func() bool {
		for i := range acks {
			if !acks[i] {
				return false
			}
		}
		return true
	}
	for !haveAllAcks() {
		select {
		case recvMsgNoIndex, ok := <-recvCh:
			if !ok {
				return errors.New("recv_channel_closed")
			}
			senderIndex, err := g.PeerIndexByPubKey(recvMsgNoIndex.SenderPubKey)
			if err != nil {
				g.log.Warnf(
					"Dropping message %v -> %v, MsgType=%v because of %v",
					recvMsgNoIndex.SenderPubKey.String(), g.netProvider.Self().PubKey().String(),
					recvMsgNoIndex.MsgType, err,
				)
				continue
			}
			recvMsg := peering.PeerMessageGroupIn{
				PeerMessageIn: *recvMsgNoIndex,
				SenderIndex:   senderIndex,
			}
			if acks[recvMsg.SenderIndex] { // Only consider first successful message.
				g.log.Warnf(
					"Dropping duplicate message %v -> %v, receiver=%v, MsgType=%v",
					recvMsg.SenderPubKey.String(), g.netProvider.Self().PubKey().String(),
					recvMsg.MsgReceiver, recvMsg.MsgType,
				)
				continue
			}
			if acks[recvMsg.SenderIndex], err = recvCB(&recvMsg); err != nil {
				errs[recvMsg.SenderIndex] = err
				continue
			}
			if acks[recvMsg.SenderIndex] {
				// Clear previous errors on success.
				delete(errs, recvMsg.SenderIndex)
			}
		case <-retryCh:
			for i := range peers {
				if !acks[i] {
					sendCB(i, peers[i])
				}
			}
			retryCh = time.After(retryTimeout)
		case <-giveUpCh:
			var errMsg string
			for i := range peers {
				if acks[i] {
					continue
				}
				if errs[i] != nil {
					errMsg += fmt.Sprintf("[%v:%v]", i, errs[i].Error())
				} else {
					errMsg += fmt.Sprintf("[%v:%v]", i, "round_timeout")
				}
			}
			return errors.New(errMsg)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	var errMsg string
	for i := range errs {
		errMsg += fmt.Sprintf("[%v:%v]", i, errs[i].Error())
	}
	return errors.New(errMsg)
}

// AllNodes returns a map of all nodes in the group.
func (g *groupImpl) AllNodes(except ...uint16) map[uint16]peering.PeerSender {
	all := make(map[uint16]peering.PeerSender)
	exceptions := make(map[uint16]struct{})
	for _, i := range except {
		exceptions[i] = struct{}{}
	}
	for i := range g.nodes {
		if _, ok := exceptions[uint16(i)]; !ok {
			all[uint16(i)] = g.nodes[i]
		}
	}
	return all
}

// OtherNodes returns a map of all nodes in the group, excluding the self node.
func (g *groupImpl) OtherNodes(except ...uint16) map[uint16]peering.PeerSender {
	if len(except) == 0 {
		return g.other
	}
	ret := make(map[uint16]peering.PeerSender)
	exceptions := make(map[uint16]struct{})
	for _, i := range except {
		exceptions[i] = struct{}{}
	}
	for i := range g.other {
		if _, ok := exceptions[i]; !ok {
			ret[i] = g.other[i]
		}
	}
	return ret
}

// Attach starts listening for messages. Messages in this case will be filtered
// to those received from nodes in the group only. SenderIndex will be filled
// for the messages according to the message source.
func (g *groupImpl) Attach(receiver byte, callback func(recv *peering.PeerMessageGroupIn)) interface{} {
	attachID := g.netProvider.Attach(&g.peeringID, receiver, func(recv *peering.PeerMessageIn) {
		idx, err := g.PeerIndexByPubKey(recv.SenderPubKey)
		if idx == NotInGroup {
			err = xerrors.Errorf("sender does not belong to the group")
		}
		if err != nil {
			g.log.Warnf("dropping message for receiver=%v MsgType=%v from %v: %v.",
				recv.MsgReceiver, recv.MsgType, recv.SenderPubKey.String(), err)
			return
		}
		gRecv := &peering.PeerMessageGroupIn{
			PeerMessageIn: *recv,
			SenderIndex:   idx,
		}
		callback(gRecv)
	})
	g.attachIDs = append(g.attachIDs, attachID)
	return attachID
}

// Detach terminates listening for messages.
func (g *groupImpl) Detach(attachID interface{}) {
	g.netProvider.Detach(attachID)
}

// Close implements peering.GroupProvider.
func (g *groupImpl) Close() {
	for _, attachID := range g.attachIDs {
		g.Detach(attachID)
	}
	for i := range g.nodes {
		g.nodes[i].Close()
	}
}
