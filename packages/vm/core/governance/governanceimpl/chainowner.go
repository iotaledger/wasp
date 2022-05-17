// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// claimChainOwnership changes the chain owner to the delegated agentID (if any)
// Checks authorisation if the caller is the one to which the ownership is delegated
// Note that ownership is only changed by the successful call to  claimChainOwnership
func claimChainOwnership(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	state := ctx.State()

	stateDecoder := kvdecoder.New(state, ctx.Log())
	currentOwner := stateDecoder.MustGetAgentID(governance.VarChainOwnerID)
	nextOwner := stateDecoder.MustGetAgentID(governance.VarChainOwnerIDDelegated, currentOwner)

	ctx.Requiref(!nextOwner.Equals(currentOwner), "governance.claimChainOwnership: not delegated to another chain owner")
	ctx.Requiref(nextOwner.Equals(ctx.Caller()), "governance.claimChainOwnership: not authorized")

	state.Set(governance.VarChainOwnerID, codec.EncodeAgentID(nextOwner))
	state.Del(governance.VarChainOwnerIDDelegated)
	ctx.Log().Debugf("governance.chainChainOwner.success: chain owner changed: %s --> %s",
		currentOwner.String(ctx.L1Params().Protocol.Bech32HRP),
		nextOwner.String(ctx.L1Params().Protocol.Bech32HRP),
	)
	return nil
}

// delegateChainOwnership stores next possible (delegated) chain owner to another agentID
// checks authorisation by the current owner
// Two step process allow/change is in order to avoid mistakes
func delegateChainOwnership(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	ctx.RequireCallerIsChainOwner()

	newOwnerID := ctx.Params().MustGetAgentID(governance.ParamChainOwner)
	ctx.State().Set(governance.VarChainOwnerIDDelegated, codec.EncodeAgentID(newOwnerID))
	ctx.Log().Debugf("governance.delegateChainOwnership.success: chain ownership delegated to %s", newOwnerID.String(ctx.L1Params().Protocol.Bech32HRP))
	return nil
}

func getChainOwner(ctx iscp.SandboxView) dict.Dict {
	ret := dict.New()
	ret.Set(governance.ParamChainOwner, ctx.State().MustGet(governance.VarChainOwnerID))
	return ret
}
