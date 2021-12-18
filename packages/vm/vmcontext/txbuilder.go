package vmcontext

import (
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

func (vmctx *VMContext) loadNativeTokenOutput(id *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blocklog.GetNativeTokenOutput(vmctx.State(), id)
}

func (vmctx *VMContext) loadFoundry(serNum uint32) (*iotago.FoundryOutput, *iotago.UTXOInput) {
	vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	defer vmctx.popCallContext()

	return blocklog.GetFoundryOutput(vmctx.State(), serNum)
}
