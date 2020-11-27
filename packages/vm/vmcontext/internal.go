package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger
// It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) {
	if len(vmctx.callStack) > 0 {
		vmctx.log.Panicf("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accountsc.CreditToAccount(codec.NewMustCodec(vmctx), agentID, transfer)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	vmctx.pushCallContext(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.DebitFromAccount(codec.NewMustCodec(vmctx), agentID, transfer)
}

func (vmctx *VMContext) moveBetweenAccounts(fromAgentID, toAgentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	if len(vmctx.callStack) == 0 {
		vmctx.log.Panicf("moveBetweenAccounts can't be called from request context")
	}

	vmctx.pushCallContext(accountsc.Hname, nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.MoveBetweenAccounts(codec.NewMustCodec(vmctx), fromAgentID, toAgentID, transfer)
}

func (vmctx *VMContext) findContractByHname(contractHname coretypes.Hname) (*root.ContractRecord, bool) {
	vmctx.pushCallContext(root.Hname, nil, nil)
	defer vmctx.popCallContext()

	ret, err := root.FindContract(codec.NewMustCodec(vmctx), contractHname)
	if err != nil {
		return nil, false
	}
	return ret, true
}

func (vmctx *VMContext) getBinary(deploymentHash *hashing.HashValue) ([]byte, error) {
	vmctx.pushCallContext(root.Hname, nil, nil)
	defer vmctx.popCallContext()

	return root.GetBinary(codec.NewMustCodec(vmctx), *deploymentHash)
}

func (vmctx *VMContext) getBalance(col balance.Color) int64 {
	vmctx.pushCallContext(root.Hname, nil, nil)
	defer vmctx.popCallContext()

	return accountsc.GetBalance(codec.NewMustCodec(vmctx), vmctx.MyAgentID(), col)
}

func (vmctx *VMContext) getMyBalances() coretypes.ColoredBalances {
	vmctx.pushCallContext(root.Hname, nil, nil)
	defer vmctx.popCallContext()

	ret, _ := accountsc.GetAccountBalances(codec.NewMustCodec(vmctx), vmctx.MyAgentID())
	return cbalances.NewFromMap(ret)
}

func (vmctx *VMContext) moveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool {
	vmctx.pushCallContext(root.Hname, nil, nil)
	defer vmctx.popCallContext()

	return accountsc.MoveBetweenAccounts(
		codec.NewMustCodec(vmctx),
		vmctx.MyAgentID(),
		target,
		cbalances.NewFromMap(map[balance.Color]int64{col: amount}),
	)
}
