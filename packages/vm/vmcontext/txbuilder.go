package vmcontext

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
)

// implements transaction builder used internally by the VM during batch run

type txbuilder struct {
	chainOutput    *iotago.AliasOutput
	consumedInputs []iotago.UTXOInput
	outputsCount   int
	// TODO
}

// TODO
func newTxBuilder(chainOutput *iotago.AliasOutput, chainOutputID iotago.UTXOInput) *txbuilder {
	return &txbuilder{
		chainOutput:    chainOutput,
		consumedInputs: []iotago.UTXOInput{chainOutputID},
	}
}

func (txb *txbuilder) clone() *txbuilder {
	ret := newTxBuilder(txb.chainOutput, txb.consumedInputs[0])
	copy(ret.consumedInputs, txb.consumedInputs)
	ret.outputsCount = txb.outputsCount
	// TODO
	return ret
}

func (txb *txbuilder) addConsumedInput(inp iotago.UTXOInput) int {
	txb.consumedInputs = append(txb.consumedInputs, inp)
	return len(txb.consumedInputs) - 1
}

func (txb *txbuilder) numConsumedInputs() int {
	return len(txb.consumedInputs)
}

func (txb *txbuilder) inputs() iotago.Inputs {
	ret := make(iotago.Inputs, len(txb.consumedInputs))
	for i := range txb.consumedInputs {
		ret[i] = &txb.consumedInputs[i]
	}
	return ret
}

func (txb *txbuilder) inputsFull() bool {
	return len(txb.consumedInputs) >= iotago.MaxInputsCount
}

func (txb *txbuilder) outputsFull() bool {
	return txb.outputsCount >= iotago.MaxOutputsCount
}

/////////// vmcontext methods

// initTxBuilder creates anchor transaction builder for the block
func (vmctx *VMContext) initTxBuilder() {
	vmctx.txbuilder = newTxBuilder(vmctx.chainInput, vmctx.chainInputID)
	vmctx.txbuilderSnapshots = make(map[int]*txbuilder)
}

func (vmctx *VMContext) BuildTransactionEssence(stateHash hashing.HashValue, timestamp time.Time) (*iotago.TransactionEssence, error) {
	//if err := vmctx.txBuilder.AddAliasOutputAsRemainder(vmctx.chainID.AsAddress(), stateHash[:]); err != nil {
	//	return nil, xerrors.Errorf("mustFinalizeRequestCall: %v", err)
	//}
	//tx, _, err := vmctx.txBuilder.WithTimestamp(timestamp).BuildEssence()
	//if err != nil {
	//	return nil, err
	//}
	//return tx, nil
	return nil, nil
}

func (vmctx *VMContext) createTxBuilderSnapshot(id int) {
	vmctx.txbuilderSnapshots[id] = vmctx.txbuilder.clone()
}

func (vmctx *VMContext) restoreTxBuilderSnapshot(id int) {
	vmctx.txbuilder = vmctx.txbuilderSnapshots[id]
	delete(vmctx.txbuilderSnapshots, id)
}

func (vmctx *VMContext) clearTxBuilderSnapshots() {
	vmctx.txbuilderSnapshots = make(map[int]*txbuilder)
}
