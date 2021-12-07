// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// delegateChainOwnership stores next possible (delegated) chain owner to another agentID
// checks authorisation by the current owner
// Two step process allow/change is in order to avoid mistakes
func delegateChainOwnership(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	a := assert.NewAssert(ctx.Log())
	a.Require(governance.CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "governance.delegateChainOwnership: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	newOwnerID := params.MustGetAgentID(governance.ParamChainOwner)
	ctx.State().Set(governance.VarChainOwnerIDDelegated, codec.EncodeAgentID(newOwnerID))
	ctx.Log().Debugf("governance.delegateChainOwnership.success: chain ownership delegated to %s", newOwnerID.String())
	return nil, nil
}

// claimChainOwnership changes the chain owner to the delegated agentID (if any)
// Checks authorisation if the caller is the one to which the ownership is delegated
// Note that ownership is only changed by the successful call to  claimChainOwnership
func claimChainOwnership(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	state := ctx.State()
	a := assert.NewAssert(ctx.Log())

	stateDecoder := kvdecoder.New(state, ctx.Log())
	currentOwner := stateDecoder.MustGetAgentID(governance.VarChainOwnerID)
	nextOwner := stateDecoder.MustGetAgentID(governance.VarChainOwnerIDDelegated, currentOwner)

	a.Require(!nextOwner.Equals(currentOwner), "governance.claimChainOwnership: not delegated to another chain owner")
	a.Require(nextOwner.Equals(ctx.Caller()), "governance.claimChainOwnership: not authorized")

	state.Set(governance.VarChainOwnerID, codec.EncodeAgentID(nextOwner))
	state.Del(governance.VarChainOwnerIDDelegated)
	ctx.Log().Debugf("governance.chainChainOwner.success: chain owner changed: %s --> %s",
		currentOwner.String(), nextOwner.String())
	return nil, nil
}

func getChainOwner(ctx iscp.SandboxView) (dict.Dict, error) {
	ret := dict.New()
	ret.Set(governance.ParamChainOwner, ctx.State().MustGet(governance.VarChainOwnerID))
	return ret, nil
}
