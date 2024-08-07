package vmimpl

import (
	"math"
	"math/big"

	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
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
	"github.com/iotaledger/wasp/sui-go/sui"
)

// creditToAccount credits assets to the chain ledger
func (reqctx *requestContext) creditToAccount(agentID isc.AgentID, ftokens *isc.Assets) {
	reqctx.accountsStateWriter(false).CreditToAccount(agentID, ftokens, reqctx.ChainID())
}

// creditToAccountFullDecimals credits assets to the chain ledger
func (reqctx *requestContext) creditToAccountFullDecimals(agentID isc.AgentID, amount *big.Int, gasBurn bool) {
	reqctx.accountsStateWriter(gasBurn).CreditToAccountFullDecimals(agentID, amount, reqctx.ChainID())
}

func (reqctx *requestContext) creditNFTToAccount(agentID isc.AgentID) {
	req := reqctx.req
	nft := req.NFT()
	if nft == nil {
		return
	}
	o := req.Output()
	nftOutput := o.(*iotago.NFTOutput)
	if nftOutput.NFTID.Empty() {
		nftOutput.NFTID = util.NFTIDFromNFTOutput(nftOutput, req.RequestID()) // handle NFTs that were minted diractly to the chain
	}
	reqctx.accountsStateWriter(false).CreditNFTToAccount(agentID, nftOutput, reqctx.ChainID())
}

// debitFromAccount subtracts tokens from account if there are enough.
func (reqctx *requestContext) debitFromAccount(agentID isc.AgentID, transfer *isc.Assets, gasBurn bool) {
	reqctx.accountsStateWriter(gasBurn).DebitFromAccount(agentID, transfer, reqctx.ChainID())
}

// debitFromAccountFullDecimals subtracts basetokens tokens from account if there are enough.
func (reqctx *requestContext) debitFromAccountFullDecimals(agentID isc.AgentID, amount *big.Int, gasBurn bool) {
	reqctx.accountsStateWriter(gasBurn).DebitFromAccountFullDecimals(agentID, amount, reqctx.ChainID())
}

// debitNFTFromAccount removes a NFT from an account.
func (reqctx *requestContext) debitNFTFromAccount(agentID isc.AgentID, nftID sui.ObjectID, gasBurn bool) {
	reqctx.accountsStateWriter(gasBurn).DebitNFTFromAccount(agentID, nftID, reqctx.ChainID())
}

func (reqctx *requestContext) mustMoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets, gasBurn bool) {
	lo.Must0(reqctx.accountsStateWriter(gasBurn).MoveBetweenAccounts(fromAgentID, toAgentID, assets, reqctx.ChainID()))
}

func findContractByHname(chainState kv.KVStore, contractHname isc.Hname) (ret *root.ContractRecord) {
	return root.NewStateReaderFromChainState(chainState).FindContract(contractHname)
}

func (reqctx *requestContext) GetBaseTokensBalance(agentID isc.AgentID) (bts uint64, remainder *big.Int) {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		bts, remainder = s.GetBaseTokensBalance(agentID, reqctx.ChainID())
	})
	return
}

func (reqctx *requestContext) GetBaseTokensBalanceDiscardRemainder(agentID isc.AgentID) (bts uint64) {
	bal, _ := reqctx.GetBaseTokensBalance(agentID)
	return bal
}

func (reqctx *requestContext) HasEnoughForAllowance(agentID isc.AgentID, allowance *isc.Assets) bool {
	var ret bool
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.HasEnoughForAllowance(agentID, allowance, reqctx.ChainID())
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID coin.Type) *big.Int {
	var ret *big.Int
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNativeTokenBalance(agentID, nativeTokenID, reqctx.ChainID())
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokenBalanceTotal(coinType coin.Type) *big.Int {
	var ret *big.Int
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNativeTokenBalanceTotal(coinType)
	})
	return ret
}

func (reqctx *requestContext) GetNativeTokens(agentID isc.AgentID) isc.CoinBalances {
	var ret isc.CoinBalances
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNativeTokens(agentID, reqctx.ChainID())
	})
	return ret
}

func (reqctx *requestContext) GetAccountNFTs(agentID isc.AgentID) (ret []sui.ObjectID) {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetAccountNFTs(agentID)
	})
	return ret
}

func (reqctx *requestContext) GetNFTData(nftID sui.ObjectID) (ret *isc.NFT) {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		ret = s.GetNFTData(nftID)
	})
	return ret
}

func (reqctx *requestContext) GetSenderTokenBalanceForFees() uint64 {
	sender := reqctx.req.SenderAccount()
	if sender == nil {
		return 0
	}
	return reqctx.GetBaseTokensBalanceDiscardRemainder(sender)
}

func (reqctx *requestContext) requestLookupKey() blocklog.RequestLookupKey {
	return blocklog.NewRequestLookupKey(reqctx.vm.stateDraft.BlockIndex(), reqctx.requestIndex)
}

func (reqctx *requestContext) eventLookupKey() *blocklog.EventLookupKey {
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
		err = blocklog.NewStateWriter(s).SaveRequestReceipt(receipt, key)
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
		blocklog.NewStateWriter(s).SaveEvent(eventKey, event)
	})
	reqctx.requestEventIndex++
}

// updateOffLedgerRequestNonce updates stored nonce for ISC off ledger requests
func (reqctx *requestContext) updateOffLedgerRequestNonce() {
	reqctx.callAccounts(func(s *accounts.StateWriter) {
		s.IncrementNonce(reqctx.req.SenderAccount(), reqctx.ChainID())
	})
}

// adjustL2BaseTokensIfNeeded adjust L2 ledger for base tokens if the L1 changed because of storage deposit changes
func (reqctx *requestContext) adjustL2BaseTokensIfNeeded(adjustment int64, account isc.AgentID) {
	if adjustment == 0 {
		return
	}
	err := panicutil.CatchPanicReturnError(func() {
		reqctx.callAccounts(func(s *accounts.StateWriter) {
			s.AdjustAccountBaseTokens(account, adjustment, reqctx.ChainID())
		})
	}, accounts.ErrNotEnoughFunds)
	if err != nil {
		panic(vmexceptions.ErrNotEnoughFundsForInternalStorageDeposit)
	}
}
