package vmcontext

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/iscp"
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
func (vmctx *VMContext) loadNativeTokensOnChain(id *iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput) {
	inp := vmctx.findNativeTokenUTXOInput(id)
	if inp == nil {
		return nil, nil
	}
	b := vmctx.GetNativeTokenBalanceTotal(id)
	if b == nil {
		return nil, nil
	}
	return b, inp
}

// findNativeTokenUTXOInput call `blocklog` to find the UTXO input for the native token ID
func (vmctx *VMContext) findNativeTokenUTXOInput(id *iotago.NativeTokenID) *iotago.UTXOInput {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blocklog.GetUTXOIDForAsset(vmctx.State(), id)
}

func (vmctx *VMContext) loadFoundry(serNum uint32) (*iotago.FoundryOutput, *iotago.UTXOInput) {
	// TODO
	return nil, nil
}
