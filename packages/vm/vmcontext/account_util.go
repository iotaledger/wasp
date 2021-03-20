package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// Accrue calls "deposit" entry point of the accounts contract.
// Can only be called from full sandbox context
func Accrue(ctx coretypes.Sandbox, target coretypes.AgentID, tokens *ledgerstate.ColoredBalances) error {
	if tokens == nil {
		return nil
	}
	p := codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: target,
	})
	_, err := ctx.Call(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), p, tokens)
	return err
}

// AdjustAccount makes account of the chain owner and all core contracts equal to (chainID, 0)
func IsAdjustableAccount(chainID coretypes.ChainID, chainOwnerID, agentID *coretypes.AgentID) bool {
	chainAddress := chainID.AsAddress()
	if !agentID.Address().Equals(chainAddress) {
		// no need to adjust
		return false
	}
	hname := agentID.Hname()
	if hname != root.Interface.Hname() &&
		hname != accounts.Interface.Hname() &&
		hname != blob.Interface.Hname() &&
		hname != eventlog.Interface.Hname() &&
		!chainOwnerID.Equals(agentID) {
		return false
	}
	return true
}
