// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// The algorithm here follows the specification `WaspChainAccessNodesV4.tla`.
// The specification actions are mapped to GPA inputs here as follows:
//
//   - ChainActivate  -- first reception of inputAccessNodes.
//   - ChainDeactivate -- inputChainDisabled.
//   - AccessNodeAdd -- inputAccessNodes, then compare with info we had before.
//   - AccessNodeDel --  inputAccessNodes, then compare with info we had before.
//   - Reboot -- inputTrustedNodes.
package amDist

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/core/logger"
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
	nodes            map[gpa.NodeID]*accessMgrNode             // State for each peer.
	chains           map[isc.ChainID]*accessMgrChain           // State for each chain.
	pubKeyToNodeID   func(*cryptolib.PublicKey) gpa.NodeID     // Convert PubKeys to NodeIDs.
	serversUpdatedCB func(isc.ChainID, []*cryptolib.PublicKey) // Called when a set fo servers has changed for a chain.
	dismissPeerCB    func(*cryptolib.PublicKey)                // To stop redelivery at the upper layer.
	log              *logger.Logger
}

var _ gpa.GPA = &accessMgrDist{}

func NewAccessMgr(
	pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID,
	serversUpdatedCB func(isc.ChainID, []*cryptolib.PublicKey),
	dismissPeerCB func(*cryptolib.PublicKey),
	log *logger.Logger,
) AccessMgr {
	return &accessMgrDist{
		nodes:            map[gpa.NodeID]*accessMgrNode{},
		chains:           map[isc.ChainID]*accessMgrChain{},
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
	if ch, ok := amd.chains[chainID]; ok {
		return lo.Values(ch.server)
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
	return fmt.Sprintf("{accessMgr, |nodes|=%v, |chains|=%v}", len(amd.nodes), len(amd.chains))
}

// > Notify all the trusted access nodes, that we will not serve the requests anymore.
func (amd *accessMgrDist) handleInputChainDisabled(input *inputChainDisabled) gpa.OutMessages {
	msgs := gpa.NoMessages()
	delete(amd.chains, input.chainID)
	for _, node := range amd.nodes {
		msgs.AddAll(node.SetChainAccess(input.chainID, false))
	}
	return msgs
}

// Access node list has updated for a particular chain.
//
// > Send disabled for nodes not in the access list anymore.
// > Send enabled for new access nodes.
func (amd *accessMgrDist) handleInputAccessNodes(input *inputAccessNodes) gpa.OutMessages {
	//
	// Update the info from the chain perspective.
	chain, ok := amd.chains[input.chainID]
	if !ok {
		chain = newAccessMgrChain(input.chainID, amd.pubKeyToNodeID, amd.serversUpdatedCB)
		amd.chains[input.chainID] = chain
	}
	chain.AccessGrantedFor(input.accessNodes)
	//
	// Update the info for each node.
	msgs := gpa.NoMessages()
	for nodeID, node := range amd.nodes {
		msgs.AddAll(node.SetChainAccess(input.chainID, chain.IsAccessGrantedFor(nodeID)))
	}
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
		if _, ok := amd.nodes[trustedNodeID]; ok {
			continue
		}
		accessFor := newChainSet()
		for chainID, chain := range amd.chains {
			if chain.IsAccessGrantedFor(trustedNodeID) {
				accessFor.Add(chainID)
			}
		}
		trustedNode, trustedNodeMsgs := newAccessMgrNode(trustedNodeID, trustedNodePubKey, accessFor)
		msgs.AddAll(trustedNodeMsgs)
		amd.nodes[trustedNodeID] = trustedNode
	}
	//
	// Disconnect distrusted peers.
	for nodeID, node := range amd.nodes {
		if _, ok := trustedIndex[nodeID]; ok {
			continue
		}
		delete(amd.nodes, nodeID)
		amd.dismissPeerCB(node.pubKey)
	}
	return msgs
}

func (amd *accessMgrDist) handleMsgAccess(msg *msgAccess) gpa.OutMessages {
	node, nodeFound := amd.nodes[msg.Sender()]
	if !nodeFound {
		return nil
	}
	msgs := node.handleMsgAccess(msg)

	for chainID, chain := range amd.chains {
		chain.MarkAsServerFor(node.pubKey, node.accessFor.Has(chainID))
	}

	return msgs
}

////////////////////////////////////////////////////////////////////////////////

type accessMgrChain struct {
	chainID          isc.ChainID
	access           map[gpa.NodeID]*cryptolib.PublicKey
	server           map[gpa.NodeID]*cryptolib.PublicKey
	pubKeyToNodeID   func(*cryptolib.PublicKey) gpa.NodeID
	serversUpdatedCB func(isc.ChainID, []*cryptolib.PublicKey)
}

func newAccessMgrChain(
	chainID isc.ChainID,
	pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID,
	serversUpdatedCB func(isc.ChainID, []*cryptolib.PublicKey),
) *accessMgrChain {
	return &accessMgrChain{
		chainID:          chainID,
		access:           map[gpa.NodeID]*cryptolib.PublicKey{},
		server:           map[gpa.NodeID]*cryptolib.PublicKey{},
		pubKeyToNodeID:   pubKeyToNodeID,
		serversUpdatedCB: serversUpdatedCB,
	}
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
	_, wasServer := amc.server[nodeID]
	if granted {
		amc.server[nodeID] = nodePubKey
	} else {
		delete(amc.server, nodeID)
	}
	if wasServer != granted {
		amc.serversUpdatedCB(amc.chainID, lo.Values(amc.server))
	}
}

func (amc *accessMgrChain) IsAccessGrantedFor(nodeID gpa.NodeID) bool {
	_, ok := amc.access[nodeID]
	return ok
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
	elements map[isc.ChainID]interface{}
}

func newChainSet() *chainSet {
	return &chainSet{elements: map[isc.ChainID]interface{}{}}
}

func (cs *chainSet) Add(elem isc.ChainID) {
	cs.elements[elem] = nil
}

func (cs *chainSet) Delete(elem isc.ChainID) {
	delete(cs.elements, elem)
}

func (cs *chainSet) Has(elem isc.ChainID) bool {
	_, ok := cs.elements[elem]
	return ok
}

func (cs *chainSet) AsSlice() []isc.ChainID {
	return lo.Keys(cs.elements)
}

func (cs *chainSet) FromSlice(els []isc.ChainID) {
	cs.elements = map[isc.ChainID]interface{}{}
	for _, el := range els {
		cs.elements[el] = nil
	}
}

func (cs *chainSet) Equals(other *chainSet) bool {
	if len(cs.elements) != len(other.elements) {
		return false
	}
	for e := range cs.elements {
		if _, ok := other.elements[e]; !ok {
			return false
		}
	}
	return true
}
