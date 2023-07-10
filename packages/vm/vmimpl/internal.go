package vmimpl

import (
	"math"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger. It is used when new tokens arrive with a request
func (vmctx *vmContext) creditToAccount(agentID isc.AgentID, ftokens *isc.Assets) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.CreditToAccount(s, agentID, ftokens)
	})
}

func (vmctx *vmContext) creditNFTToAccount(agentID isc.AgentID, nft *isc.NFT) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.CreditNFTToAccount(s, agentID, nft)
	})
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func (vmctx *vmContext) debitFromAccount(agentID isc.AgentID, transfer *isc.Assets) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.DebitFromAccount(s, agentID, transfer)
	})
}

// debitNFTFromAccount removes a NFT from account.
// should be called only when posting request
func (vmctx *vmContext) debitNFTFromAccount(agentID isc.AgentID, nftID iotago.NFTID) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.DebitNFTFromAccount(s, agentID, nftID)
	})
}

func (vmctx *vmContext) mustMoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.MustMoveBetweenAccounts(s, fromAgentID, toAgentID, assets)
	})
}

func (vmctx *vmContext) findContractByHname(contractHname isc.Hname) (ret *root.ContractRecord) {
	vmctx.callCore(root.Contract, func(s kv.KVStore) {
		ret = root.FindContract(s, contractHname)
	})
	return ret
}

func (vmctx *vmContext) GetBaseTokensBalance(agentID isc.AgentID) uint64 {
	var ret uint64
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetBaseTokensBalance(s, agentID)
	})
	return ret
}

func (vmctx *vmContext) HasEnoughForAllowance(agentID isc.AgentID, allowance *isc.Assets) bool {
	var ret bool
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.HasEnoughForAllowance(s, agentID, allowance)
	})
	return ret
}

func (vmctx *vmContext) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokenBalance(s, agentID, nativeTokenID)
	})
	return ret
}

func (vmctx *vmContext) GetNativeTokenBalanceTotal(nativeTokenID iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokenBalanceTotal(s, nativeTokenID)
	})
	return ret
}

func (vmctx *vmContext) GetNativeTokens(agentID isc.AgentID) iotago.NativeTokens {
	var ret iotago.NativeTokens
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokens(s, agentID)
	})
	return ret
}

func (vmctx *vmContext) GetAccountNFTs(agentID isc.AgentID) (ret []iotago.NFTID) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetAccountNFTs(s, agentID)
	})
	return ret
}

func (vmctx *vmContext) GetNFTData(nftID iotago.NFTID) (ret *isc.NFT) {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.MustGetNFTData(s, nftID)
	})
	return ret
}

func (vmctx *vmContext) GetSenderTokenBalanceForFees() uint64 {
	sender := vmctx.reqctx.req.SenderAccount()
	if sender == nil {
		return 0
	}
	return vmctx.GetBaseTokensBalance(sender)
}

func (vmctx *vmContext) requestLookupKey() blocklog.RequestLookupKey {
	return blocklog.NewRequestLookupKey(vmctx.stateDraft.BlockIndex(), vmctx.reqctx.requestIndex)
}

func (vmctx *vmContext) eventLookupKey() blocklog.EventLookupKey {
	return blocklog.NewEventLookupKey(vmctx.stateDraft.BlockIndex(), vmctx.reqctx.requestIndex, vmctx.reqctx.requestEventIndex)
}

func (vmctx *vmContext) writeReceiptToBlockLog(vmError *isc.VMError) *blocklog.RequestReceipt {
	receipt := &blocklog.RequestReceipt{
		Request:       vmctx.reqctx.req,
		GasBudget:     vmctx.reqctx.gas.budgetAdjusted,
		GasBurned:     vmctx.reqctx.gas.burned,
		GasFeeCharged: vmctx.reqctx.gas.feeCharged,
		GasBurnLog:    vmctx.reqctx.gas.burnLog,
		SDCharged:     vmctx.reqctx.sdCharged,
	}

	if vmError != nil {
		b := vmError.Bytes()
		if len(b) > isc.VMErrorMessageLimit {
			vmError = coreerrors.ErrErrorMessageTooLong
		}
		receipt.Error = vmError.AsUnresolvedError()
	}

	vmctx.Debugf("writeReceiptToBlockLog - reqID:%s err: %v", vmctx.reqctx.req.ID(), vmError)

	key := vmctx.requestLookupKey()
	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		err = blocklog.SaveRequestReceipt(s, receipt, key)
	})
	if err != nil {
		panic(err)
	}
	if vmctx.reqctx.evmFailed != nil {
		// save failed EVM transactions
		vmctx.callCore(evm.Contract, func(s kv.KVStore) {
			evmimpl.AddFailedTx(NewSandbox(vmctx), vmctx.reqctx.evmFailed.tx, vmctx.reqctx.evmFailed.receipt)
		})
	}
	return receipt
}

func (vmctx *vmContext) storeUnprocessable(unprocessable []isc.OnLedgerRequest, lastInternalAssetUTXOIndex uint16) {
	if len(unprocessable) == 0 {
		return
	}
	blockIndex := vmctx.task.AnchorOutput.StateIndex + 1

	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		for _, r := range unprocessable {
			txsnapshot := vmctx.createTxBuilderSnapshot()
			err := panicutil.CatchPanic(func() {
				position := vmctx.txbuilder.ConsumeUnprocessable(r)
				outputIndex := position + int(lastInternalAssetUTXOIndex)
				if blocklog.HasUnprocessable(s, r.ID()) {
					panic("already in unprocessable list")
				}
				// save the unprocessable requests and respective output indices onto the state so they can be retried later
				blocklog.SaveUnprocessable(s, r, blockIndex, uint16(outputIndex))
			})
			if err != nil {
				// protocol exception triggered. Rollback
				vmctx.restoreTxBuilderSnapshot(txsnapshot)
			}
		}
	})
}

func (vmctx *vmContext) MustSaveEvent(hContract isc.Hname, topic string, payload []byte) {
	if vmctx.reqctx.requestEventIndex == math.MaxUint16 {
		panic(vm.ErrTooManyEvents)
	}
	vmctx.Debugf("MustSaveEvent/%s: topic: '%s'", hContract.String(), topic)

	event := &isc.Event{
		ContractID: hContract,
		Topic:      topic,
		Payload:    payload,
		Timestamp:  uint64(vmctx.Timestamp().UnixNano()),
	}
	eventKey := vmctx.eventLookupKey().Bytes()
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		blocklog.SaveEvent(s, eventKey, event)
	})
	vmctx.reqctx.requestEventIndex++
}

// updateOffLedgerRequestNonce updates stored nonce for ISC off ledger requests
func (vmctx *vmContext) updateOffLedgerRequestNonce() {
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.IncrementNonce(s, vmctx.reqctx.req.SenderAccount())
	})
}

// adjustL2BaseTokensIfNeeded adjust L2 ledger for base tokens if the L1 changed because of storage deposit changes
func (vmctx *vmContext) adjustL2BaseTokensIfNeeded(adjustment int64, account isc.AgentID) {
	if adjustment == 0 {
		return
	}
	err := panicutil.CatchPanicReturnError(func() {
		vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
			accounts.AdjustAccountBaseTokens(s, account, adjustment)
		})
	}, accounts.ErrNotEnoughFunds)
	if err != nil {
		panic(vmexceptions.ErrNotEnoughFundsForInternalStorageDeposit)
	}
}
