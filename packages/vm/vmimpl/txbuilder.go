package vmimpl

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func (vmctx *vmContext) StateMetadata(l1Commitment *state.L1Commitment, gasCoin *coin.CoinWithRef) []byte {
	stateMetadata := transaction.StateMetadata{
		L1Commitment: l1Commitment,
	}

	stateMetadata.SchemaVersion = root.NewStateReaderFromChainState(vmctx.stateDraft).GetSchemaVersion()

	// On error, the publicURL is len(0)
	govState := governance.NewStateReaderFromChainState(vmctx.stateDraft)
	stateMetadata.PublicURL = govState.GetPublicURL()
	stateMetadata.GasFeePolicy = govState.GetGasFeePolicy()
	stateMetadata.GasCoinObjectID = gasCoin.Ref.ObjectID

	return stateMetadata.Bytes()
}

func (vmctx *vmContext) createTxBuilderSnapshot() vmtxbuilder.TransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *vmContext) restoreTxBuilderSnapshot(snapshot vmtxbuilder.TransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *vmContext) getTotalL2Coins() isc.CoinBalances {
	return vmctx.accountsStateWriterFromChainState(vmctx.stateDraft).GetTotalL2FungibleTokens()
}

func (vmctx *vmContext) deductTopUpFeeFromCommonAccount(fee coin.Value) {
	bal := isc.NewCoinBalances()
	bal.AddBaseTokens(fee)

	before := vmctx.commonAccountBalance()
	vmctx.withStateUpdate(func(chainState kv.KVStore) {
		vmctx.accountsStateWriterFromChainState(chainState).
			DebitFromAccount(accounts.CommonAccount(), bal, vmctx.ChainID())
	})
	after := vmctx.commonAccountBalance()
	vmctx.task.Log.Debugf("deducted %s from common account, balance before: %s, after: %s", fee, before, after)
}

func (vmctx *vmContext) commonAccountBalance() coin.Value {
	return accounts.NewStateReaderFromChainState(vmctx.schemaVersion, vmctx.stateDraft).
		GetBaseTokensBalanceDiscardExtraDecimals(accounts.CommonAccount(), vmctx.ChainID())
}
