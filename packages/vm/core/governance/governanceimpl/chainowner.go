// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

var errOwnerNotDelegated = coreerrors.Register("not delegated to another chain owner").Create()

// claimChainOwnership changes the chain owner to the delegated agentID (if any)
// Checks authorization if the caller is the one to which the ownership is delegated
// Note that ownership is only changed by the successful call to  claimChainOwnership
func claimChainOwnership(ctx isc.Sandbox) {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	state := governance.NewStateWriterFromSandbox(ctx)

	currentOwner := state.GetChainOwnerID()
	nextOwner := state.GetChainOwnerIDDelegated()

	if nextOwner == nil {
		panic(errOwnerNotDelegated)
	}
	ctx.RequireCaller(nextOwner)

	state.SetChainOwnerID(nextOwner)
	ctx.Log().Debugf("governance.chainChainOwner.success: chain owner changed: %s --> %s",
		currentOwner.String(),
		nextOwner.String(),
	)
}

// delegateChainOwnership stores next possible (delegated) chain owner to another agentID
// checks authorization by the current owner
// Two-step process allow/change is in order to avoid mistakes
func delegateChainOwnership(ctx isc.Sandbox, newOwnerID isc.AgentID) {
	ctx.Log().Debugf("governance.delegateChainOwnership.begin")
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetChainOwnerIDDelegated(newOwnerID)
	ctx.Log().Debugf("governance.delegateChainOwnership.success: chain ownership delegated to %s", newOwnerID.String())
}

func setPayoutAgentID(ctx isc.Sandbox, agent isc.AgentID) {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetPayoutAgentID(agent)
}

func getPayoutAgentID(ctx isc.SandboxView) isc.AgentID {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetPayoutAgentID()
}

func setMinCommonAccountBalance(ctx isc.Sandbox, minCommonAccountBalance uint64) {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetMinCommonAccountBalance(minCommonAccountBalance)
}

func getMinCommonAccountBalance(ctx isc.SandboxView) uint64 {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetMinCommonAccountBalance()
}

func getChainOwner(ctx isc.SandboxView) isc.AgentID {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetChainOwnerID()
}
