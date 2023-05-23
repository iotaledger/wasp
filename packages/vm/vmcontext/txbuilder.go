package vmcontext

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
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

func (vmctx *VMContext) StateMetadata(stateCommitment *state.L1Commitment) []byte {
	stateMetadata := transaction.StateMetadata{
		Version:      transaction.StateMetadataSupportedVersion,
		L1Commitment: stateCommitment,
	}

	vmctx.callCore(root.Contract, func(s kv.KVStore) {
		stateMetadata.SchemaVersion = root.GetSchemaVersion(s)
	})

	vmctx.callCore(governance.Contract, func(s kv.KVStore) {
		// On error, the publicURL is len(0)
		stateMetadata.PublicURL, _ = governance.GetPublicURL(s)
		stateMetadata.GasFeePolicy = governance.MustGetGasFeePolicy(s)
	})

	return stateMetadata.Bytes()
}

func (vmctx *VMContext) BuildTransactionEssence(stateCommitment *state.L1Commitment, assertTxbuilderBalanced bool) (*iotago.TransactionEssence, []byte) {
	if vmctx.currentStateUpdate == nil {
		// create a temporary empty state update, so that vmctx.callCore works and contracts state can be read
		vmctx.currentStateUpdate = NewStateUpdate()
		defer func() { vmctx.currentStateUpdate = nil }()
	}
	stateMetadata := vmctx.StateMetadata(stateCommitment)
	essence, inputsCommitment := vmctx.txbuilder.BuildTransactionEssence(stateMetadata)
	if assertTxbuilderBalanced {
		vmctx.txbuilder.MustBalanced()
	}
	return essence, inputsCommitment
}

func (vmctx *VMContext) createTxBuilderSnapshot() *vmtxbuilder.AnchorTransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *VMContext) restoreTxBuilderSnapshot(snapshot *vmtxbuilder.AnchorTransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *VMContext) loadNativeTokenOutput(nativeTokenID iotago.NativeTokenID) (*iotago.BasicOutput, iotago.OutputID) {
	var retOut *iotago.BasicOutput
	var blockIndex uint32
	var outputIndex uint16
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		retOut, blockIndex, outputIndex = accounts.GetNativeTokenOutput(s, nativeTokenID, vmctx.ChainID())
	})
	if retOut == nil {
		return nil, iotago.OutputID{}
	}

	outputID := vmctx.getOutputID(blockIndex, outputIndex)

	return retOut, outputID
}

func (vmctx *VMContext) loadFoundry(serNum uint32) (*iotago.FoundryOutput, iotago.OutputID) {
	var foundryOutput *iotago.FoundryOutput
	var blockIndex uint32
	var outputIndex uint16
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		foundryOutput, blockIndex, outputIndex = accounts.GetFoundryOutput(s, serNum, vmctx.ChainID())
	})
	if foundryOutput == nil {
		return nil, iotago.OutputID{}
	}

	outputID := vmctx.getOutputID(blockIndex, outputIndex)

	return foundryOutput, outputID
}

func (vmctx *VMContext) getOutputID(blockIndex uint32, outputIndex uint16) iotago.OutputID {
	if blockIndex == vmctx.StateAnchor().StateIndex {
		return iotago.OutputIDFromTransactionIDAndIndex(vmctx.StateAnchor().OutputID.TransactionID(), outputIndex)
	}
	var outputID iotago.OutputID
	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		outputID, err = blocklog.GetOutputID(s, blockIndex, outputIndex)
	})
	if err != nil {
		panic(fmt.Errorf("internal: can't find UTXO input for block index %d, output index %d, error: %w", blockIndex, outputIndex, err))
	}
	return outputID
}

func (vmctx *VMContext) loadNFT(id iotago.NFTID) (*iotago.NFTOutput, iotago.OutputID) {
	var nftOutput *iotago.NFTOutput
	var blockIndex uint32
	var outputIndex uint16
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		nftOutput, blockIndex, outputIndex = accounts.GetNFTOutput(s, id, vmctx.ChainID())
	})
	if nftOutput == nil {
		return nil, iotago.OutputID{}
	}

	outputID := vmctx.getOutputID(blockIndex, outputIndex)

	return nftOutput, outputID
}

func (vmctx *VMContext) loadTotalFungibleTokens() *isc.Assets {
	var totalAssets *isc.Assets
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		totalAssets = accounts.GetTotalL2FungibleTokens(s)
	})
	return totalAssets
}
