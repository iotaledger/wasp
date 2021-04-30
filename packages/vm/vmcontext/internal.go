package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger
// It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID *coretypes.AgentID, transfer *ledgerstate.ColoredBalances) {
	if len(vmctx.callStack) > 0 {
		vmctx.log.Panicf("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accounts.CreditToAccount(vmctx.State(), agentID, transfer)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID *coretypes.AgentID, transfer *ledgerstate.ColoredBalances) bool {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accounts.DebitFromAccount(vmctx.State(), agentID, transfer)
}

func (vmctx *VMContext) moveBetweenAccounts(fromAgentID, toAgentID *coretypes.AgentID, transfer *ledgerstate.ColoredBalances) bool {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	return accounts.MoveBetweenAccounts(vmctx.State(), fromAgentID, toAgentID, transfer)
}

func (vmctx *VMContext) totalAssets() *ledgerstate.ColoredBalances {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetTotalAssets(vmctx.State())
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

func (vmctx *VMContext) getFeeInfo() (ledgerstate.Color, uint64, uint64) {
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

func (vmctx *VMContext) getBalance(col ledgerstate.Color) uint64 {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	aid := vmctx.MyAgentID()
	return accounts.GetBalance(vmctx.State(), aid, col)
}

func (vmctx *VMContext) getMyBalances() *ledgerstate.ColoredBalances {
	agentID := vmctx.MyAgentID()

	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	r, _ := accounts.GetAccountBalances(vmctx.State(), agentID)
	ret := ledgerstate.NewColoredBalances(r)
	return ret
}

func (vmctx *VMContext) moveBalance(target coretypes.AgentID, col ledgerstate.Color, amount uint64) bool {
	vmctx.pushCallContext(accounts.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	aid := vmctx.MyAgentID()
	bals := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{col: amount})
	return accounts.MoveBetweenAccounts(vmctx.State(), aid, &target, bals)
}

func (vmctx *VMContext) requestLookupKey() blocklog.RequestLookupKey {
	return blocklog.NewRequestLookupKey(vmctx.virtualState.BlockIndex(), vmctx.requestIndex)
}

func (vmctx *VMContext) mustLogRequestToBlockLog(err error) {
	vmctx.pushCallContext(blocklog.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	var data []byte
	if err != nil {
		data = []byte(fmt.Sprintf("%v", err))
	}
	err1 := blocklog.SaveRequestLogRecord(vmctx.State(), &blocklog.RequestLogRecord{
		RequestID: vmctx.req.ID(),
		OffLedger: vmctx.req.Output() == nil,
		LogData:   data,
	}, vmctx.requestLookupKey())
	if err1 != nil {
		vmctx.Panicf("logRequestToBlockLog: %v", err)
	}
}

func (vmctx *VMContext) StoreToEventLog(contract coretypes.Hname, data []byte) {
	vmctx.pushCallContext(eventlog.Interface.Hname(), nil, nil)
	defer vmctx.popCallContext()

	vmctx.log.Debugf("StoreToEventLog/%s: data: '%s'", contract.String(), string(data))
	eventlog.AppendToLog(vmctx.State(), vmctx.virtualState.Timestamp().UnixNano(), contract, data)
}
