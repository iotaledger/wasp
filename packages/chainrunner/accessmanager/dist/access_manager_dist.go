// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package dist implements chain access management following the specification `WaspChainAccessNodesV4.tla`.
// The specification actions are mapped to GPA inputs here as follows:
//
//   - ChainActivate  -- first reception of inputAccessNodes.
//   - ChainDeactivate -- inputChainDisabled.
//   - AccessNodeAdd -- inputAccessNodes, then compare with info we had before.
//   - AccessNodeDel -- inputAccessNodes, then compare with info we had before.
//   - Reboot -- inputTrustedNodes.
package dist

import (
	"fmt"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type AccessMgr interface {
	AsGPA() gpa.GPA
}

type Output interface {
	ChainServerNodes() []*cryptolib.PublicKey
}

type accessMgrDist struct {
	nodes            *shrinkingmap.ShrinkingMap[gpa.NodeID, *accessMgrNode] // State for each peer.
	chain            *accessMgrChain                                        // State of chain.
	pubKeyToNodeID   func(*cryptolib.PublicKey) gpa.NodeID                  // Convert PubKeys to NodeIDs.
	serversUpdatedCB func([]*cryptolib.PublicKey)                           // Called when a set of servers has changed for a chain.
	dismissPeerCB    func(*cryptolib.PublicKey)                             // To stop redelivery at the upper layer.
	log              log.Logger
}

var _ gpa.GPA = &accessMgrDist{}

func NewAccessMgr(
	pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID,
	serversUpdatedCB func(servers []*cryptolib.PublicKey),
	dismissPeerCB func(*cryptolib.PublicKey),
	log log.Logger,
) AccessMgr {
	return &accessMgrDist{
		nodes:            shrinkingmap.New[gpa.NodeID, *accessMgrNode](),
		pubKeyToNodeID:   pubKeyToNodeID,
		serversUpdatedCB: serversUpdatedCB,
		dismissPeerCB:    dismissPeerCB,
		log:              log,
	}
}

// Implements the AccessMgr interface.
func (amd *accessMgrDist) AsGPA() gpa.GPA {
	return amd
}

// Implements the Output interface.
func (amd *accessMgrDist) ChainServerNodes() []*cryptolib.PublicKey {
	if amd.chain != nil {
		return amd.chain.server.Values()
	}
	return []*cryptolib.PublicKey{}
}

// Implements the gpa.GPA interface.
func (amd *accessMgrDist) Input(input gpa.Input) gpa.OutMessages {
	switch input := input.(type) {
	case *inputChainDisabled:
		return amd.handleInputChainDisabled(input)
	case *inputAccessNodes:
		return amd.handleInputAccessNodes(input)
	case *inputTrustedNodes:
		return amd.handleInputTrustedNodes(input)
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (amd *accessMgrDist) Message(msg gpa.Message) gpa.OutMessages {
	if msg, ok := msg.(*msgAccess); ok {
		return amd.handleMsgAccess(msg)
	}
	panic(fmt.Errorf("unexpected message %T: %+v", msg, msg))
}

// Implements the gpa.GPA interface.
func (amd *accessMgrDist) Output() gpa.Output {
	return amd
}

// Implements the gpa.GPA interface.
func (amd *accessMgrDist) StatusString() string {
	return fmt.Sprintf("{accessMgr, |nodes|=%v, |hasChain|=%v}", amd.nodes.Size(), amd.chain != nil)
}

// > Notify all the trusted access nodes, that we will not serve the requests anymore.
func (amd *accessMgrDist) handleInputChainDisabled(_ *inputChainDisabled) gpa.OutMessages {
	if amd.chain == nil {
		return nil // Already disabled.
	}
	amd.chain.Disabled()
	amd.chain = nil
	msgs := gpa.NoMessages()
	amd.nodes.ForEach(func(_ gpa.NodeID, node *accessMgrNode) bool {
		msgs.AddAll(node.SetChainAccess(false))
		return true
	})
	return msgs
}

// Access node list has updated for a particular chain.
//
// > Send disabled for nodes not in the access list anymore.
// > Send enabled for new access nodes.
func (amd *accessMgrDist) handleInputAccessNodes(input *inputAccessNodes) gpa.OutMessages {
	//
	// Update the info from the chain perspective.
	if amd.chain == nil {
		initialServers := []*cryptolib.PublicKey{}
		amd.nodes.ForEach(func(_ gpa.NodeID, node *accessMgrNode) bool {
			if node.isServer {
				initialServers = append(initialServers, node.pubKey)
			}
			return true
		})
		amd.chain = newAccessMgrChain(amd.pubKeyToNodeID, initialServers, amd.serversUpdatedCB, amd.log)
	}
	amd.chain.AccessGrantedFor(input.accessNodes)
	//
	// Update the info for each node.
	msgs := gpa.NoMessages()
	amd.nodes.ForEach(func(nodeID gpa.NodeID, node *accessMgrNode) bool {
		msgs.AddAll(node.SetChainAccess(amd.chain.IsAccessGrantedFor(nodeID)))
		return true
	})
	return msgs
}

func (amd *accessMgrDist) handleInputTrustedNodes(input *inputTrustedNodes) gpa.OutMessages {
	msgs := gpa.NoMessages()
	//
	// Setup new nodes.
	trustedIndex := map[gpa.NodeID]bool{}
	for _, trustedNodePubKey := range input.trustedNodes {
		trustedNodeID := amd.pubKeyToNodeID(trustedNodePubKey)
		trustedIndex[trustedNodeID] = true
		if amd.nodes.Has(trustedNodeID) {
			continue
		}
		hasAccess := false
		if amd.chain != nil && amd.chain.IsAccessGrantedFor(trustedNodeID) {
			hasAccess = true
		}
		trustedNode, trustedNodeMsgs := newAccessMgrNode(trustedNodeID, trustedNodePubKey, hasAccess)
		msgs.AddAll(trustedNodeMsgs)
		amd.nodes.Set(trustedNodeID, trustedNode)
	}
	//
	// Disconnect distrusted peers.
	amd.nodes.ForEach(func(nodeID gpa.NodeID, node *accessMgrNode) bool {
		if _, ok := trustedIndex[nodeID]; ok {
			return true
		}
		if amd.chain != nil {
			amd.chain.MarkAsServerFor(node.pubKey, false)
			msgs.AddAll(node.SetChainAccess(false))
		}
		amd.nodes.Delete(nodeID)
		amd.dismissPeerCB(node.pubKey)
		return true
	})
	return msgs
}

func (amd *accessMgrDist) handleMsgAccess(msg *msgAccess) gpa.OutMessages {
	node, exists := amd.nodes.Get(msg.Sender())
	if !exists {
		return nil
	}
	msgs := node.handleMsgAccess(msg)

	if amd.chain != nil {
		amd.chain.MarkAsServerFor(node.pubKey, node.isServer)
	}

	return msgs
}

////////////////////////////////////////////////////////////////////////////////

type accessMgrChain struct {
	access           map[gpa.NodeID]*cryptolib.PublicKey
	server           *shrinkingmap.ShrinkingMap[gpa.NodeID, *cryptolib.PublicKey]
	pubKeyToNodeID   func(*cryptolib.PublicKey) gpa.NodeID
	serversUpdatedCB func([]*cryptolib.PublicKey)
	log              log.Logger
}

func newAccessMgrChain(
	pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID,
	initialServers []*cryptolib.PublicKey,
	serversUpdatedCB func([]*cryptolib.PublicKey),
	log log.Logger,
) *accessMgrChain {
	amc := &accessMgrChain{
		access:           map[gpa.NodeID]*cryptolib.PublicKey{},
		server:           shrinkingmap.New[gpa.NodeID, *cryptolib.PublicKey](),
		pubKeyToNodeID:   pubKeyToNodeID,
		serversUpdatedCB: serversUpdatedCB,
		log:              log,
	}
	for i := range initialServers {
		nodeID := pubKeyToNodeID(initialServers[i])
		amc.server.Set(nodeID, initialServers[i])
	}

	serverNodes := amc.server.Values()
	amc.log.LogDebugf("Chain server nodes updated to %+v on init.", serverNodes)
	amc.serversUpdatedCB(serverNodes)
	return amc
}

func (amc *accessMgrChain) AccessGrantedFor(accessPubKeys []*cryptolib.PublicKey) {
	accessNodes := map[gpa.NodeID]*cryptolib.PublicKey{}
	for _, accessNodePubKey := range accessPubKeys {
		accessNodes[amc.pubKeyToNodeID(accessNodePubKey)] = accessNodePubKey
	}
	amc.access = accessNodes
}

func (amc *accessMgrChain) MarkAsServerFor(nodePubKey *cryptolib.PublicKey, granted bool) {
	nodeID := amc.pubKeyToNodeID(nodePubKey)
	wasServer := amc.server.Has(nodeID)
	if granted {
		amc.server.Set(nodeID, nodePubKey)
	} else {
		amc.server.Delete(nodeID)
	}
	if wasServer != granted {
		serverNodes := amc.server.Values()
		amc.log.LogDebugf("Chain server nodes updated to %+v.", serverNodes)
		amc.serversUpdatedCB(serverNodes)
	}
}

func (amc *accessMgrChain) IsAccessGrantedFor(nodeID gpa.NodeID) bool {
	_, ok := amc.access[nodeID]
	return ok
}

func (amc *accessMgrChain) Disabled() {
	if amc.server.Size() == 0 {
		return
	}
	amc.log.LogDebugf("Chain server nodes updated to [] on dismiss.")
	amc.serversUpdatedCB([]*cryptolib.PublicKey{})
}

////////////////////////////////////////////////////////////////////////////////

type accessMgrNode struct {
	nodeID    gpa.NodeID
	pubKey    *cryptolib.PublicKey
	ourLC     int
	peerLC    int
	hasAccess bool
	isServer  bool
}

func newAccessMgrNode(
	nodeID gpa.NodeID,
	pubKey *cryptolib.PublicKey,
	hasAccess bool,
) (*accessMgrNode, gpa.OutMessages) {
	amn := &accessMgrNode{
		nodeID:    nodeID,
		pubKey:    pubKey,
		ourLC:     1,
		peerLC:    0,
		hasAccess: hasAccess,
		isServer:  false,
	}
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, amn.hasAccess, amn.isServer))
	return amn, msgs
}

func (amn *accessMgrNode) SetChainAccess(access bool) gpa.OutMessages {
	if access {
		return amn.grantAccess()
	}
	return amn.revokeAccess()
}

func (amn *accessMgrNode) grantAccess() gpa.OutMessages {
	if amn.hasAccess {
		return nil
	}
	amn.hasAccess = true
	amn.ourLC++
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, amn.hasAccess, amn.isServer))
	return msgs
}

