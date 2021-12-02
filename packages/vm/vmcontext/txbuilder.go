package vmcontext

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"

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

// loadNativeTokensOnChain calls
// 1. `blocklog` to find UTXO ID for a specific token ID, if any
// 2. `accounts` to load the balance
// Returns nil if balance is empty (zero)
func (vmctx *VMContext) loadNativeTokensOnChain(id iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
	inp, ok := vmctx.findNativeTokenUTXOInput(id)
	if !ok {
		return nil, nil
	}
	b := vmctx.GetTokenBalanceTotal(&id)
	if b == nil {
		return nil, nil
	}
	return b, inp
}

// findNativeTokenUTXOInput call `blocklog` to find the UTXO input for the native token ID
func (vmctx *VMContext) findNativeTokenUTXOInput(id iotago.NativeTokenID) (*iotago.UTXOInput, bool) {
	stateIndex, outputIndex, err := vmctx.getNativeTokenUtxoIndex(id)
	if err != nil {
		return nil, false
	}
	return &iotago.UTXOInput{
		TransactionID:          vmctx.getTxIDForStateIndex(stateIndex),
		TransactionOutputIndex: outputIndex,
	}, true
}

func (vmctx *VMContext) getNativeTokenUtxoIndex(id iotago.NativeTokenID) (stateIndex uint32, outputIndex uint16, err error) {
	vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return accounts.GetUtxoForAsset(vmctx.State(), id)
}

func (vmctx *VMContext) getTxIDForStateIndex(stateIndex uint32) iotago.TransactionID {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	transactionId := blocklog.GetAnchorTransactionIDByBlockIndex(vmctx.State(), stateIndex)

	return transactionId
}
