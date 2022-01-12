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
//
package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// SC Query Function handler.
//
//  getChainNodes() => (
//      accessNodeCandidates :: map(pubKey => AccessNodeInfo),
//      accessNodes          :: map(pubKey => ())
//  )
//
func getChainNodesFuncHandler(ctx iscp.SandboxView) (dict.Dict, error) {
	res := dict.New()
	ac := collections.NewMap(res, string(governance.ParamGetChainNodesAccessNodeCandidates))
	an := collections.NewMap(res, string(governance.ParamGetChainNodesAccessNodes))
	collections.NewMapReadOnly(ctx.State(), governance.VarAccessNodeCandidates).MustIterate(func(key, value []byte) bool {
		ac.MustSetAt(key, value)
		return true
	})
	collections.NewMapReadOnly(ctx.State(), governance.VarAccessNodes).MustIterate(func(key, value []byte) bool {
		an.MustSetAt(key, value)
		return true
	})
	return res, nil
}

// SC Command Function handler.
// Can only be invoked by the access node owner (verified via the Certificate field).
//
//  addCandidateNode(
//      accessNodeInfo{NodePubKey, Certificate, ForCommittee, AccessAPI}
//  ) => ()
//
func addCandidateNodeFuncHandler(ctx iscp.Sandbox) (dict.Dict, error) {
	ani := governance.NewAccessNodeInfoFromAddCandidateNodeParams(ctx)
	ctx.Requiref(ani.ValidateCertificate(ctx), "certificate invalid")

	accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
	accessNodeCandidates.MustSetAt(ani.NodePubKey, ani.Bytes())

	return nil, nil
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
func revokeAccessNodeFuncHandler(ctx iscp.Sandbox) (dict.Dict, error) {
	ani := governance.NewAccessNodeInfoFromRevokeAccessNodeParams(ctx)
	ctx.Requiref(ani.ValidateCertificate(ctx), "certificate invalid")

	accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
	accessNodeCandidates.MustDelAt(ani.NodePubKey)
	accessNodes := collections.NewMap(ctx.State(), governance.VarAccessNodes)
	accessNodes.MustDelAt(ani.NodePubKey)

	return nil, nil
}

// SC Command Function handler.
// Can only be invoked by the chain owner.
//
//  changeAccessNodes(
//      actions: map(pubKey => ChangeAccessNodeAction)
//  ) => ()
//
func changeAccessNodesFuncHandler(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.RequireCallerIsChainOwner()

	accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
	accessNodes := collections.NewMap(ctx.State(), governance.VarAccessNodes)
	paramNodeActions := collections.NewMapReadOnly(ctx.Params(), string(governance.ParamChangeAccessNodesActions))
	paramNodeActions.MustIterate(func(pubKey, actionBin []byte) bool {
		ctx.Requiref(len(actionBin) == 1, "action should be a single byte")
		switch governance.ChangeAccessNodeAction(actionBin[0]) {
		case governance.ChangeAccessNodeActionRemove:
			accessNodes.MustDelAt(pubKey)
		case governance.ChangeAccessNodeActionAccept:
			accessNodes.MustSetAt(pubKey, make([]byte, 0))
		case governance.ChangeAccessNodeActionDrop:
			accessNodes.MustDelAt(pubKey)
			accessNodeCandidates.MustDelAt(pubKey)
		default:
			ctx.Requiref(false, "unexpected action")
		}
		return true
	})
	return nil, nil
}
