package vmtxbuilder

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// nativeTokenBalance represents on-chain account of the specific native token
type nativeTokenBalance struct {
	nativeTokenID     iotago.NativeTokenID
	accountingInputID iotago.OutputID     // if in != nil, otherwise zeroOutputID
	accountingInput   *iotago.BasicOutput // if nil it means output does not exist, this is new account for the token_id
	accountingOutput  *iotago.BasicOutput // current balance of the token_id on the chain
}

func (n *nativeTokenBalance) Clone() *nativeTokenBalance {
	nativeTokenID := iotago.NativeTokenID{}
	copy(nativeTokenID[:], n.nativeTokenID[:])

	outputID := iotago.OutputID{}
	copy(outputID[:], n.accountingInputID[:])

	return &nativeTokenBalance{
		nativeTokenID:     nativeTokenID,
		accountingInputID: outputID,
		accountingInput:   cloneInternalBasicOutputOrNil(n.accountingInput),
		accountingOutput:  cloneInternalBasicOutputOrNil(n.accountingOutput),
	}
}

// producesAccountingOutput if value update produces UTXO of the corresponding total native token balance
func (n *nativeTokenBalance) producesAccountingOutput() bool {
	if n.identicalInOut() {
		// value didn't change
		return false
	}
	if util.IsZeroBigInt(n.getOutValue()) {
		// end value is 0
		return false
	}
	return true
}

// requiresExistingAccountingUTXOAsInput returns if value change requires input in the transaction
func (n *nativeTokenBalance) requiresExistingAccountingUTXOAsInput() bool {
	if n.identicalInOut() {
		// value didn't change
		return false
	}
	return n.accountingInput != nil
}

func (n *nativeTokenBalance) getOutValue() *big.Int {
	return n.accountingOutput.NativeTokens[0].Amount
}

func (n *nativeTokenBalance) add(delta *big.Int) *nativeTokenBalance {
	amount := new(big.Int).Add(n.getOutValue(), delta)
	if amount.Sign() < 0 {
		panic(fmt.Errorf("(id: %s, delta: %d): %v",
			n.nativeTokenID, delta, vm.ErrNotEnoughNativeAssetBalance))
	}
	if amount.Cmp(util.MaxUint256) > 0 {
		panic(vm.ErrOverflow)
	}
	n.accountingOutput.NativeTokens[0].Amount = amount
	return n
}

// updateMinSD uptates the resulting output to have the minimum SD
func (n *nativeTokenBalance) updateMinSD() {
	minSD := parameters.L1().Protocol.RentStructure.MinRent(n.accountingOutput)
	if minSD > n.accountingOutput.Amount {
		// sd for internal output can only ever increase
		n.accountingOutput.Amount = minSD
	}
}

func (n *nativeTokenBalance) identicalInOut() bool {
	switch {
	case n.accountingInput == n.accountingOutput:
		panic("identicalBasicOutputs: internal inconsistency 1")
	case n.accountingInput == nil || n.accountingOutput == nil:
		return false
	case !n.accountingInput.Ident().Equal(n.accountingOutput.Ident()):
		return false
	case n.accountingInput.Amount != n.accountingOutput.Amount:
		return false
	case !n.accountingInput.NativeTokens.Equal(n.accountingOutput.NativeTokens):
		return false
	case !n.accountingInput.Features.Equal(n.accountingOutput.Features):
		return false
	case len(n.accountingInput.NativeTokens) != 1:
		panic("identicalBasicOutputs: internal inconsistency 2")
	case len(n.accountingOutput.NativeTokens) != 1:
		panic("identicalBasicOutputs: internal inconsistency 3")
	case n.accountingInput.NativeTokens[0].ID != n.nativeTokenID:
		panic("identicalBasicOutputs: internal inconsistency 4")
	case n.accountingOutput.NativeTokens[0].ID != n.nativeTokenID:
		panic("identicalBasicOutputs: internal inconsistency 5")
	}
	return true
}

func cloneInternalBasicOutputOrNil(o *iotago.BasicOutput) *iotago.BasicOutput {
	if o == nil {
		return nil
	}
	return o.Clone().(*iotago.BasicOutput)
}

func (txb *AnchorTransactionBuilder) newInternalTokenOutput(aliasID iotago.AliasID, nativeTokenID iotago.NativeTokenID) *iotago.BasicOutput {
	out := &iotago.BasicOutput{
		Amount: 0,
		NativeTokens: iotago.NativeTokens{{
			ID:     nativeTokenID,
			Amount: big.NewInt(0),
		}},
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: aliasID.ToAddress()},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: aliasID.ToAddress(),
			},
		},
	}
	return out
}

