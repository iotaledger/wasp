// Package group implements a generic peering.GroupProvider.
package group

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"errors"

	"github.com/iotaledger/wasp/packages/peering"
	"go.dedis.ch/kyber/v3"
)

//
// groupImpl implements peering.GroupProvider
//
type groupImpl struct {
	netProvider peering.NetworkProvider
	nodes       []peering.PeerSender
	other       map[uint16]peering.PeerSender
}

// NewPeeringGroupProvider creates a generic peering group.
// That should be used as a helper for peering implementations.
func NewPeeringGroupProvider(netProvider peering.NetworkProvider, nodes []peering.PeerSender) peering.GroupProvider {
	other := make(map[uint16]peering.PeerSender)
	for i := range nodes {
		if nodes[i].Location() != netProvider.Self().Location() {
			other[uint16(i)] = nodes[i]
		}
	}
	return &groupImpl{
		netProvider: netProvider,
		nodes:       nodes,
		other:       other,
	}
}

// PeerIndex implements peering.GroupProvider.
func (g *groupImpl) PeerIndex(peer peering.PeerSender) (uint16, error) {
	return g.PeerIndexByPub(peer.PubKey())
}

// PeerIndexByLoc implements peering.GroupProvider.
func (g *groupImpl) PeerIndexByLoc(peerLoc string) (uint16, error) {
	for i := range g.nodes {
		if g.nodes[i].Location() == peerLoc {
			return uint16(i), nil
		}
	}
	return uint16(0xFFFF), errors.New("peer_not_found_by_loc")
}

// PeerIndexByPub implements peering.GroupProvider.
func (g *groupImpl) PeerIndexByPub(peerPub kyber.Point) (uint16, error) {
	for i := range g.nodes {
		pubKey := g.nodes[i].PubKey()
		if pubKey != nil && pubKey.Equal(peerPub) {
			return uint16(i), nil
		}
	}
	return uint16(0xFFFF), errors.New("peer_not_found_by_pub")
}

// Broadcast implements peering.GroupProvider.
func (g *groupImpl) Broadcast(msg *peering.PeerMessage, includingSelf bool) {
	var peers map[uint16]peering.PeerSender
	if includingSelf {
		peers = g.AllNodes()
	} else {
		peers = g.OtherNodes()
	}
	for i := range peers {
		peers[i].SendMsg(msg)
	}
}

func (g *groupImpl) AllNodes() map[uint16]peering.PeerSender {
	all := make(map[uint16]peering.PeerSender)
	for i := range g.nodes {
		all[uint16(i)] = g.nodes[i]
	}
	return all
}

func (g *groupImpl) OtherNodes() map[uint16]peering.PeerSender {
	return g.other
}

// SendMsgByIndex implements peering.GroupProvider.
func (g *groupImpl) SendMsgByIndex(peerIdx uint16, msg *peering.PeerMessage) {
	g.nodes[peerIdx].SendMsg(msg)
}

// Close implements peering.GroupProvider.
func (g *groupImpl) Close() {
	for i := range g.nodes {
		g.nodes[i].Close()
	}
}
