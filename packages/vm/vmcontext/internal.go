package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger
// It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) {
	if len(vmctx.callStack) > 0 {
		vmctx.log.Panicf("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accounts.CreditToAccount(vmctx.State(), agentID, transfer)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accounts.DebitFromAccount(vmctx.State(), agentID, transfer)
}

func (vmctx *VMContext) moveBetweenAccounts(fromAgentID, toAgentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	if len(vmctx.callStack) == 0 {
		vmctx.log.Panicf("moveBetweenAccounts can't be called from request context")
	}

	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accounts.MoveBetweenAccounts(vmctx.State(), fromAgentID, toAgentID, transfer)
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

func (vmctx *VMContext) mustGetChainInfo() root.ChainInfo {
	vmctx.pushCallContext(root.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return root.MustGetChainInfo(vmctx.State())
}

func (vmctx *VMContext) getFeeInfo() (balance.Color, int64, int64) {
	vmctx.pushCallContext(root.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return root.GetFeeInfoByContractRecord(vmctx.State(), vmctx.contractRecord)
}

func (vmctx *VMContext) getBinary(programHash hashing.HashValue) (string, []byte, error) {
	vmtype, ok := processors.GetBuiltinProcessorType(programHash)
	if ok {
		return vmtype, nil, nil
	}
	vmctx.pushCallContext(blob.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blob.LocateProgram(vmctx.State(), programHash)
}

func (vmctx *VMContext) getBalance(col balance.Color) int64 {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetBalance(vmctx.State(), vmctx.MyAgentID(), col)
}

func (vmctx *VMContext) getMyBalances() coretypes.ColoredBalances {
	agentID := vmctx.MyAgentID()

	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret, _ := accounts.GetAccountBalances(vmctx.State(), agentID)
	return cbalances.NewFromMap(ret)
}

func (vmctx *VMContext) moveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.MoveBetweenAccounts(
		vmctx.State(),
		vmctx.MyAgentID(),
		target,
		cbalances.NewFromMap(map[balance.Color]int64{col: amount}),
	)
}

func (vmctx *VMContext) StoreToEventLog(contract coretypes.Hname, data []byte) {
	vmctx.pushCallContext(eventlog.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	vmctx.log.Debugf("StoreToEventLog/%s: data: '%s'", contract.String(), string(data))
	eventlog.AppendToLog(vmctx.State(), vmctx.timestamp, contract, data)
}
