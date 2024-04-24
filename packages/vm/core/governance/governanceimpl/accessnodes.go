// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This file provides implementation for the governance SC, the ChainNode
// management functions.
//
// State of the SC (the ChainNodes part):
//
//	VarAccessNodeCandidates:  map[pubKey] => AccessNodeInfo    // A set of Access Node Info.
//	VarAccessNodes:           map[pubKey] => byte[0]           // A set of nodes.
//	VarValidatorNodes:        pubKey[]                         // An ordered list of nodes.
package governanceimpl

import (
	"encoding/base64"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

var errInvalidCertificate = coreerrors.Register("invalid certificate").Create()

// SC Command Function handler.
// Can only be invoked by the access node owner (verified via the Certificate field).
//
//	addCandidateNode(
//	    accessNodeInfo{NodePubKey, Certificate, ForCommittee, AccessAPI}
//	) => ()
func addCandidateNode(ctx isc.Sandbox) dict.Dict {
	ani := governance.AccessNodeInfoFromAddCandidateNodeParams(ctx)
	if !ani.ValidateCertificate() {
		panic(errInvalidCertificate)
	}
	pubKeyStr := base64.StdEncoding.EncodeToString(ani.NodePubKey)

	state := ctx.State()
	governance.AccessNodeCandidatesMap(state).SetAt(ani.NodePubKey, ani.Bytes())
	ctx.Log().Infof("Governance::AddCandidateNode: accessNodeCandidate added, pubKey=%s", pubKeyStr)

	if ctx.ChainOwnerID().Equals(ctx.Request().SenderAccount()) {
		governance.AccessNodesMap(state).SetAt(ani.NodePubKey, codec.Bool.Encode(true))
		ctx.Log().Infof("Governance::AddCandidateNode: accessNode added, pubKey=%s", pubKeyStr)
	}

	return nil
}

// SC Command Function handler.
// Can only be invoked by the access node owner (verified via the Certificate field).
//
//	revokeAccessNode(
//	    accessNodeInfo{NodePubKey, Certificate}
//	) => ()
//
// It is possible that after executing `revokeAccessNode(...)` a node will stay
// in the list of validators, and will be absent in the candidate or an access node set.
// The node is removed from the list of access nodes immediately, but the validator rotation
// must be initiated by the chain owner explicitly.
func revokeAccessNode(ctx isc.Sandbox) dict.Dict {
	ani := governance.AccessNodeInfoFromRevokeAccessNodeParams(ctx)
	if !ani.ValidateCertificate() {
		panic(errInvalidCertificate)
	}

	state := ctx.State()
	governance.AccessNodeCandidatesMap(state).DelAt(ani.NodePubKey)
	governance.AccessNodesMap(state).DelAt(ani.NodePubKey)

	return nil
}

var errInvalidAction = coreerrors.Register("invalid action").Create()

// SC Command Function handler.
// Can only be invoked by the chain owner.
//
//	changeAccessNodes(
//	    actions: map(pubKey => ChangeAccessNodeAction)
//	) => ()
func changeAccessNodes(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	state := ctx.State()
	accessNodeCandidates := governance.AccessNodeCandidatesMap(state)
	accessNodes := governance.AccessNodesMap(state)
	paramNodeActions := collections.NewMapReadOnly(ctx.Params(), governance.ParamChangeAccessNodesActions)
	ctx.Log().Debugf("changeAccessNodes: actions len: %d", paramNodeActions.Len())

	paramNodeActions.Iterate(func(pubKey, actionBin []byte) bool {
		if len(actionBin) != 1 {
			panic(errInvalidAction)
		}
		switch governance.ChangeAccessNodeAction(actionBin[0]) {
		case governance.ChangeAccessNodeActionRemove:
			accessNodes.DelAt(pubKey)
		case governance.ChangeAccessNodeActionAccept:
			// TODO should the list of candidates be checked? we are just adding any pubkey
			accessNodes.SetAt(pubKey, codec.Bool.Encode(true))
			// TODO should the node be removed from the list of candidates? // accessNodeCandidates.DelAt(pubKey)
		case governance.ChangeAccessNodeActionDrop:
			accessNodes.DelAt(pubKey)
			accessNodeCandidates.DelAt(pubKey)
		default:
			panic(errInvalidAction)
		}
		return true
	})
	return nil
}

// SC Query Function handler.
//
//	getChainNodes() => (
//	    accessNodeCandidates :: map(pubKey => AccessNodeInfo),
//	    accessNodes          :: map(pubKey => ())
//	)
func getChainNodes(ctx isc.SandboxView) dict.Dict {
	res := dict.New()
	candidates := collections.NewMap(res, governance.ParamGetChainNodesAccessNodeCandidates)
	nodes := collections.NewMap(res, governance.ParamGetChainNodesAccessNodes)
	stateR := ctx.StateR()
	governance.AccessNodeCandidatesMapR(stateR).Iterate(func(key, value []byte) bool {
		candidates.SetAt(key, value)
		return true
	})
	governance.AccessNodesMapR(stateR).IterateKeys(func(key []byte) bool {
		nodes.SetAt(key, []byte{0x01})
		return true
	})
	return res
}
