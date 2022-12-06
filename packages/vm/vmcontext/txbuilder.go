package vmcontext

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

func (vmctx *VMContext) BuildTransactionEssence(stateData *state.L1Commitment) (*iotago.TransactionEssence, []byte) {
	return vmctx.txbuilder.BuildTransactionEssence(stateData)
}

// CalcTransactionSubEssenceHash builds transaction essence from tx builder
// data assuming all zeroes in the L1 commitment. Returns hash of it.
// It is needed for fraud proofs
func (vmctx *VMContext) CalcTransactionSubEssenceHash() blocklog.TransactionEssenceHash {
	essence, _ := vmctx.txbuilder.BuildTransactionEssence(state.L1CommitmentNil)

	return blocklog.CalcTransactionEssenceHash(essence)
}

func (vmctx *VMContext) createTxBuilderSnapshot() *vmtxbuilder.AnchorTransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *VMContext) restoreTxBuilderSnapshot(snapshot *vmtxbuilder.AnchorTransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *VMContext) loadNativeTokenOutput(id *iotago.NativeTokenID) (*iotago.BasicOutput, iotago.OutputID) {
	var retOut *iotago.BasicOutput
	var blockIndex uint32
	var outputIndex uint16
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		retOut, blockIndex, outputIndex = accounts.GetNativeTokenOutput(s, id, vmctx.ChainID())
	})
	if retOut == nil {
		return nil, iotago.OutputID{}
	}

	outputID, err := vmctx.getOutputID(blockIndex, outputIndex)
	if err != nil {
		panic(fmt.Errorf("internal: can't find UTXO input for block index %d, output index %d, error: %w", blockIndex, outputIndex, err))
	}

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

	outputID, err := vmctx.getOutputID(blockIndex, outputIndex)
	if err != nil {
		panic(fmt.Errorf("internal: can't find UTXO input for block index %d, output index %d, error: %w", blockIndex, outputIndex, err))
	}

	return foundryOutput, outputID
}

func (vmctx *VMContext) getOutputID(blockIndex uint32, outputIndex uint16) (iotago.OutputID, error) {
	var outputID iotago.OutputID
	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		outputID, err = blocklog.GetOutputID(s, blockIndex, outputIndex)
	})
	if err != nil {
		return iotago.OutputID{}, err
	}

	return outputID, nil
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

	outputID, err := vmctx.getOutputID(blockIndex, outputIndex)
	if err != nil {
		panic(fmt.Errorf("internal: can't find UTXO input for block index %d, output index %d, error: %w", blockIndex, outputIndex, err))
	}

	return nftOutput, outputID
}
