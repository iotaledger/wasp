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
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// creditToAccount deposits transfer from request to chain account of of the called contract
// It adds new tokens to the chain ledger. It is used when new tokens arrive with a request
func creditToAccount(chainState kv.KVStore, agentID isc.AgentID, ftokens *isc.Assets) {
	withContractState(chainState, accounts.Contract, func(s kv.KVStore) {
		accounts.CreditToAccount(s, agentID, ftokens)
	})
}

func creditNFTToAccount(chainState kv.KVStore, agentID isc.AgentID, nft *isc.NFT) {
	withContractState(chainState, accounts.Contract, func(s kv.KVStore) {
		accounts.CreditNFTToAccount(s, agentID, nft)
	})
}

// debitFromAccount subtracts tokens from account if it is enough of it.
// should be called only when posting request
func debitFromAccount(chainState kv.KVStore, agentID isc.AgentID, transfer *isc.Assets) {
	withContractState(chainState, accounts.Contract, func(s kv.KVStore) {
		accounts.DebitFromAccount(s, agentID, transfer)
	})
}

// debitNFTFromAccount removes a NFT from account.
// should be called only when posting request
func debitNFTFromAccount(chainState kv.KVStore, agentID isc.AgentID, nftID iotago.NFTID) {
	withContractState(chainState, accounts.Contract, func(s kv.KVStore) {
		accounts.DebitNFTFromAccount(s, agentID, nftID)
	})
}

func mustMoveBetweenAccounts(chainState kv.KVStore, fromAgentID, toAgentID isc.AgentID, assets *isc.Assets) {
	withContractState(chainState, accounts.Contract, func(s kv.KVStore) {
		accounts.MustMoveBetweenAccounts(s, fromAgentID, toAgentID, assets)
	})
}

func findContractByHname(chainState kv.KVStore, contractHname isc.Hname) (ret *root.ContractRecord) {
	withContractState(chainState, root.Contract, func(s kv.KVStore) {
		ret = root.FindContract(s, contractHname)
	})
	return ret
}

func (reqctx *requestContext) GetBaseTokensBalance(agentID isc.AgentID) uint64 {
	var ret uint64
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetBaseTokensBalance(s, agentID)
	})
	return ret
}

func (reqctx *requestContext) HasEnoughForAllowance(agentID isc.AgentID, allowance *isc.Assets) bool {
	var ret bool
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.HasEnoughForAllowance(s, agentID, allowance)
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokenBalance(s, agentID, nativeTokenID)
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokenBalanceTotal(nativeTokenID iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokenBalanceTotal(s, nativeTokenID)
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokens(agentID isc.AgentID) iotago.NativeTokens {
	var ret iotago.NativeTokens
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetNativeTokens(s, agentID)
	})
	return ret
}

func (reqctx *requestContext) GetAccountNFTs(agentID isc.AgentID) (ret []iotago.NFTID) {
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.GetAccountNFTs(s, agentID)
	})
	return ret
}

func (reqctx *requestContext) GetNFTData(nftID iotago.NFTID) (ret *isc.NFT) {
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		ret = accounts.MustGetNFTData(s, nftID)
	})
	return ret
}

func (reqctx *requestContext) GetSenderTokenBalanceForFees() uint64 {
	sender := reqctx.req.SenderAccount()
	if sender == nil {
		return 0
	}
	return reqctx.GetBaseTokensBalance(sender)
}

func (reqctx *requestContext) requestLookupKey() blocklog.RequestLookupKey {
	return blocklog.NewRequestLookupKey(reqctx.vm.stateDraft.BlockIndex(), reqctx.requestIndex)
}

func (reqctx *requestContext) eventLookupKey() blocklog.EventLookupKey {
	return blocklog.NewEventLookupKey(reqctx.vm.stateDraft.BlockIndex(), reqctx.requestIndex, reqctx.requestEventIndex)
}

func (reqctx *requestContext) writeReceiptToBlockLog(vmError *isc.VMError) *blocklog.RequestReceipt {
	receipt := &blocklog.RequestReceipt{
		Request:       reqctx.req,
		GasBudget:     reqctx.gas.budgetAdjusted,
		GasBurned:     reqctx.gas.burned,
		GasFeeCharged: reqctx.gas.feeCharged,
		GasBurnLog:    reqctx.gas.burnLog,
		SDCharged:     reqctx.sdCharged,
	}

	if vmError != nil {
		b := vmError.Bytes()
		if len(b) > isc.VMErrorMessageLimit {
			vmError = coreerrors.ErrErrorMessageTooLong
		}
		receipt.Error = vmError.AsUnresolvedError()
	}

	reqctx.Debugf("writeReceiptToBlockLog - reqID:%s err: %v", reqctx.req.ID(), vmError)

	key := reqctx.requestLookupKey()
	var err error
	reqctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		err = blocklog.SaveRequestReceipt(s, receipt, key)
	})
	if err != nil {
		panic(err)
	}
	for _, f := range reqctx.onWriteReceipt {
		reqctx.callCore(corecontracts.All[f.contract], func(s kv.KVStore) {
			f.callback(s)
		})
	}
	return receipt
}

func (vmctx *vmContext) storeUnprocessable(chainState kv.KVStore, unprocessable []isc.OnLedgerRequest, lastInternalAssetUTXOIndex uint16) {
	if len(unprocessable) == 0 {
		return
	}
	blockIndex := vmctx.task.AnchorOutput.StateIndex + 1

	withContractState(chainState, blocklog.Contract, func(s kv.KVStore) {
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

func (reqctx *requestContext) mustSaveEvent(hContract isc.Hname, topic string, payload []byte) {
	if reqctx.requestEventIndex == math.MaxUint16 {
		panic(vm.ErrTooManyEvents)
	}
	reqctx.Debugf("MustSaveEvent/%s: topic: '%s'", hContract.String(), topic)

	event := &isc.Event{
		ContractID: hContract,
		Topic:      topic,
		Payload:    payload,
		Timestamp:  uint64(reqctx.Timestamp().UnixNano()),
	}
	eventKey := reqctx.eventLookupKey().Bytes()
	reqctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		blocklog.SaveEvent(s, eventKey, event)
	})
	reqctx.requestEventIndex++
}

// updateOffLedgerRequestNonce updates stored nonce for ISC off ledger requests
func (reqctx *requestContext) updateOffLedgerRequestNonce() {
	reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
		accounts.IncrementNonce(s, reqctx.req.SenderAccount())
	})
}

// adjustL2BaseTokensIfNeeded adjust L2 ledger for base tokens if the L1 changed because of storage deposit changes
func (reqctx *requestContext) adjustL2BaseTokensIfNeeded(adjustment int64, account isc.AgentID) {
	if adjustment == 0 {
		return
	}
	err := panicutil.CatchPanicReturnError(func() {
		reqctx.callCore(accounts.Contract, func(s kv.KVStore) {
			accounts.AdjustAccountBaseTokens(s, account, adjustment)
		})
	}, accounts.ErrNotEnoughFunds)
	if err != nil {
		panic(vmexceptions.ErrNotEnoughFundsForInternalStorageDeposit)
	}
}
