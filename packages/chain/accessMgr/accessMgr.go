// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accessMgr

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type AccessMgr interface {
	AsGPA() gpa.GPA
}

type Output interface {
	ChainServerNodes(chainID isc.ChainID) []*cryptolib.PublicKey
}

type accessMgrImpl struct {
	nodes          map[gpa.NodeID]*accessMgrNode
	chains         map[isc.ChainID]*accessMgrChain
	pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID
}

var _ gpa.GPA = &accessMgrImpl{}

func NewAccessMgr(pubKeyToNodeID func(*cryptolib.PublicKey) gpa.NodeID) AccessMgr {
	return &accessMgrImpl{
		nodes:          map[gpa.NodeID]*accessMgrNode{},
		chains:         map[isc.ChainID]*accessMgrChain{},
		pubKeyToNodeID: pubKeyToNodeID,
	}
}

// Implements the AccessMgr interface.
func (ami *accessMgrImpl) AsGPA() gpa.GPA {
	return ami
}

// Implements the Output interface.
func (ami *accessMgrImpl) ChainServerNodes(chainID isc.ChainID) []*cryptolib.PublicKey {
	if ch, ok := ami.chains[chainID]; ok {
		return lo.Values(ch.server)
	}
	return []*cryptolib.PublicKey{}
}

// Implements the gpa.GPA interface.
func (ami *accessMgrImpl) Input(input gpa.Input) gpa.OutMessages {
	switch input := input.(type) {
	case *inputNodeStarted:
		return ami.handleInputNodeStarted()
	case *inputChainDisabled:
		return ami.handleInputChainDisabled(input)
	case *inputAccessNodes:
		return ami.handleInputAccessNodes(input)
	case *inputTrustedNodes:
		return ami.handleInputTrustedNodes(input)
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (ami *accessMgrImpl) Message(msg gpa.Message) gpa.OutMessages {
	if msg, ok := msg.(*msgAccess); ok {
		return ami.handleMsgAccess(msg)
	}
	panic(fmt.Errorf("unexpected message %T: %+v", msg, msg))
}

// Implements the gpa.GPA interface.
func (ami *accessMgrImpl) Output() gpa.Output {
	return ami
}

// Implements the gpa.GPA interface.
func (ami *accessMgrImpl) StatusString() string {
	return fmt.Sprintf("{accessMgr, |nodes|=%v, |chains|=%v}", len(ami.nodes), len(ami.chains))
}

// > Notify all the trusted access nodes, that we will not serve the requests anymore.
func (ami *accessMgrImpl) handleInputChainDisabled(input *inputChainDisabled) gpa.OutMessages {
	// ch, ok := ami.chains[input.chainID]
	// if !ok {
	// 	return nil
	// }
	// msgs := gpa.NoMessages()
	// for an := range ch.access {
	// 	if _, ok := ami.trusted[an]; ok {
	// 		msgs.Add(newMsgChainDisabled(an, input.chainID))
	// 	}
	// }
	// return msgs
	return nil // TODO: ...
}

// Node just started. We have to ask other's, what nodes we have access to.
func (ami *accessMgrImpl) handleInputNodeStarted() gpa.OutMessages { // TODO: Handled already in the handleInputTrustedNodes
	return nil // TODO: use it, implement it.
}

// Access node list has updated for a particular chain.
//
// > Send disabled for nodes not in the access list anymore.
// > Send enabled for new access nodes.
func (ami *accessMgrImpl) handleInputAccessNodes(input *inputAccessNodes) gpa.OutMessages {
	// msgs := gpa.NoMessages()
	// if !ami.nodeUp {
	// for range allNodes {
	// 	...
	// }
	// }
	return nil // TODO: ...
}

func (ami *accessMgrImpl) handleInputTrustedNodes(input *inputTrustedNodes) gpa.OutMessages {
	// msgs := gpa.NoMessages()
	// for _, trustedPubKey := range input.trustedNodes {
	// 	trustedNodeID := ami.pubKeyToNodeID(trustedPubKey)
	// 	if _, ok := ami.trusted[trustedNodeID]; !ok {
	// 		msgs.Add(newMsgNodeUp(trustedNodeID)) // TODO: How about message redelivery?
	// 		ami.trusted[trustedNodeID] = trustedPubKey
	// 	}
	// }
	// return msgs
	return nil // TODO: ...
}

func (ami *accessMgrImpl) handleMsgAccess(msg *msgAccess) gpa.OutMessages {
	return nil // TODO: ...
}

////////////////////////////////////////////////////////////////////////////////

type accessMgrChain struct {
	access map[gpa.NodeID]*cryptolib.PublicKey
	server map[gpa.NodeID]*cryptolib.PublicKey
}

////////////////////////////////////////////////////////////////////////////////

type accessMgrNode struct {
	nodeID    gpa.NodeID
	pubKey    *cryptolib.PublicKey
	ourLC     int
	peerLC    int
	accessTo  map[isc.ChainID]bool
	serverFor map[isc.ChainID]bool
}

func (amn *accessMgrNode) SetAccess(chainID isc.ChainID, access bool) gpa.OutMessages {
	if access {
		return amn.grantAccess(chainID)
	}
	return amn.revokeAccess(chainID)
}

func (amn *accessMgrNode) grantAccess(chainID isc.ChainID) gpa.OutMessages {
	if _, ok := amn.accessTo[chainID]; ok {
		return nil
	}
	amn.accessTo[chainID] = true
	amn.ourLC++
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, lo.Keys(amn.accessTo)))
	return msgs
}

func (amn *accessMgrNode) revokeAccess(chainID isc.ChainID) gpa.OutMessages {
	if _, ok := amn.accessTo[chainID]; !ok {
		return nil
	}
	delete(amn.accessTo, chainID)
	amn.ourLC++
	msgs := gpa.NoMessages()
	msgs.Add(newMsgAccess(amn.nodeID, amn.ourLC, amn.peerLC, lo.Keys(amn.accessTo)))
	return msgs
}
