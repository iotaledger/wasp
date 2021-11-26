package vmcontext

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/iscp"
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
func (vmctx *VMContext) loadNativeTokensOnChain(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
	inp, ok := vmctx.findNativeTokenUTXOInput(id)
	if !ok {
		return nil, iotago.UTXOInput{}
	}
	b := vmctx.GetTokenBalanceTotal(&id)
	if b == nil {
		return nil, iotago.UTXOInput{}
	}
	return b, inp
}

// findNativeTokenUTXOInput call `blocklog` to find the UTXO input for the native token ID
func (vmctx *VMContext) findNativeTokenUTXOInput(id iotago.NativeTokenID) (iotago.UTXOInput, bool) {
	// TODO
	panic("not implemented")
}