func (amn *accessMgrNode) revokeAccess() gpa.OutMessages {
	if !amn.hasAccess {
		return nil
	}
	amn.hasAccess = false
	amn.ourLC++
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, amn.hasAccess, amn.isServer))
	return msgs
}

func (amn *accessMgrNode) handleMsgAccess(msg *msgAccess) gpa.OutMessages {
	// This has to be checked before updating the state.
	// > IF /\ m.access = isServer(n, m.src)    \* Peer's info hasn't changed, so we don't need to ack it.
	// >    /\ m.server = H(hasAccess(n, m.src)) \* Our info echoed, so that was an ack.
	// >    /\ m.src_lc >= lClock[n][m.src]            \* Peer's clock is not outdated, we don't need to push it forward.
	// >    /\ m.dst_lc <= lClock[n][n]                \* And the echoed clock don't exceed our clock, so we don't need to push it.
	// > THEN sendAndAck(m, {})
	// > ELSE sendAndAck(m, accessMsgs(n))
	sendDone := true &&
		msg.hasAccess == amn.isServer &&
		msg.isServer == amn.hasAccess &&
		msg.senderLClock >= amn.peerLC &&
		msg.receiverLClock <= amn.ourLC
	//
	// Update isServer and peerLC.
	if msg.senderLClock > amn.peerLC {
		amn.isServer = msg.hasAccess
		amn.peerLC = msg.senderLClock
	}
	//
	// Update ourLC.
	if amn.ourLC <= msg.receiverLClock {
		amn.ourLC = msg.receiverLClock
		if !amn.hasAccess == msg.isServer {
			amn.ourLC++
		}
	}
	//
	// Send message back, if needed.
	if !sendDone {
		return gpa.NoMessages().Add(
			newMsgAccess(msg.Sender(), amn.ourLC, amn.peerLC, amn.hasAccess, amn.isServer),
		)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
