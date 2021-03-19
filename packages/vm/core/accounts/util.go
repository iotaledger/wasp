package accounts

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// Accrue calls "deposit" entry point of the accounts contract.
// Can only be called from full sandbox context
func Accrue(ctx coretypes.Sandbox, target coretypes.AgentID, tokens *ledgerstate.ColoredBalances) error {
	if tokens == nil {
		return nil
	}
	p := codec.MakeDict(map[string]interface{}{
		ParamAgentID: target,
	})
	_, err := ctx.Call(Interface.Hname(), coretypes.Hn(FuncDeposit), p, tokens)
	return err
}

// ChainOwnersAccount the account of chain owner depends on the chain ID
func ChainOwnersAccount(ctx coretypes.Sandbox) *coretypes.AgentID {
	return coretypes.NewAgentIDFromAddress(ctx.ContractID().ChainID().AsAddress())
}

// AccountFromAgentID if agentID represents the owner, the account is owner's account
// Otherwise it is equal to the agentID
func AccountFromAgentID(ctx coretypes.Sandbox, agentID *coretypes.AgentID) *coretypes.AgentID {
	if ctx.ChainOwnerID().Equals(agentID) {
		return ChainOwnersAccount(ctx)
	}
	return agentID
}
