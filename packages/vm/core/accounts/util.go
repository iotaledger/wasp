package accounts

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// Accrue calls "deposit" entry point of the accounts contract.
// Can only be called from full sandbox context
func Accrue(ctx coretypes.Sandbox, target coretypes.AgentID, tokens coretypes.ColoredBalances) error {
	if tokens == nil || tokens.Len() == 0 {
		return nil
	}
	p := codec.MakeDict(map[string]interface{}{
		ParamAgentID: target,
	})
	_, err := ctx.Call(Interface.Hname(), coretypes.Hn(FuncDeposit), p, tokens)
	return err
}
