package vmimpl

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func (vmctx *vmContext) stateMetadata(stateCommitment *state.L1Commitment) []byte {
	stateMetadata := transaction.StateMetadata{
		Version:      transaction.StateMetadataSupportedVersion,
		L1Commitment: stateCommitment,
	}

	withContractState(vmctx.stateDraft, root.Contract, func(s kv.KVStore) {
		stateMetadata.SchemaVersion = root.GetSchemaVersion(s)
	})

	withContractState(vmctx.stateDraft, governance.Contract, func(s kv.KVStore) {
		// On error, the publicURL is len(0)
		stateMetadata.PublicURL, _ = governance.GetPublicURL(s)
		stateMetadata.GasFeePolicy = governance.MustGetGasFeePolicy(s)
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
	withContractState(vmctx.stateDraft, accounts.Contract, func(s kv.KVStore) {
		out, id = accounts.GetNativeTokenOutput(s, nativeTokenID, vmctx.ChainID())
	})
	return
}

func (vmctx *vmContext) loadFoundry(serNum uint32) (out *iotago.FoundryOutput, id iotago.OutputID) {
	withContractState(vmctx.stateDraft, accounts.Contract, func(s kv.KVStore) {
		out, id = accounts.GetFoundryOutput(s, serNum, vmctx.ChainID())
	})
	return
}

func (vmctx *vmContext) getOutputID(blockIndex uint32, outputIndex uint16) iotago.OutputID {
	if blockIndex == vmctx.stateAnchor().StateIndex {
		return iotago.OutputIDFromTransactionIDAndIndex(vmctx.stateAnchor().OutputID.TransactionID(), outputIndex)
	}
	var outputID iotago.OutputID
	var ok bool
	withContractState(vmctx.stateDraft, blocklog.Contract, func(s kv.KVStore) {
		outputID, ok = blocklog.GetOutputID(s, blockIndex, outputIndex)
	})
	if !ok {
		panic(fmt.Errorf("UTXO input for block index %d, output index %d not found", blockIndex, outputIndex))
	}
	return outputID
}

func (vmctx *vmContext) loadNFT(nftID iotago.NFTID) (out *iotago.NFTOutput, id iotago.OutputID) {
	withContractState(vmctx.stateDraft, accounts.Contract, func(s kv.KVStore) {
		out, id = accounts.GetNFTOutput(s, nftID)
	})
	return
}

func (vmctx *vmContext) loadTotalFungibleTokens() *isc.Assets {
	var totalAssets *isc.Assets
	withContractState(vmctx.stateDraft, accounts.Contract, func(s kv.KVStore) {
		totalAssets = accounts.GetTotalL2FungibleTokens(s)
	})
	return totalAssets
}
