// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

var errOwnerNotDelegated = coreerrors.Register("not delegated to another chain owner").Create()

// claimChainOwnership changes the chain owner to the delegated agentID (if any)
// Checks authorization if the caller is the one to which the ownership is delegated
// Note that ownership is only changed by the successful call to  claimChainOwnership
func claimChainOwnership(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	state := ctx.State()

	stateDecoder := kvdecoder.New(state, ctx.Log())
	currentOwner := stateDecoder.MustGetAgentID(governance.VarChainOwnerID)
	nextOwner := stateDecoder.MustGetAgentID(governance.VarChainOwnerIDDelegated, currentOwner)

	if nextOwner.Equals(currentOwner) {
		panic(errOwnerNotDelegated)
	}
	ctx.RequireCaller(nextOwner)

	state.Set(governance.VarChainOwnerID, codec.AgentID.Encode(nextOwner))
	state.Del(governance.VarChainOwnerIDDelegated)
	ctx.Log().Debugf("governance.chainChainOwner.success: chain owner changed: %s --> %s",
		currentOwner.String(),
		nextOwner.String(),
	)
	return nil
}

// delegateChainOwnership stores next possible (delegated) chain owner to another agentID
// checks authorization by the current owner
// Two-step process allow/change is in order to avoid mistakes
func delegateChainOwnership(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	ctx.RequireCallerIsChainOwner()

	newOwnerID := ctx.Params().MustGetAgentID(governance.ParamChainOwner)
	ctx.State().Set(governance.VarChainOwnerIDDelegated, codec.AgentID.Encode(newOwnerID))
	ctx.Log().Debugf("governance.delegateChainOwnership.success: chain ownership delegated to %s", newOwnerID.String())
	return nil
}

func setPayoutAgentID(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	agent := ctx.Params().MustGetAgentID(governance.ParamSetPayoutAgentID)
	ctx.State().Set(governance.VarPayoutAgentID, codec.AgentID.Encode(agent))
	return nil
}

func getPayoutAgentID(ctx isc.SandboxView) dict.Dict {
	ret := dict.New()
	ret.Set(governance.ParamSetPayoutAgentID, ctx.StateR().Get(governance.VarPayoutAgentID))
	return ret
}

func setMinCommonAccountBalance(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	minCommonAccountBalance := ctx.Params().MustGetUint64(governance.ParamSetMinCommonAccountBalance)
	ctx.State().Set(governance.VarMinBaseTokensOnCommonAccount, codec.Uint64.Encode(minCommonAccountBalance))
	return nil
}

func getMinCommonAccountBalance(ctx isc.SandboxView) dict.Dict {
	return dict.Dict{
		governance.ParamSetMinCommonAccountBalance: ctx.StateR().Get(governance.VarMinBaseTokensOnCommonAccount),
	}
}

func getChainOwner(ctx isc.SandboxView) dict.Dict {
	ret := dict.New()
	ret.Set(governance.ParamChainOwner, ctx.StateR().Get(governance.VarChainOwnerID))
	return ret
}
