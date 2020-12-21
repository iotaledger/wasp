package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/hardcoded"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger
// It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) {
	if len(vmctx.callStack) > 0 {
		vmctx.log.Panicf("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accountsc.CreditToAccount(vmctx.State(), agentID, transfer)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.DebitFromAccount(vmctx.State(), agentID, transfer)
}

func (vmctx *VMContext) moveBetweenAccounts(fromAgentID, toAgentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	if len(vmctx.callStack) == 0 {
		vmctx.log.Panicf("moveBetweenAccounts can't be called from request context")
	}

	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accountsc.MoveBetweenAccounts(vmctx.State(), fromAgentID, toAgentID, transfer)
}

func (vmctx *VMContext) findContractByHname(contractHname coretypes.Hname) (*root.ContractRecord, bool) {
	vmctx.pushCallContext(root.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret, err := root.FindContract(vmctx.State(), contractHname)
	if err != nil {
		return nil, false
	}
	return ret, true
}

func (vmctx *VMContext) getChainInfo() *root.ChainInfo {
	vmctx.pushCallContext(root.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return root.GetChainInfo(vmctx.State())
}

func (vmctx *VMContext) getFeeInfo(contractHname coretypes.Hname) (balance.Color, int64, int64, bool) {
	vmctx.pushCallContext(root.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	col, ownerFee, validatorFee, err := root.GetFeeInfo(vmctx.State(), contractHname)
	if err != nil {
		return balance.Color{}, 0, 0, false
	}
	return col, ownerFee, validatorFee, true
}

func (vmctx *VMContext) getBinary(programHash hashing.HashValue) (string, []byte, error) {
	vmtype, ok := hardcoded.LocateHardcodedProgram(programHash)
	if ok {
		return vmtype, nil, nil
	}
	vmctx.pushCallContext(blob.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blob.LocateProgram(vmctx.State(), programHash)
}

func (vmctx *VMContext) getBalance(col balance.Color) int64 {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accountsc.GetBalance(vmctx.State(), vmctx.MyAgentID(), col)
}

func (vmctx *VMContext) getMyBalances() coretypes.ColoredBalances {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret, _ := accountsc.GetAccountBalances(vmctx.State(), vmctx.MyAgentID())
	return cbalances.NewFromMap(ret)
}

func (vmctx *VMContext) moveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool {
	vmctx.pushCallContext(accountsc.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accountsc.MoveBetweenAccounts(
		vmctx.State(),
		vmctx.MyAgentID(),
		target,
		cbalances.NewFromMap(map[balance.Color]int64{col: amount}),
	)
}

func (vmctx *VMContext) StoreToChainLog(contract coretypes.Hname, data []byte) {
	vmctx.pushCallContext(chainlog.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	vmctx.log.Debugf("StoreToChainLog/%s: data: '%s'", contract.String(), string(data))
	chainlog.AppendToChainLog(vmctx.State(), vmctx.timestamp, contract, data)
}
