package vmcontext

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

func (vmctx *VMContext) BuildTransactionEssence(stateData *iscp.StateData) (*iotago.TransactionEssence, []byte) {
	return vmctx.txbuilder.BuildTransactionEssence(stateData)
}

func (vmctx *VMContext) createTxBuilderSnapshot() *vmtxbuilder.AnchorTransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *VMContext) restoreTxBuilderSnapshot(snapshot *vmtxbuilder.AnchorTransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *VMContext) loadNativeTokenOutput(id *iotago.NativeTokenID) (*iotago.BasicOutput, *iotago.UTXOInput) {
	var retOut *iotago.BasicOutput
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
		panic(xerrors.Errorf("internal: can't find UTXO input for block index %d, output index %d", blockIndex, outputIndex))
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
		panic(xerrors.Errorf("internal: can't find UTXO input for block index %d, output index %d", blockIndex, outputIndex))
	}
	return retOut, retInp
}

func (vmctx *VMContext) getUTXOInput(blockIndex uint32, outputIndex uint16) (ret *iotago.UTXOInput) {
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		ret = blocklog.GetUTXOInput(s, blockIndex, outputIndex)
	})
	return
}

func (vmctx *VMContext) loadNFT(id iotago.NFTID) (*iotago.NFTOutput, *iotago.UTXOInput) {
	var retOut *iotago.NFTOutput
	var retInp *iotago.UTXOInput
	var blockIndex uint32
	var outputIndex uint16

	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		retOut, blockIndex, outputIndex = accounts.GetNFTOutput(s, id, vmctx.ChainID())
	})
	if retOut == nil {
		return nil, nil
	}
	if retInp = vmctx.getUTXOInput(blockIndex, outputIndex); retOut == nil {
		panic(xerrors.Errorf("internal: can't find UTXO input for block index %d, output index %d", blockIndex, outputIndex))
	}
	return retOut, retInp
}
