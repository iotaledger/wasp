package vmimpl

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func (vmctx *vmContext) StateMetadata(l1Commitment *state.L1Commitment) []byte {
	stateMetadata := transaction.StateMetadata{
		L1Commitment: l1Commitment,
	}

	stateMetadata.SchemaVersion = root.NewStateReaderFromChainState(vmctx.stateDraft).GetSchemaVersion()

	// On error, the publicURL is len(0)
	govState := governance.NewStateReaderFromChainState(vmctx.stateDraft)
	stateMetadata.PublicURL = govState.GetPublicURL()
	stateMetadata.GasFeePolicy = govState.GetGasFeePolicy()

	return stateMetadata.Bytes()
}

func (vmctx *vmContext) BuildTransactionEssence(stateMetadata []byte) iotago.TransactionData {
	ptb := vmctx.txbuilder.BuildTransactionEssence(stateMetadata)
	return iotago.NewProgrammable(
		vmctx.stateAnchor().Owner.AsSuiAddress(),
		ptb,
		[]*iotago.ObjectRef{}, // TODO: handle L1 gas
		0,
		0,
	)
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
