package vmcontext

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) updateLatestAnchorID() {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	blocklog.SetAnchorTransactionIDOfLatestBlock(vmctx.State(), vmctx.task.AnchorOutputID.TransactionID)
}

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger. It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID *iscp.AgentID, assets *iscp.Assets) {
	if len(vmctx.callStack) > 0 {
		panic("creditToAccount must be called only from request")
	}
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accounts.CreditToAccount(vmctx.State(), agentID, assets)
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID *iscp.AgentID, transfer *iscp.Assets) {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	accounts.DebitFromAccount(vmctx.State(), agentID, transfer)
}

func (vmctx *VMContext) mustMoveBetweenAccounts(fromAgentID, toAgentID *iscp.AgentID, transfer *iscp.Assets) {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil) // create local context for the state
	defer vmctx.popCallContext()

	accounts.MustMoveBetweenAccounts(vmctx.State(), fromAgentID, toAgentID, transfer)
}

func (vmctx *VMContext) totalAssets() *iscp.Assets {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetTotalAssets(vmctx.State())
}

func (vmctx *VMContext) findContractByHname(contractHname iscp.Hname) *root.ContractRecord {
	vmctx.pushCallContext(root.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	if contractHname == root.Contract.Hname() && vmctx.isInitChainRequest() {
		return root.NewContractRecord(root.Contract, &iscp.NilAgentID)
	}
	return root.FindContract(vmctx.State(), contractHname)
}

func (vmctx *VMContext) getChainInfo() *governance.ChainInfo {
	vmctx.pushCallContext(governance.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return governance.MustGetChainInfo(vmctx.State())
}

func (vmctx *VMContext) GetIotaBalance(agentID *iscp.AgentID) uint64 {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetIotaBalance(vmctx.State(), agentID)
}

func (vmctx *VMContext) GetNativeTokenBalance(agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetNativeTokenBalance(vmctx.State(), agentID, tokenID)
}

func (vmctx *VMContext) GetNativeTokenBalanceTotal(tokenID *iotago.NativeTokenID) *big.Int {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetNativeTokenBalanceTotal(vmctx.State(), tokenID)
}

func (vmctx *VMContext) GetAssets(agentID *iscp.AgentID) *iscp.Assets {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	ret := accounts.GetAssets(vmctx.State(), agentID)
	if ret == nil {
		ret = &iscp.Assets{}
	}
	return ret
}

func (vmctx *VMContext) getBinary(programHash hashing.HashValue) (string, []byte, error) {
	vmtype, ok := vmctx.task.Processors.Config.GetNativeProcessorType(programHash)
	if ok {
		return vmtype, nil, nil
	}
	vmctx.pushCallContext(blob.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blob.LocateProgram(vmctx.State(), programHash)
}

func (vmctx *VMContext) requestLookupKey() blocklog.RequestLookupKey {
	return blocklog.NewRequestLookupKey(vmctx.virtualState.BlockIndex(), vmctx.requestIndex)
}

func (vmctx *VMContext) eventLookupKey() blocklog.EventLookupKey {
	return blocklog.NewEventLookupKey(vmctx.virtualState.BlockIndex(), vmctx.requestIndex, vmctx.requestEventIndex)
}

func (vmctx *VMContext) logRequestToBlockLog(errProvided error) {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	errStr := ""
	if errProvided != nil {
		errStr = errProvided.Error()
	}
	err := blocklog.SaveRequestLogRecord(vmctx.State(), &blocklog.RequestReceipt{
		RequestData: vmctx.req,
		Error:       errStr,
	}, vmctx.requestLookupKey())
	if err != nil {
		vmctx.Panicf("logRequestToBlockLog: %v", err)
	}
}

func (vmctx *VMContext) MustSaveEvent(contract iscp.Hname, msg string) {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	if vmctx.requestEventIndex > vmctx.chainInfo.MaxEventsPerReq {
		vmctx.Panicf("too many events issued for contract: %s, request index: %d", contract.String(), vmctx.requestIndex)
	}

	if len([]byte(msg)) > int(vmctx.chainInfo.MaxEventSize) {
		vmctx.Panicf("event too large: %s, request index: %d", contract.String(), vmctx.requestIndex)
	}

	vmctx.Debugf("MustSaveEvent/%s: msg: '%s'", contract.String(), msg)
	err := blocklog.SaveEvent(vmctx.State(), msg, vmctx.eventLookupKey(), contract)
	if err != nil {
		vmctx.Panicf("MustSaveEvent: %v", err)
	}
	vmctx.requestEventIndex++
}

// updateOffLedgerRequestMaxAssumedNonce updates stored nonce for off ledger requests
func (vmctx *VMContext) updateOffLedgerRequestMaxAssumedNonce() {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	accounts.SaveMaxAssumedNonce(
		vmctx.State(),
		vmctx.req.SenderAddress(),
		vmctx.req.Unwrap().OffLedger().Nonce(),
	)
}
