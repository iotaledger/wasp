package vmcontext

import (
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
	"math"
	"math/big"

	"github.com/iotaledger/wasp/packages/vm/gas"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger. It is used when new tokens arrive with a request
func (vmctx *VMContext) creditToAccount(agentID *iscp.AgentID, assets *iscp.Assets) {
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

func (vmctx *VMContext) totalL2Assets() *iscp.Assets {
	var ret *iscp.Assets
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetTotalL2Assets(s)
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
	vmctx.GasBurn(gas.BurnCodeGetBalance)

	var ret uint64
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetIotaBalance(s, agentID)
	})
	return ret
}

func (vmctx *VMContext) GetNativeTokenBalance(agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	vmctx.GasBurn(gas.BurnCodeGetBalance)

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
	vmctx.GasBurn(gas.BurnCodeGetBalance)

	var ret *iscp.Assets
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetAssets(s, agentID)
		if ret == nil {
			ret = &iscp.Assets{}
		}
	})
	return ret
}

func (vmctx *VMContext) GetSenderTokenBalanceForFees() uint64 {
	sender := vmctx.req.SenderAccount()
	if sender == nil {
		return 0
	}
	if vmctx.chainInfo.GasFeePolicy.GasFeeTokenID == nil {
		// iotas are used as gas tokens
		return vmctx.GetIotaBalance(sender)
	}
	// native tokens are used for gas fee
	tokenID := vmctx.chainInfo.GasFeePolicy.GasFeeTokenID
	// to pay for gas chain is configured to use some native token, not IOTA
	tokensAvailableBig := vmctx.GetNativeTokenBalance(sender, tokenID)
	if tokensAvailableBig.IsUint64() {
		return tokensAvailableBig.Uint64()
	}
	return math.MaxUint64
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

func (vmctx *VMContext) writeReceiptToBlockLog(errProvided error) *blocklog.RequestReceipt {
	errStr := ""
	if errProvided != nil {
		errStr = errProvided.Error()
	}
	receipt := &blocklog.RequestReceipt{
		Request:       vmctx.req,
		ErrorStr:      errStr,
		GasBudget:     vmctx.gasBudgetAdjusted,
		GasBurned:     vmctx.gasBurned,
		GasFeeCharged: vmctx.gasFeeCharged,
	}
	receipt.GasBurnLog = vmctx.gasBurnLog
	if vmctx.task.EnableGasBurnLogging {
		vmctx.gasBurnLog = gas.NewGasBurnLog()
	}
	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		err = blocklog.SaveRequestReceipt(vmctx.State(), receipt, vmctx.requestLookupKey())
	})
	if err != nil {
		panic(err)
	}
	return receipt
}

func (vmctx *VMContext) MustSaveEvent(contract iscp.Hname, msg string) {
	vmctx.GasBurn(gas.BurnCodeEmitEventFixed)

	if vmctx.requestEventIndex > vmctx.chainInfo.MaxEventsPerReq {
		panic(ErrTooManyEvents)
	}
	if len([]byte(msg)) > int(vmctx.chainInfo.MaxEventSize) {
		panic(ErrTooLargeEvent)
	}
	vmctx.Debugf("MustSaveEvent/%s: msg: '%s'", contract.String(), msg)

	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		err = blocklog.SaveEvent(vmctx.State(), msg, vmctx.eventLookupKey(), contract)
	})
	if err != nil {
		panic(err)
	}
	vmctx.requestEventIndex++
}

// updateOffLedgerRequestMaxAssumedNonce updates stored nonce for off ledger requests
func (vmctx *VMContext) updateOffLedgerRequestMaxAssumedNonce() {
	vmctx.gasBurnEnable(false)
	defer vmctx.gasBurnEnable(true)
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.SaveMaxAssumedNonce(
			s,
			vmctx.req.SenderAddress(),
			vmctx.req.AsOffLedger().Nonce(),
		)
	})
}

// adjustL2IotasIfNeeded adjust L2 ledger for iotas if the L1 changed because of dust deposit changes
func (vmctx *VMContext) adjustL2IotasIfNeeded(adjustment int64, account *iscp.AgentID) {
	err := util.CatchPanicReturnError(func() {
		vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
			accounts.AdjustAccountIotas(s, account, adjustment)
		})
	}, accounts.ErrNotEnoughFunds)
	if err != nil {
		panic(vmexceptions.ErrNotEnoughFundsForInternalDustDeposit)
	}
}
