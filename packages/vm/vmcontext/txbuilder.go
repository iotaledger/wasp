package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (vmctx *VMContext) BuildTransactionEssence(stateData *iscp.StateData) *iotago.TransactionEssence {
	return vmctx.txbuilder.BuildTransactionEssence(stateData)
}

func (vmctx *VMContext) createTxBuilderSnapshot() *vmtxbuilder.AnchorTransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *VMContext) restoreTxBuilderSnapshot(snapshot *vmtxbuilder.AnchorTransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *VMContext) loadNativeTokenOutput(id *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
	var retOut *iotago.ExtendedOutput
	var retInp *iotago.UTXOInput
	var blockIndex uint32
	var outputIndex uint16

	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		retOut, blockIndex, outputIndex = accounts.GetNativeTokenOutput(s, id, vmctx.ChainID())
	})
	if retOut == nil {
		return nil, nil
	}
	if retInp = vmctx.getUTXOInput(blockIndex, outputIndex); retOut == nil {
		panic(xerrors.Errorf("internal: can't find AsUTXO input for block index %d, output index %d", blockIndex, outputIndex))
	}
	return retOut, retInp
}

func (vmctx *VMContext) loadFoundry(serNum uint32) (*iotago.FoundryOutput, *iotago.UTXOInput) {
	var retOut *iotago.FoundryOutput
	var retInp *iotago.UTXOInput
	var blockIndex uint32
	var outputIndex uint16

	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		retOut, blockIndex, outputIndex = accounts.GetFoundryOutput(s, serNum, vmctx.ChainID())
	})
	if retOut == nil {
		return nil, nil
	}
	if retInp = vmctx.getUTXOInput(blockIndex, outputIndex); retOut == nil {
		panic(xerrors.Errorf("internal: can't find AsUTXO input for block index %d, output index %d", blockIndex, outputIndex))
	}
	return retOut, retInp
}

func (vmctx *VMContext) getUTXOInput(blockIndex uint32, outputIndex uint16) (ret *iotago.UTXOInput) {
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		ret = blocklog.GetUTXOInput(s, blockIndex, outputIndex)
	})
	return
}
