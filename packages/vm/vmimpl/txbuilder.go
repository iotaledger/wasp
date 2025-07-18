package vmimpl

import (
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
	"github.com/iotaledger/wasp/v2/packages/vm/vmtxbuilder"
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

func (vmctx *vmContext) deductTopUpFeeFromCommonAccount(fee coin.Value) {
	bal := isc.NewCoinBalances()
	bal.AddBaseTokens(fee)
	vmctx.withStateUpdate(func(chainState kv.KVStore) {
		vmctx.accountsStateWriterFromChainState(chainState).
			DebitFromAccount(accounts.CommonAccount(), bal)
	})
}

func (vmctx *vmContext) commonAccountBalance() coin.Value {
	return accounts.NewStateReaderFromChainState(vmctx.schemaVersion, vmctx.stateDraft).
		GetBaseTokensBalanceDiscardExtraDecimals(accounts.CommonAccount())
}
