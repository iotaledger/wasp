// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This file provides implementation for the governance SC, the ChainNode
// management functions.
//
// State of the SC (the ChainNodes part):
//
//    VarAccessNodeCandidates: 	map[pubKey] => AccessNodeInfo    // A set of Access Node Info.
//    VarAccessNodes:           map[pubKey] => byte[0]           // A set of nodes.
//    VarValidatorNodes:        pubKey[]                         // An ordered list of nodes.
//
package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// SC Query Function handler.
//
//	getChainNodes() => (
//		accessNodeCandidates :: map(pubKey => AccessNodeInfo),
//		accessNodes		     :: map(pubKey => ())
//	)
//
func getChainNodesFuncHandler(ctx iscp.SandboxView) (dict.Dict, error) {
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
	return res, nil
}

// SC Command Function handler.
//
//	candidateNode(
//		candidate:	bool = true		// true, if we are adding the node as a candidate to access nodes.
//		validator:	bool = false	// true, if we also want the node to become a validator.
//		pubKey:		[]byte			// Public key of the node.
//		cert:		[]byte			// Signature by the node over its public key.
//      api:		string = ""		// Optional: API URL for the access node.
//	) => ()
//
// It is possible that after executing `candidateNode(false, false, ...)` a node will stay
// in the list of validators, and will be absent in the candidate or an access node set.
// The node is removed from the list of access nodes immediately, but the validator rotation
// must be initiated by the chain owner explicitly.
//
func candidateNodeFuncHandler(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	paramCandidate := params.MustGetBool(governance.ParamCandidateNodeCandidate, true)
	paramValidator := params.MustGetBool(governance.ParamCandidateNodeValidator, false)
	paramsPubKey := params.MustGetBytes(governance.ParamCandidateNodePubKey)
	paramsCert := params.MustGetBytes(governance.ParamCandidateNodeCert)
	paramsAPI := params.MustGetString(governance.ParamCandidateNodeAPI, "")

	// TODO: Check against the Request sender? This approach is vulnerable to the replay attacks.
	a.Require(ctx.Utils().ED25519().ValidSignature(paramsPubKey, paramsPubKey, paramsCert), "certificate invalid")

	if paramCandidate {
		ani := governance.AccessNodeInfo{
			Validator: paramValidator,
			API:       paramsAPI,
		}
		accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
		accessNodeCandidates.MustSetAt(paramsPubKey, ani.Bytes())
	} else {
		accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
		accessNodeCandidates.MustDelAt(paramsPubKey)
		accessNodes := collections.NewMap(ctx.State(), governance.VarAccessNodes)
		accessNodes.MustDelAt(paramsPubKey)
	}
	return nil, nil
}

// SC Command Function handler.
// Can only be invoked by the chain owner.
//
//  changeAccessNodes(
//    actions: map(pubKey => {0:remove, 1:accept, 2:drop})
//  ) => ()
//
func changeAccessNodesFuncHandler(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireCaller(ctx, []*iscp.AgentID{ctx.ChainOwnerID()})

	accessNodeCandidates := collections.NewMap(ctx.State(), governance.VarAccessNodeCandidates)
	accessNodes := collections.NewMap(ctx.State(), governance.VarAccessNodes)
	paramNodeActions := collections.NewMapReadOnly(ctx.Params(), governance.ParamChangeAccessNodesActions)
	paramNodeActions.MustIterate(func(pubKey, actionBin []byte) bool {
		a.Require(len(actionBin) == 1, "action should be a single byte")
		switch actionBin[0] {
		case 0: // remove from a list of access nodes.
			accessNodes.MustDelAt(pubKey)
		case 1: // accept to a list of access nodes.
			accessNodes.MustSetAt(pubKey, make([]byte, 0))
		case 2: // drop from a list of candidates and the access nodes.
			accessNodes.MustDelAt(pubKey)
			accessNodeCandidates.MustDelAt(pubKey)
		default:
			a.Require(false, "unexpected action")
		}
		return true
	})
	return nil, nil
}
