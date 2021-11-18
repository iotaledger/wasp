package vmcontext

import (
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
)

// implements transaction builder used internally by the VM during batch run

type txbuilder struct {
	chainOutput        *iotago.AliasOutput
	consumedInputs     []iotago.UTXOInput
	outputsCount       int // maintained during the run. Final outputs will be constructed in the end
	initialIotaAmount  uint64
	updateIotaAmountTo uint64
	deltaNativeTokens  map[iotago.NativeTokenID]*big.Int
	// TODO
}

// TODO
func newTxBuilder(chainOutput *iotago.AliasOutput, chainOutputID iotago.UTXOInput) *txbuilder {
	return &txbuilder{
		chainOutput:       chainOutput,
		consumedInputs:    []iotago.UTXOInput{chainOutputID},
		deltaNativeTokens: make(map[iotago.NativeTokenID]*big.Int),
	}
}

func (txb *txbuilder) clone() *txbuilder {
	ret := newTxBuilder(txb.chainOutput, txb.consumedInputs[0])
	copy(ret.consumedInputs, txb.consumedInputs)
	ret.outputsCount = txb.outputsCount
	for k, v := range txb.deltaNativeTokens {
		ret.deltaNativeTokens[k] = new(big.Int).Set(v)
	}
	// TODO the rest
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

func (txb *txbuilder) addDeltaIotas(delta uint64) {
	// safe arithmetics
	n := txb.updateIotaAmountTo + delta
	if n < txb.updateIotaAmountTo {
		panic("addDeltaIotas: overflow")
	}
	txb.updateIotaAmountTo = n
}

func (txb *txbuilder) subDeltaIotas(delta uint64) {
	// safe arithmetics
	if delta > txb.updateIotaAmountTo {
		panic("subDeltaIotas: overflow")
	}
	txb.updateIotaAmountTo -= delta
}

// use negative to subtract
func (txb *txbuilder) addDeltaNativeToken(id iotago.NativeTokenID, delta *big.Int) {
	b, ok := txb.deltaNativeTokens[id]
	if !ok {
		b = new(big.Int).Set(delta)
	} else {
		// TODO safe arithmetic
		b.Add(b, delta)
	}
	// see https://stackoverflow.com/questions/64257065/is-there-another-way-of-testing-if-a-big-int-is-0
	if len(b.Bits()) == 0 {
		delete(txb.deltaNativeTokens, id)
	} else {
		txb.deltaNativeTokens[id] = b
	}
}

/////////// vmcontext methods

// initTxBuilder creates anchor transaction builder for the block
func (vmctx *VMContext) initTxBuilder() {
	vmctx.txbuilder = newTxBuilder(vmctx.chainInput, vmctx.chainInputID)
	vmctx.txbuilderSnapshots = make(map[int]*txbuilder)
	// fetch current iota amount from the state into initialIotaAmount
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
