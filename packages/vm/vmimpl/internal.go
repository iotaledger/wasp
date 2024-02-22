package vmimpl

import (
	"math"
	"math/big"

	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// creditToAccount credits assets to the chain ledger
func (reqctx *requestContext) creditToAccount(agentID isc.AgentID, ftokens *isc.FungibleTokens) {
	reqctx.vm.withAccountsState(reqctx.uncommittedState, func(s *accounts.StateWriter) {
		s.CreditToAccount(agentID, ftokens)
	})
}

// creditToAccountFullDecimals credits assets to the chain ledger
func (reqctx *requestContext) creditToAccountFullDecimals(agentID isc.AgentID, amount *big.Int, gasBurn bool) {
	reqctx.vm.withAccountsState(reqctx.chainState(gasBurn), func(s *accounts.StateWriter) {
		s.CreditToAccountFullDecimals(agentID, amount, nil)
	})
}

func (reqctx *requestContext) creditNFTToAccount(agentID isc.AgentID) {
	req := reqctx.req.(isc.OnLedgerRequest)
	nft := req.NFT()
	if nft == nil {
		return
	}
	o := req.Output()
	nftOutput := o.(*iotago.NFTOutput)
	if nftOutput.NFTID.Empty() {
		nftOutput.NFTID = util.NFTIDFromNFTOutput(nftOutput, req.OutputID()) // handle NFTs that were minted diractly to the chain
	}
	reqctx.vm.withAccountsState(reqctx.uncommittedState, func(s *accounts.StateWriter) {
		s.CreditNFTToAccount(agentID, nftOutput)
	})
}

// debitFromAccount subtracts tokens from account if there are enough.
func (reqctx *requestContext) debitFromAccount(agentID isc.AgentID, transfer *isc.FungibleTokens, gasBurn bool) {
	reqctx.vm.withAccountsState(reqctx.chainState(gasBurn), func(s *accounts.StateWriter) {
		s.DebitFromAccount(agentID, transfer)
	})
}

// debitFromAccountFullDecimals subtracts basetokens tokens from account if there are enough.
func (reqctx *requestContext) debitFromAccountFullDecimals(agentID isc.AgentID, amount *big.Int, gasBurn bool) {
	reqctx.vm.withAccountsState(reqctx.chainState(gasBurn), func(s *accounts.StateWriter) {
		s.DebitFromAccountFullDecimals(agentID, amount, nil)
	})
}

// debitNFTFromAccount removes a NFT from an account.
func (reqctx *requestContext) debitNFTFromAccount(agentID isc.AgentID, nftID iotago.NFTID, gasBurn bool) {
	reqctx.vm.withAccountsState(reqctx.chainState(gasBurn), func(s *accounts.StateWriter) {
		s.DebitNFTFromAccount(agentID, nftID)
	})
}

func (reqctx *requestContext) mustMoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets, gasBurn bool) {
	reqctx.vm.withAccountsState(reqctx.chainState(gasBurn), func(s *accounts.StateWriter) {
		lo.Must0(s.MoveBetweenAccounts(fromAgentID, toAgentID, assets))
	})
}

func findContractByHname(chainState kv.KVStore, contractHname isc.Hname) (ret *root.ContractRecord) {
	withContractState(chainState, root.Contract, func(s kv.KVStore) {
		ret = root.FindContract(s, contractHname)
	})
	return ret
}

func (reqctx *requestContext) GetBaseTokensBalance(agentID isc.AgentID) (bts iotago.BaseToken, remainder *big.Int) {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		bts, remainder = s.GetBaseTokensBalance(agentID)
	})
	return
}

func (reqctx *requestContext) GetBaseTokensBalanceDiscardExtraDecimals(agentID isc.AgentID) (bts iotago.BaseToken) {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		bts, _ = s.GetBaseTokensBalance(agentID)
	})
	return
}

func (reqctx *requestContext) HasEnoughForAllowance(agentID isc.AgentID, allowance *isc.Assets) bool {
	var ret bool
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.HasEnoughForAllowance(agentID, allowance)
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNativeTokenBalance(agentID, nativeTokenID)
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokenBalanceTotal(nativeTokenID iotago.NativeTokenID) *big.Int {
	var ret *big.Int
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNativeTokenBalanceTotal(nativeTokenID)
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokens(agentID isc.AgentID) iotago.NativeTokenSum {
	var ret iotago.NativeTokenSum
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNativeTokens(agentID)
	})
	return ret
}

func (reqctx *requestContext) GetAccountNFTs(agentID isc.AgentID) (ret []iotago.NFTID) {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetAccountNFTs(agentID)
	})
	return ret
}

func (reqctx *requestContext) GetNFTData(nftID iotago.NFTID) (ret *isc.NFT) {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNFTData(nftID)
	})
	return ret
}

func (reqctx *requestContext) GetSenderTokenBalanceForFees() iotago.BaseToken {
	sender := reqctx.req.SenderAccount()
	if sender == nil {
		return 0
	}
	return reqctx.GetBaseTokensBalanceDiscardExtraDecimals(sender)
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

	reqctx.LogDebugf("writeReceiptToBlockLog - reqID:%s err: %v", reqctx.req.ID(), vmError)

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
			f.callback(s, receipt.GasBurned)
		})
	}
	return receipt
}

func (vmctx *vmContext) storeUnprocessable(chainState kv.KVStore, unprocessable []isc.OnLedgerRequest, lastInternalAssetUTXOIndex uint16) {
	if len(unprocessable) == 0 {
		return
	}
	blockIndex := vmctx.task.Inputs.AnchorOutput.StateIndex + 1

	withContractState(chainState, blocklog.Contract, func(s kv.KVStore) {
		for _, r := range unprocessable {
			if r.SenderAccount() == nil {
				continue
			}
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
	reqctx.LogDebugf("MustSaveEvent/%s: topic: '%s'", hContract.String(), topic)

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
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		s.IncrementNonce(reqctx.req.SenderAccount())
	})
}

// adjustL2BaseTokensIfNeeded adjust L2 ledger for base tokens if the L1 changed because of storage deposit changes
func (reqctx *requestContext) adjustL2BaseTokensIfNeeded(adjustment int64, account isc.AgentID) {
	if adjustment == 0 {
		return
	}
	err := panicutil.CatchPanicReturnError(func() {
		reqctx.callAccounts(func(s *accounts.StateWriter) {
			s.AdjustAccountBaseTokens(account, adjustment)
		})
	}, accounts.ErrNotEnoughFunds)
	if err != nil {
		panic(vmexceptions.ErrNotEnoughFundsForInternalStorageDeposit)
	}
}
