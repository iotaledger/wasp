// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

var errAdminNotDelegated = coreerrors.Register("not delegated to another chain admin").Create()

// claimChainAdmin changes the chain admin to the delegated agentID (if any)
// Checks authorization if the caller is the one to which the admin is delegated
// Note that admin is only changed by the successful call to claimChainAdmin
func claimChainAdmin(ctx isc.Sandbox) {
	state := governance.NewStateWriterFromSandbox(ctx)
	nextAdmin := state.GetChainAdminDelegated()
	if nextAdmin == nil {
		panic(errAdminNotDelegated)
	}
	ctx.RequireCaller(nextAdmin)
	state.SetChainAdmin(nextAdmin)
}

// delegateChainAdmin stores next possible (delegated) chain admin to another agentID
// checks authorization by the current admin
// Two-step process allow/change is in order to avoid mistakes
func delegateChainAdmin(ctx isc.Sandbox, newAdmin isc.AgentID) {
	ctx.Log().Debugf("governance.delegateChainAdmin.begin")
	ctx.RequireCallerIsChainAdmin()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetChainAdminDelegated(newAdmin)
	ctx.Log().Debugf("governance.delegateChainAdmin.success: chain admin delegated to %s", newAdmin.String())
}

func setPayoutAgentID(ctx isc.Sandbox, agent isc.AgentID) {
	ctx.RequireCallerIsChainAdmin()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetPayoutAgentID(agent)
}

func getPayoutAgentID(ctx isc.SandboxView) isc.AgentID {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetPayoutAgentID()
}

func setGasCoinTargetValue(ctx isc.Sandbox, v coin.Value) {
	ctx.RequireCallerIsChainAdmin()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetGasCoinTargetValue(v)
}

func getGasCoinTargetValue(ctx isc.SandboxView) coin.Value {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetGasCoinTargetValue()
}

func getChainAdmin(ctx isc.SandboxView) isc.AgentID {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetChainAdmin()
}
