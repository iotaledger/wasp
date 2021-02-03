package accounts

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// Deposit calls "deposit" entry point of the accounts contract.
// Can only be called from full sandbox context
func Deposit(ctx coretypes.Sandbox, target coretypes.AgentID, tokens coretypes.ColoredBalances) error {
	p := codec.MakeDict(map[string]interface{}{
		ParamAgentID: target,
	})
	_, err := ctx.Call(Interface.Hname(), coretypes.Hn(FuncDeposit), p, tokens)
	return err
}
