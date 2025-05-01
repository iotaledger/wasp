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

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

type AccessMgr interface {
	AsGPA() gpa.GPA
}

type Output interface {
	ChainServerNodes(chainID isc.ChainID) []*cryptolib.PublicKey
}

type accessMgrDist struct {
	nodes            *shrinkingmap.ShrinkingMap[gpa.NodeID, *accessMgrNode]   // State for each peer.
	chains           *shrinkingmap.ShrinkingMap[isc.ChainID, *accessMgrChain] // State for each chain.
	pubKeyToNodeID   func(*cryptolib.PublicKey) gpa.NodeID                    // Convert PubKeys to NodeIDs.
	serversUpdatedCB func(isc.ChainID, []*cryptolib.PublicKey)                // Called when a set of servers has changed for a chain.
	dismissPeerCB    func(*cryptolib.PublicKey)                               // To stop redelivery at the upper layer.
	log              log.Logger
}

var _ gpa.GPA = &accessMgrDist{}

func NewAccessMgr(
	pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID,
	serversUpdatedCB func(chainID isc.ChainID, servers []*cryptolib.PublicKey),
	dismissPeerCB func(*cryptolib.PublicKey),
	log log.Logger,
) AccessMgr {
	return &accessMgrDist{
		nodes:            shrinkingmap.New[gpa.NodeID, *accessMgrNode](),
		chains:           shrinkingmap.New[isc.ChainID, *accessMgrChain](),
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
func (amd *accessMgrDist) ChainServerNodes(chainID isc.ChainID) []*cryptolib.PublicKey {
	if chain, exists := amd.chains.Get(chainID); exists {
		return chain.server.Values()
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
	return fmt.Sprintf("{accessMgr, |nodes|=%v, |chains|=%v}", amd.nodes.Size(), amd.chains.Size())
}

// > Notify all the trusted access nodes, that we will not serve the requests anymore.
func (amd *accessMgrDist) handleInputChainDisabled(input *inputChainDisabled) gpa.OutMessages {
	chain, exists := amd.chains.Get(input.chainID)
	if !exists {
		return nil // Already disabled.
	}
	chain.Disabled()
	amd.chains.Delete(input.chainID)
	msgs := gpa.NoMessages()
	amd.nodes.ForEach(func(_ gpa.NodeID, node *accessMgrNode) bool {
		msgs.AddAll(node.SetChainAccess(input.chainID, false))
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
	chain, exists := amd.chains.Get(input.chainID)
	if !exists {
		initialServers := []*cryptolib.PublicKey{}
		amd.nodes.ForEach(func(_ gpa.NodeID, node *accessMgrNode) bool {
			if node.serverFor.Has(input.chainID) {
				initialServers = append(initialServers, node.pubKey)
			}
			return true
		})
		chain = newAccessMgrChain(input.chainID, amd.pubKeyToNodeID, initialServers, amd.serversUpdatedCB, amd.log)
		amd.chains.Set(input.chainID, chain)
	}
	chain.AccessGrantedFor(input.accessNodes)
	//
	// Update the info for each node.
	msgs := gpa.NoMessages()
	amd.nodes.ForEach(func(nodeID gpa.NodeID, node *accessMgrNode) bool {
		msgs.AddAll(node.SetChainAccess(input.chainID, chain.IsAccessGrantedFor(nodeID)))
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
		accessFor := newChainSet()
		amd.chains.ForEach(func(chainID isc.ChainID, chain *accessMgrChain) bool {
			if chain.IsAccessGrantedFor(trustedNodeID) {
				accessFor.Add(chainID)
			}
			return true
		})
		trustedNode, trustedNodeMsgs := newAccessMgrNode(trustedNodeID, trustedNodePubKey, accessFor)
		msgs.AddAll(trustedNodeMsgs)
		amd.nodes.Set(trustedNodeID, trustedNode)
	}
	//
	// Disconnect distrusted peers.
	amd.nodes.ForEach(func(nodeID gpa.NodeID, node *accessMgrNode) bool {
		if _, ok := trustedIndex[nodeID]; ok {
			return true
		}
		amd.chains.ForEach(func(_ isc.ChainID, chain *accessMgrChain) bool {
			chain.MarkAsServerFor(node.pubKey, false)
			msgs.AddAll(node.SetChainAccess(chain.chainID, false))
			return true
		})
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

	amd.chains.ForEach(func(chainID isc.ChainID, chain *accessMgrChain) bool {
		chain.MarkAsServerFor(node.pubKey, node.serverFor.Has(chainID))
		return true
	})

	return msgs
}

////////////////////////////////////////////////////////////////////////////////

type accessMgrChain struct {
	chainID          isc.ChainID
	access           map[gpa.NodeID]*cryptolib.PublicKey
	server           *shrinkingmap.ShrinkingMap[gpa.NodeID, *cryptolib.PublicKey]
	pubKeyToNodeID   func(*cryptolib.PublicKey) gpa.NodeID
	serversUpdatedCB func(isc.ChainID, []*cryptolib.PublicKey)
	log              log.Logger
}

func newAccessMgrChain(
	chainID isc.ChainID,
	pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID,
	initialServers []*cryptolib.PublicKey,
	serversUpdatedCB func(isc.ChainID, []*cryptolib.PublicKey),
	log log.Logger,
) *accessMgrChain {
	amc := &accessMgrChain{
		chainID:          chainID,
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
	amc.log.LogDebugf("Chain %v server nodes updated to %+v on init.", amc.chainID.ShortString(), serverNodes)
	amc.serversUpdatedCB(amc.chainID, serverNodes)
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
		amc.log.LogDebugf("Chain %v server nodes updated to %+v.", amc.chainID.ShortString(), serverNodes)
		amc.serversUpdatedCB(amc.chainID, serverNodes)
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
	amc.log.LogDebugf("Chain %v server nodes updated to [] on dismiss.", amc.chainID.ShortString())
	amc.serversUpdatedCB(amc.chainID, []*cryptolib.PublicKey{})
}

////////////////////////////////////////////////////////////////////////////////

type accessMgrNode struct {
	nodeID    gpa.NodeID
	pubKey    *cryptolib.PublicKey
	ourLC     int
	peerLC    int
	accessFor *chainSet
	serverFor *chainSet
}

func newAccessMgrNode(
	nodeID gpa.NodeID,
	pubKey *cryptolib.PublicKey,
	accessFor *chainSet,
) (*accessMgrNode, gpa.OutMessages) {
	amn := &accessMgrNode{
		nodeID:    nodeID,
		pubKey:    pubKey,
		ourLC:     1,
		peerLC:    0,
		accessFor: accessFor,
		serverFor: newChainSet(),
	}
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, amn.accessFor.AsSlice(), amn.serverFor.AsSlice()))
	return amn, msgs
}

func (amn *accessMgrNode) SetChainAccess(chainID isc.ChainID, access bool) gpa.OutMessages {
	if access {
		return amn.grantAccess(chainID)
	}
	return amn.revokeAccess(chainID)
}

func (amn *accessMgrNode) grantAccess(chainID isc.ChainID) gpa.OutMessages {
	if amn.accessFor.Has(chainID) {
		return nil
	}
	amn.accessFor.Add(chainID)
	amn.ourLC++
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, amn.accessFor.AsSlice(), amn.serverFor.AsSlice()))
	return msgs
}

func (amn *accessMgrNode) revokeAccess(chainID isc.ChainID) gpa.OutMessages {
	if !amn.accessFor.Has(chainID) {
		return nil
	}
	amn.accessFor.Delete(chainID)
	amn.ourLC++
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, amn.accessFor.AsSlice(), amn.serverFor.AsSlice()))
	return msgs
}

func (amn *accessMgrNode) handleMsgAccess(msg *msgAccess) gpa.OutMessages {
	// This has to be checked before updating the state.
	// > IF /\ m.access = serverForChains(n, m.src)    \* Peer's info hasn't changed, so we don't need to ack it.
	// >    /\ m.server = H(accessForChains(n, m.src)) \* Our info echoed, so that was an ack.
	// >    /\ m.src_lc >= lClock[n][m.src]            \* Peer's clock is not outdated, we don't need to push it forward.
	// >    /\ m.dst_lc <= lClock[n][n]                \* And the echoed clock don't exceed our clock, so we don't need to push it.
	// > THEN sendAndAck(m, {})
	// > ELSE sendAndAck(m, accessMsgs(n))
	sendDone := true &&
		util.Same(msg.accessForChains, amn.serverFor.AsSlice()) &&
		util.Same(msg.serverForChains, amn.accessFor.AsSlice()) &&
		msg.senderLClock >= amn.peerLC &&
		msg.receiverLClock <= amn.ourLC
	//
	// Update serverFor and peerLC.
	if msg.senderLClock > amn.peerLC {
		amn.serverFor.FromSlice(msg.accessForChains)
		amn.peerLC = msg.senderLClock
	}
	//
	// Update ourLC.
	if amn.ourLC <= msg.receiverLClock {
		amn.ourLC = msg.receiverLClock
		msgServerFor := newChainSet()
		msgServerFor.FromSlice(msg.serverForChains)
		if !amn.accessFor.Equals(msgServerFor) {
			amn.ourLC++
		}
	}
	//
	// Send message back, if needed.
	if !sendDone {
		return gpa.NoMessages().Add(
			newMsgAccess(msg.Sender(), amn.ourLC, amn.peerLC, amn.accessFor.AsSlice(), amn.serverFor.AsSlice()),
		)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type chainSet struct {
	elements *shrinkingmap.ShrinkingMap[isc.ChainID, struct{}]
}

func newChainSet() *chainSet {
	return &chainSet{elements: shrinkingmap.New[isc.ChainID, struct{}]()}
}

func (cs *chainSet) Add(elem isc.ChainID) {
	cs.elements.Set(elem, struct{}{})
}

func (cs *chainSet) Delete(elem isc.ChainID) {
	cs.elements.Delete(elem)
}

func (cs *chainSet) Has(elem isc.ChainID) bool {
	return cs.elements.Has(elem)
}

func (cs *chainSet) AsSlice() []isc.ChainID {
	return cs.elements.Keys()
}

func (cs *chainSet) FromSlice(els []isc.ChainID) {
	cs.elements = shrinkingmap.New[isc.ChainID, struct{}]()
	for _, el := range els {
		cs.elements.Set(el, struct{}{})
	}
}

func (cs *chainSet) Equals(other *chainSet) bool {
	if cs.elements.Size() != other.elements.Size() {
		return false
	}

	equal := true
	cs.elements.ForEach(func(ci isc.ChainID, s struct{}) bool {
		if !other.elements.Has(ci) {
			equal = false
			return false
		}
		return true
	})

	return equal
}
