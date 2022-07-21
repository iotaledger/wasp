// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This file provides implementation for the governance SC, the ChainNode
// management functions.
//
// State of the SC (the ChainNodes part):
//
//    VarAccessNodeCandidates:  map[pubKey] => AccessNodeInfo    // A set of Access Node Info.
//    VarAccessNodes:           map[pubKey] => byte[0]           // A set of nodes.
//    VarValidatorNodes:        pubKey[]                         // An ordered list of nodes.
package governanceimpl

import (
	"encoding/base64"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// SC Command Function handler.
// Can only be invoked by the access node owner (verified via the Certificate field).
//
//  addCandidateNode(
//      accessNodeInfo{NodePubKey, Certificate, ForCommittee, AccessAPI}
//  ) => ()
//
func addCandidateNode(ctx iscp.Sandbox) dict.Dict {
	ani := governance.NewAccessNodeInfoFromAddCandidateNodeParams(ctx)
	ctx.Requiref(ani.ValidateCertificate(ctx), "certificate invalid")
	pubKeyStr := base64.StdEncoding.EncodeToString(ani.NodePubKey)

	accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
	accessNodeCandidates.MustSetAt(ani.NodePubKey, ani.Bytes())
	ctx.Log().Infof("Governance::AddCandidateNode: accessNodeCandidate added, pubKey=%s", pubKeyStr)

	if ctx.ChainOwnerID().Equals(ctx.Request().SenderAccount()) {
		accessNodes := collections.NewMap(ctx.State(), governance.VarAccessNodes)
		accessNodes.MustSetAt(ani.NodePubKey, make([]byte, 0))
		ctx.Log().Infof("Governance::AddCandidateNode: accessNode added, pubKey=%s", pubKeyStr)
	}

	return nil
}

// SC Command Function handler.
// Can only be invoked by the access node owner (verified via the Certificate field).
//
//  revokeAccessNode(
//      accessNodeInfo{NodePubKey, Certificate}
//  ) => ()
//
// It is possible that after executing `revokeAccessNode(...)` a node will stay
// in the list of validators, and will be absent in the candidate or an access node set.
// The node is removed from the list of access nodes immediately, but the validator rotation
// must be initiated by the chain owner explicitly.
//
func revokeAccessNode(ctx iscp.Sandbox) dict.Dict {
	ani := governance.NewAccessNodeInfoFromRevokeAccessNodeParams(ctx)
	ctx.Requiref(ani.ValidateCertificate(ctx), "certificate invalid")

	accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
	accessNodeCandidates.MustDelAt(ani.NodePubKey)
	accessNodes := collections.NewMap(ctx.State(), governance.VarAccessNodes)
	accessNodes.MustDelAt(ani.NodePubKey)

	return nil
}

// SC Command Function handler.
// Can only be invoked by the chain owner.
//
//  changeAccessNodes(
//      actions: map(pubKey => ChangeAccessNodeAction)
//  ) => ()
//
func changeAccessNodes(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
	accessNodes := collections.NewMap(ctx.State(), governance.VarAccessNodes)
	paramNodeActions := collections.NewMapReadOnly(ctx.Params(), governance.ParamChangeAccessNodesActions)
	paramNodeActions.MustIterate(func(pubKey, actionBin []byte) bool {
		ctx.Requiref(len(actionBin) == 1, "action should be a single byte")
		switch governance.ChangeAccessNodeAction(actionBin[0]) {
		case governance.ChangeAccessNodeActionRemove:
			accessNodes.MustDelAt(pubKey)
		case governance.ChangeAccessNodeActionAccept:
			// TODO should the list of candidates be checked? we are just adding any pubkey
			accessNodes.MustSetAt(pubKey, make([]byte, 0))
			// TODO should the node be removed from the list of candidates? // accessNodeCandidates.MustDelAt(pubKey)
		case governance.ChangeAccessNodeActionDrop:
			accessNodes.MustDelAt(pubKey)
			accessNodeCandidates.MustDelAt(pubKey)
		default:
			ctx.Requiref(false, "unexpected action")
		}
		return true
	})
	return nil
}

// SC Query Function handler.
//
//  getChainNodes() => (
//      accessNodeCandidates :: map(pubKey => AccessNodeInfo),
//      accessNodes          :: map(pubKey => ())
//  )
//
func getChainNodes(ctx iscp.SandboxView) dict.Dict {
	res := dict.New()
	ac := collections.NewMap(res, governance.ParamGetChainNodesAccessNodeCandidates)
	an := collections.NewMap(res, governance.ParamGetChainNodesAccessNodes)
	collections.NewMapReadOnly(ctx.State(), governance.VarAccessNodeCandidates).MustIterate(func(key, value []byte) bool {
		ac.MustSetAt(key, value)
		return true
	})
	collections.NewMapReadOnly(ctx.State(), governance.VarAccessNodes).MustIterate(func(key, value []byte) bool {
		an.MustSetAt(key, value)
		return true
	})
	return res
}
