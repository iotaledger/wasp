package vmcontext

import (
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
)

// implements transaction builder used internally by the VM during batch run

type txbuilder struct {
	vmctx *VMContext
	// anchor output of the chain
	chainOutput *iotago.AliasOutput
	// already consumed inputs
	consumedInputs []iotago.UTXOInput
	// maintained during the run. Final outputs will be constructed in the end
	outputsCount        int
	balanceIotasInitial *uint64
	balanceIotas        uint64
	balanceNativeTokens map[iotago.NativeTokenID]*nativeTokenBalance
	// TODO
}

type nativeTokenBalance struct {
	initial *big.Int
	balance *big.Int
}

func (txb *txbuilder) clone() *txbuilder {
	ret := txb.vmctx.newTxBuilder()
	copy(ret.consumedInputs, txb.consumedInputs)
	ret.outputsCount = txb.outputsCount
	ret.balanceIotas = txb.balanceIotas
	i := *txb.balanceIotasInitial
	ret.balanceIotasInitial = &i
	ret.balanceIotas = txb.balanceIotas
	for k, v := range txb.balanceNativeTokens {
		ret.balanceNativeTokens[k] = &nativeTokenBalance{
			initial: new(big.Int).Set(v.balance),
			balance: new(big.Int).Set(v.balance),
		}
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

func (txb *txbuilder) numInputsConsumed() int {
	ret := len(txb.consumedInputs)
	if txb.balanceIotasInitial != nil && *txb.balanceIotasInitial != txb.balanceIotas && txb.balanceIotas > 0 {
		ret++
	}
	for _, v := range txb.balanceNativeTokens {
		if v.balance.Cmp(v.initial) != 0 {
			ret++
		}
	}
	return ret
}

func (txb *txbuilder) inputsFull() bool {
	return txb.numInputsConsumed() >= iotago.MaxInputsCount
}

func (txb *txbuilder) numOutputsConsumed() int {
	ret := 1 // for chain output
	for _, v := range txb.balanceNativeTokens {
		// see https://stackoverflow.com/questions/64257065/is-there-another-way-of-testing-if-a-big-int-is-0
		if v.balance.Cmp(v.initial) != 0 && len(v.balance.Bits()) > 0 {
			ret++
		}
	}
	return ret
}

func (txb *txbuilder) outputsFull() bool {
	return txb.numOutputsConsumed() >= iotago.MaxOutputsCount
}

func (txb *txbuilder) ensureIotasBalance() {
	if txb.balanceIotasInitial != nil {
		return
	}
	b := txb.vmctx.loadOnChainIotas()
	txb.balanceIotasInitial = &b
	txb.balanceIotas = b
}

func (txb *txbuilder) addDeltaIotas(delta uint64) {
	txb.ensureIotasBalance()
	// safe arithmetics
	n := txb.balanceIotas + delta
	if n < txb.balanceIotas {
		panic("addDeltaIotas: overflow")
	}
	txb.balanceIotas = n
}

func (txb *txbuilder) subDeltaIotas(delta uint64) {
	// safe arithmetics
	if delta > txb.balanceIotas {
		panic("subDeltaIotas: overflow")
	}
	txb.balanceIotas -= delta
}

func (txb *txbuilder) ensureNativeTokenBalance(id iotago.NativeTokenID) *nativeTokenBalance {
	if b, ok := txb.balanceNativeTokens[id]; ok {
		return b
	}
	b := &nativeTokenBalance{
		balance: txb.vmctx.loadNativeTokensOnChain(id),
	}
	txb.balanceNativeTokens[id] = b
	return b
}

// use negative to subtract
func (txb *txbuilder) addDeltaNativeToken(id iotago.NativeTokenID, delta *big.Int) {
	b := txb.ensureNativeTokenBalance(id)
	// TODO safe arithmetic
	b.balance.Add(b.balance, delta)
}

/////////// vmcontext methods

func (vmctx *VMContext) newTxBuilder() *txbuilder {
	return &txbuilder{
		vmctx:               vmctx,
		chainOutput:         vmctx.chainInput,
		consumedInputs:      []iotago.UTXOInput{vmctx.chainInputID},
		balanceNativeTokens: make(map[iotago.NativeTokenID]*nativeTokenBalance),
	}
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

func (vmctx *VMContext) loadOnChainIotas() uint64 {
	panic("not implemented")
}

func (vmctx *VMContext) loadNativeTokensOnChain(id iotago.NativeTokenID) *big.Int {
	panic("not implemented")
}
