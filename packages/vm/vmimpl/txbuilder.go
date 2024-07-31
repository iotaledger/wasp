package vmimpl

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func (vmctx *vmContext) StateMetadata(stateCommitment *state.L1Commitment) []byte {
	stateMetadata := transaction.StateMetadata{
		Version:      transaction.StateMetadataSupportedVersion,
		L1Commitment: stateCommitment,
	}

	stateMetadata.SchemaVersion = root.NewStateReaderFromChainState(vmctx.stateDraft).GetSchemaVersion()

	// On error, the publicURL is len(0)
	govState := governance.NewStateReaderFromChainState(vmctx.stateDraft)
	stateMetadata.PublicURL = govState.GetPublicURL()
	stateMetadata.GasFeePolicy = govState.GetGasFeePolicy()

	return stateMetadata.Bytes()
}

func (vmctx *vmContext) BuildTransactionEssence(stateCommitment *state.L1Commitment, assertTxbuilderBalanced bool) (*iotago.TransactionEssence, []byte) {
	stateMetadata := vmctx.StateMetadata(stateCommitment)
	essence, inputsCommitment := vmctx.txbuilder.BuildTransactionEssence(stateMetadata)
	if assertTxbuilderBalanced {
		vmctx.txbuilder.MustBalanced()
	}
	return essence, inputsCommitment
}

func (vmctx *vmContext) createTxBuilderSnapshot() vmtxbuilder.TransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *vmContext) restoreTxBuilderSnapshot(snapshot vmtxbuilder.TransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *vmContext) loadNativeTokenOutput(nativeTokenID isc.NativeTokenID) (out *iotago.BasicOutput, id iotago.OutputID) {
	return vmctx.accountsStateWriterFromChainState(vmctx.stateDraft).GetNativeTokenOutput(nativeTokenID, vmctx.ChainID())
}

func (vmctx *vmContext) loadFoundry(serNum uint32) (out *iotago.FoundryOutput, id iotago.OutputID) {
	return vmctx.accountsStateWriterFromChainState(vmctx.stateDraft).GetFoundryOutput(serNum, vmctx.ChainID())
}

func (vmctx *vmContext) loadNFT(nftID isc.NFTID) (out *iotago.NFTOutput, id iotago.OutputID) {
	return vmctx.accountsStateWriterFromChainState(vmctx.stateDraft).GetNFTOutput(nftID)
}

func (vmctx *vmContext) loadTotalFungibleTokens() *isc.Assets {
	return vmctx.accountsStateWriterFromChainState(vmctx.stateDraft).GetTotalL2FungibleTokens()
}