func (txb *AnchorTransactionBuilder) nativeTokenOutputsSorted() []*nativeTokenBalance {
	ret := make([]*nativeTokenBalance, 0, len(txb.balanceNativeTokens))
	for _, f := range txb.balanceNativeTokens {
		if !f.requiresExistingAccountingUTXOAsInput() && !f.producesAccountingOutput() {
			continue
		}
		ret = append(ret, f)
	}
	sort.Slice(ret, func(i, j int) bool {
		return bytes.Compare(ret[i].nativeTokenID[:], ret[j].nativeTokenID[:]) < 0
	})
	return ret
}

func (txb *AnchorTransactionBuilder) NativeTokenRecordsToBeUpdated() ([]iotago.NativeTokenID, []iotago.NativeTokenID) {
	toBeUpdated := make([]iotago.NativeTokenID, 0, len(txb.balanceNativeTokens))
	toBeRemoved := make([]iotago.NativeTokenID, 0, len(txb.balanceNativeTokens))
	for _, nt := range txb.nativeTokenOutputsSorted() {
		if nt.producesAccountingOutput() {
			toBeUpdated = append(toBeUpdated, nt.nativeTokenID)
		} else if nt.requiresExistingAccountingUTXOAsInput() {
			toBeRemoved = append(toBeRemoved, nt.nativeTokenID)
		}
	}
	return toBeUpdated, toBeRemoved
}

func (txb *AnchorTransactionBuilder) NativeTokenOutputsByTokenIDs(ids []iotago.NativeTokenID) map[iotago.NativeTokenID]*iotago.BasicOutput {
	ret := make(map[iotago.NativeTokenID]*iotago.BasicOutput)
	for _, id := range ids {
		ret[id] = txb.balanceNativeTokens[id].accountingOutput
	}
	return ret
}

// addNativeTokenBalanceDelta adds delta to the token balance. Use negative delta to subtract.
// The call may result in adding new token ID to the ledger or disappearing one
// This impacts storage deposit amount locked in the internal UTXOs which keep respective balances
// Returns delta of required storage deposit
func (txb *AnchorTransactionBuilder) addNativeTokenBalanceDelta(nativeTokenID iotago.NativeTokenID, delta *big.Int) int64 {
	if util.IsZeroBigInt(delta) {
		return 0
	}
	nt := txb.ensureNativeTokenBalance(nativeTokenID).add(delta)

	if nt.identicalInOut() {
		return 0
	}

	if util.IsZeroBigInt(nt.getOutValue()) {
		// 0 native tokens on the output side
		if nt.accountingInput == nil {
			// in this case the internar accounting output that would be created is not needed anymore, reiburse the SD
			return int64(nt.accountingOutput.Amount)
		}
		return int64(nt.accountingInput.Amount)
	}

	// update the SD in case the storage deposit has changed from the last time this output was used
	oldSD := nt.accountingOutput.Amount
	nt.updateMinSD()
	updatedSD := nt.accountingOutput.Amount

	return int64(oldSD) - int64(updatedSD)
}

// ensureNativeTokenBalance makes sure that cached output is in the builder
// if not, it asks for the in balance by calling the loader function
// Panics if the call results to exceeded limits
func (txb *AnchorTransactionBuilder) ensureNativeTokenBalance(nativeTokenID iotago.NativeTokenID) *nativeTokenBalance {
	if nativeTokenBalance, exists := txb.balanceNativeTokens[nativeTokenID]; exists {
		return nativeTokenBalance
	}

	basicOutputIn, outputID := txb.accountsView.NativeTokenOutput(nativeTokenID) // output will be nil if no such token id accounted yet
	if basicOutputIn != nil {
		if txb.InputsAreFull() {
			panic(vmexceptions.ErrInputLimitExceeded)
		}
		if txb.outputsAreFull() {
			panic(vmexceptions.ErrOutputLimitExceeded)
		}
	}

	var basicOutputOut *iotago.BasicOutput
	if basicOutputIn == nil {
		basicOutputOut = txb.newInternalTokenOutput(util.AliasIDFromAliasOutput(txb.anchorOutput, txb.anchorOutputID), nativeTokenID)
	} else {
		basicOutputOut = cloneInternalBasicOutputOrNil(basicOutputIn)
	}

	nativeTokenBalance := &nativeTokenBalance{
		nativeTokenID:     nativeTokenID,
		accountingInputID: outputID,
		accountingInput:   basicOutputIn,
		accountingOutput:  basicOutputOut,
	}
	txb.balanceNativeTokens[nativeTokenID] = nativeTokenBalance
	return nativeTokenBalance
}
