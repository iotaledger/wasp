package vmimpl

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func (vmctx *vmContext) stateMetadata(stateCommitment *state.L1Commitment) []byte {
	stateMetadata := transaction.StateMetadata{
		Version:      transaction.StateMetadataSupportedVersion,
		L1Commitment: stateCommitment,
	}

	stateMetadata.SchemaVersion = root.NewStateReaderFromChainState(vmctx.stateDraft).GetSchemaVersion()

	withContractState(vmctx.stateDraft, governance.Contract, func(s kv.KVStore) {
		// On error, the publicURL is len(0)
		govState := governance.NewStateReader(s)
		stateMetadata.PublicURL = govState.GetPublicURL()
		stateMetadata.GasFeePolicy = govState.GetGasFeePolicy()
	})

	return stateMetadata.Bytes()
}

func (vmctx *vmContext) BuildTransactionEssence(stateCommitment *state.L1Commitment, assertTxbuilderBalanced bool) (*iotago.TransactionEssence, []byte) {
	stateMetadata := vmctx.stateMetadata(stateCommitment)
	essence, inputsCommitment := vmctx.txbuilder.BuildTransactionEssence(stateMetadata)
	if assertTxbuilderBalanced {
		vmctx.txbuilder.MustBalanced()
	}
	return essence, inputsCommitment
}

func (vmctx *vmContext) createTxBuilderSnapshot() *vmtxbuilder.AnchorTransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *vmContext) restoreTxBuilderSnapshot(snapshot *vmtxbuilder.AnchorTransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *vmContext) loadNativeTokenOutput(nativeTokenID iotago.NativeTokenID) (out *iotago.BasicOutput, id iotago.OutputID) {
	vmctx.withAccountsState(vmctx.stateDraft, func(s *accounts.StateWriter) {
		out, id = s.GetNativeTokenOutput(nativeTokenID, vmctx.ChainID())
	})
	return
}

func (vmctx *vmContext) loadFoundry(serNum uint32) (out *iotago.FoundryOutput, id iotago.OutputID) {
	vmctx.withAccountsState(vmctx.stateDraft, func(s *accounts.StateWriter) {
		out, id = s.GetFoundryOutput(serNum, vmctx.ChainID())
	})
	return
}

func (vmctx *vmContext) loadNFT(nftID iotago.NFTID) (out *iotago.NFTOutput, id iotago.OutputID) {
	vmctx.withAccountsState(vmctx.stateDraft, func(s *accounts.StateWriter) {
		out, id = s.GetNFTOutput(nftID)
	})
	return
}

func (vmctx *vmContext) loadTotalFungibleTokens() *isc.Assets {
	var totalAssets *isc.Assets
	vmctx.withAccountsState(vmctx.stateDraft, func(s *accounts.StateWriter) {
		totalAssets = s.GetTotalL2FungibleTokens()
	})
	return totalAssets
}
