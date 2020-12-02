package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/coret/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/hardcoded"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger
// It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID coret.AgentID, transfer coret.ColoredBalances) {
	if len(vmctx.callStack) > 0 {
		vmctx.log.Panicf("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accountsc.CreditToAccount(codec.NewMustCodec(vmctx), agentID, transfer)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID coret.AgentID, transfer coret.ColoredBalances) bool {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.DebitFromAccount(codec.NewMustCodec(vmctx), agentID, transfer)
}

func (vmctx *VMContext) moveBetweenAccounts(fromAgentID, toAgentID coret.AgentID, transfer coret.ColoredBalances) bool {
	if len(vmctx.callStack) == 0 {
		vmctx.log.Panicf("moveBetweenAccounts can't be called from request context")
	}

	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.MoveBetweenAccounts(codec.NewMustCodec(vmctx), fromAgentID, toAgentID, transfer)
}

func (vmctx *VMContext) findContractByHname(contractHname coret.Hname) (*root.ContractRecord, bool) {
	vmctx.pushCallContext(root.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret, err := root.FindContract(codec.NewMustCodec(vmctx), contractHname)
	if err != nil {
		return nil, false
	}
	return ret, true
}

func (vmctx *VMContext) getBinary(programHash hashing.HashValue) (string, []byte, error) {
	vmtype, ok := hardcoded.LocateHardcodedProgram(programHash)
	if ok {
		return vmtype, nil, nil
	}
	vmctx.pushCallContext(blob.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blob.LocateProgram(codec.NewMustCodec(vmctx), programHash)
}

func (vmctx *VMContext) getBalance(col balance.Color) int64 {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accountsc.GetBalance(codec.NewMustCodec(vmctx), vmctx.MyAgentID(), col)
}

func (vmctx *VMContext) getMyBalances() coret.ColoredBalances {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret, _ := accountsc.GetAccountBalances(codec.NewMustCodec(vmctx), vmctx.MyAgentID())
	return cbalances.NewFromMap(ret)
}

func (vmctx *VMContext) moveBalance(target coret.AgentID, col balance.Color, amount int64) bool {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accountsc.MoveBetweenAccounts(
		codec.NewMustCodec(vmctx),
		vmctx.MyAgentID(),
		target,
		cbalances.NewFromMap(map[balance.Color]int64{col: amount}),
	)
}
