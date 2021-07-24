package vmcontext

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger
// It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID *iscp.AgentID, transfer colored.Balances) {
	if len(vmctx.callStack) > 0 {
		vmctx.log.Panicf("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accounts.CreditToAccount(vmctx.State(), agentID, transfer)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID *iscp.AgentID, transfer colored.Balances) bool {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accounts.DebitFromAccount(vmctx.State(), agentID, transfer)
}

func (vmctx *VMContext) moveBetweenAccounts(fromAgentID, toAgentID *iscp.AgentID, transfer colored.Balances) bool {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accounts.MoveBetweenAccounts(vmctx.State(), fromAgentID, toAgentID, transfer)
}

func (vmctx *VMContext) totalAssets() colored.Balances {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetTotalAssets(vmctx.State())
}

func (vmctx *VMContext) findContractByHname(contractHname iscp.Hname) (*root.ContractRecord, bool) {
	vmctx.pushCallContext(root.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret, err := root.FindContract(vmctx.State(), contractHname)
	if err != nil {
		return nil, false
	}
	return ret, true
}

func (vmctx *VMContext) mustGetChainInfo() root.ChainInfo {
	vmctx.pushCallContext(root.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return root.MustGetChainInfo(vmctx.State())
}

func (vmctx *VMContext) getFeeInfo() (colored.Color, uint64, uint64) {
	vmctx.pushCallContext(root.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return root.GetFeeInfoByContractRecord(vmctx.State(), vmctx.contractRecord)
}

func (vmctx *VMContext) getBinary(programHash hashing.HashValue) (string, []byte, error) {
	vmtype, ok := vmctx.processors.Config.GetNativeProcessorType(programHash)
	if ok {
		return vmtype, nil, nil
	}
	vmctx.pushCallContext(blob.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blob.LocateProgram(vmctx.State(), programHash)
}

func (vmctx *VMContext) getBalanceOfAccount(agentID *iscp.AgentID, col colored.Color) uint64 {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetBalance(vmctx.State(), agentID, col)
}

func (vmctx *VMContext) getBalance(col colored.Color) uint64 {
	return vmctx.getBalanceOfAccount(vmctx.MyAgentID(), col)
}

func (vmctx *VMContext) getMyBalances() colored.Balances {
	agentID := vmctx.MyAgentID()

	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret, _ := accounts.GetAccountBalances(vmctx.State(), agentID)
	return ret
}

//nolint:unused
func (vmctx *VMContext) moveBalance(target iscp.AgentID, col colored.Color, amount uint64) bool {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.MoveBetweenAccounts(
		vmctx.State(),
		vmctx.MyAgentID(),
		&target,
		colored.NewBalancesForColor(col, amount),
	)
}

func (vmctx *VMContext) requestLookupKey() blocklog.RequestLookupKey {
	return blocklog.NewRequestLookupKey(vmctx.virtualState.BlockIndex(), vmctx.requestIndex)
}

func (vmctx *VMContext) mustLogRequestToBlockLog(errProvided error) {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	errStr := ""
	if errProvided != nil {
		errStr = errProvided.Error()
	}
	err := blocklog.SaveRequestLogRecord(vmctx.State(), &blocklog.RequestReceipt{
		RequestID: vmctx.req.ID(),
		OffLedger: vmctx.req.Output() == nil,
		Error:     errStr,
	}, vmctx.requestLookupKey())
	if err != nil {
		vmctx.Panicf("logRequestToBlockLog: %v", err)
	}
}

func (vmctx *VMContext) StoreToEventLog(contract iscp.Hname, data []byte) {
	vmctx.pushCallContext(eventlog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	vmctx.log.Debugf("StoreToEventLog/%s: data: '%s'", contract.String(), string(data))
	eventlog.AppendToLog(vmctx.State(), vmctx.virtualState.Timestamp().UnixNano(), contract, data)
}
