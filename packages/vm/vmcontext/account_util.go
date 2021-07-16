package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// Accrue calls "deposit" entry point of the accounts contract.
// Can only be called from full sandbox context
func Accrue(ctx iscp.Sandbox, target *iscp.AgentID, tokens *ledgerstate.ColoredBalances) error {
	if tokens == nil {
		return nil
	}
	p := codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: target,
	})
	_, err := ctx.Call(accounts.Interface.Hname(), accounts.FuncDeposit.Hname(), p, tokens)
	return err
}
