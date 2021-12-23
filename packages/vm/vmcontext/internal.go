package vmcontext

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/kv"

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
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		blocklog.SetAnchorTransactionIDOfLatestBlock(s, vmctx.task.AnchorOutputID.TransactionID)
	})
}

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger. It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID *iscp.AgentID, assets *iscp.Assets) {
	if len(vmctx.callStack) > 0 {
		panic("creditToAccount must be called only from request")
	}
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.CreditToAccount(s, agentID, assets)
	})
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *VMContext) debitFromAccount(agentID *iscp.AgentID, transfer *iscp.Assets) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.DebitFromAccount(s, agentID, transfer)
	})
}

func (vmctx *VMContext) mustMoveBetweenAccounts(fromAgentID, toAgentID *iscp.AgentID, transfer *iscp.Assets) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.MustMoveBetweenAccounts(s, fromAgentID, toAgentID, transfer)
	})
}

func (vmctx *VMContext) totalAssets() *iscp.Assets {
	var ret *iscp.Assets
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetTotalAssets(s)
	})
	return ret
}

func (vmctx *VMContext) findContractByHname(contractHname iscp.Hname) *root.ContractRecord {
	if contractHname == root.Contract.Hname() && vmctx.isInitChainRequest() {
		return root.NewContractRecord(root.Contract, &iscp.NilAgentID)
	}

	var ret *root.ContractRecord
	vmctx.callCore(root.Contract, func(s kv.KVStore) {
		ret = root.FindContract(s, contractHname)
	})
	return ret
}

func (vmctx *VMContext) getChainInfo() *governance.ChainInfo {
	var ret *governance.ChainInfo
	vmctx.callCore(governance.Contract, func(s kv.KVStore) {
		ret = governance.MustGetChainInfo(s)
	})
	return ret
}

func (vmctx *VMContext) GetIotaBalance(agentID *iscp.AgentID) uint64 {
	var ret uint64
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetIotaBalance(s, agentID)
	})
	return ret
}

func (vmctx *VMContext) GetNativeTokenBalance(agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokenBalance(s, agentID, tokenID)
	})
	return ret
}

func (vmctx *VMContext) GetNativeTokenBalanceTotal(tokenID *iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokenBalanceTotal(s, tokenID)
	})
	return ret
}

func (vmctx *VMContext) GetAssets(agentID *iscp.AgentID) *iscp.Assets {
	var ret *iscp.Assets
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetAssets(s, agentID)
		if ret == nil {
			ret = &iscp.Assets{}
		}
	})
	return ret
}

func (vmctx *VMContext) getBinary(programHash hashing.HashValue) (string, []byte, error) {
	vmtype, ok := vmctx.task.Processors.Config.GetNativeProcessorType(programHash)
	if ok {
		return vmtype, nil, nil
	}
	var binary []byte
	var err error
	vmctx.callCore(blob.Contract, func(s kv.KVStore) {
		vmtype, binary, err = blob.LocateProgram(vmctx.State(), programHash)

	})
	return vmtype, binary, err
}

func (vmctx *VMContext) requestLookupKey() blocklog.RequestLookupKey {
	return blocklog.NewRequestLookupKey(vmctx.virtualState.BlockIndex(), vmctx.requestIndex)
}

func (vmctx *VMContext) eventLookupKey() blocklog.EventLookupKey {
	return blocklog.NewEventLookupKey(vmctx.virtualState.BlockIndex(), vmctx.requestIndex, vmctx.requestEventIndex)
}

func (vmctx *VMContext) logRequestToBlockLog(errProvided error) {
	errStr := ""
	if errProvided != nil {
		errStr = errProvided.Error()
	}
	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		err = blocklog.SaveRequestLogRecord(vmctx.State(), &blocklog.RequestReceipt{
			RequestData: vmctx.req,
			Error:       errStr,
		}, vmctx.requestLookupKey())

	})
	if err != nil {
		vmctx.Panicf("logRequestToBlockLog: %v", err)
	}
}

func (vmctx *VMContext) MustSaveEvent(contract iscp.Hname, msg string) {
	if vmctx.requestEventIndex > vmctx.chainInfo.MaxEventsPerReq {
		vmctx.Panicf("too many events issued for contract: %s, request index: %d", contract.String(), vmctx.requestIndex)
	}
	if len([]byte(msg)) > int(vmctx.chainInfo.MaxEventSize) {
		vmctx.Panicf("event too large: %s, request index: %d", contract.String(), vmctx.requestIndex)
	}
	vmctx.Debugf("MustSaveEvent/%s: msg: '%s'", contract.String(), msg)

	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		err = blocklog.SaveEvent(vmctx.State(), msg, vmctx.eventLookupKey(), contract)
	})
	if err != nil {
		vmctx.Panicf("MustSaveEvent: %v", err)
	}
	vmctx.requestEventIndex++
}

// updateOffLedgerRequestMaxAssumedNonce updates stored nonce for off ledger requests
func (vmctx *VMContext) updateOffLedgerRequestMaxAssumedNonce() {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.SaveMaxAssumedNonce(
			s,
			vmctx.req.SenderAddress(),
			vmctx.req.Unwrap().OffLedger().Nonce(),
		)
	})
}
