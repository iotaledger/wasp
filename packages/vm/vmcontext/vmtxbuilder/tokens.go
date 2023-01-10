package vmtxbuilder

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

// nativeTokenBalance represents on-chain account of the specific native token
type nativeTokenBalance struct {
	nativeTokenID         iotago.NativeTokenID
	outputID              iotago.OutputID // if in != nil, otherwise zeroOutputID
	storageDepositCharged bool
	in                    *iotago.BasicOutput // if nil it means output does not exist, this is new account for the token_id
	out                   *iotago.BasicOutput // current balance of the token_id on the chain
}

func (n *nativeTokenBalance) Clone() *nativeTokenBalance {
	nativeTokenID := iotago.NativeTokenID{}
	copy(nativeTokenID[:], n.nativeTokenID[:])

	outputID := iotago.OutputID{}
	copy(outputID[:], n.outputID[:])

	return &nativeTokenBalance{
		nativeTokenID:         nativeTokenID,
		outputID:              outputID,
		storageDepositCharged: n.storageDepositCharged,
		in:                    cloneInternalBasicOutputOrNil(n.in),
		out:                   cloneInternalBasicOutputOrNil(n.out),
	}
}

// producesOutput if value update produces UTXO of the corresponding total native token balance
func (n *nativeTokenBalance) producesOutput() bool {
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

// requiresInput returns if value change requires input in the transaction
func (n *nativeTokenBalance) requiresInput() bool {
	if n.identicalInOut() {
		// value didn't change
		return false
	}
	if n.in == nil {
		// there's no input
		return false
	}
	return true
}

func (n *nativeTokenBalance) getOutValue() *big.Int {
	return n.out.NativeTokens[0].Amount
}

func (n *nativeTokenBalance) setOutValue(v *big.Int) {
	n.out.NativeTokens[0].Amount = v
}

func (n *nativeTokenBalance) identicalInOut() bool {
	switch {
	case n.in == n.out:
		panic("identicalBasicOutputs: internal inconsistency 1")
	case n.in == nil || n.out == nil:
		return false
	case !n.in.Ident().Equal(n.out.Ident()):
		return false
	case n.in.Amount != n.out.Amount:
		return false
	case !n.in.NativeTokens.Equal(n.out.NativeTokens):
		return false
	case !n.in.Features.Equal(n.out.Features):
		return false
	case len(n.in.NativeTokens) != 1:
		panic("identicalBasicOutputs: internal inconsistency 2")
	case len(n.out.NativeTokens) != 1:
		panic("identicalBasicOutputs: internal inconsistency 3")
	case n.in.NativeTokens[0].ID != n.nativeTokenID:
		panic("identicalBasicOutputs: internal inconsistency 4")
	case n.out.NativeTokens[0].ID != n.nativeTokenID:
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
	return &iotago.BasicOutput{
		Amount: txb.storageDepositAssumption.NativeTokenOutput,
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
}

func (txb *AnchorTransactionBuilder) nativeTokenOutputsSorted() []*nativeTokenBalance {
	ret := make([]*nativeTokenBalance, 0, len(txb.balanceNativeTokens))
	for _, f := range txb.balanceNativeTokens {
		if !f.requiresInput() && !f.producesOutput() {
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
		if nt.producesOutput() {
			toBeUpdated = append(toBeUpdated, nt.nativeTokenID)
		} else if nt.requiresInput() {
			toBeRemoved = append(toBeRemoved, nt.nativeTokenID)
		}
	}
	return toBeUpdated, toBeRemoved
}

func (txb *AnchorTransactionBuilder) NativeTokenOutputsByTokenIDs(ids []iotago.NativeTokenID) map[iotago.NativeTokenID]*iotago.BasicOutput {
	ret := make(map[iotago.NativeTokenID]*iotago.BasicOutput)
	for _, id := range ids {
		ret[id] = txb.balanceNativeTokens[id].out
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
	nt := txb.ensureNativeTokenBalance(nativeTokenID)
	tmp := new(big.Int).Add(nt.getOutValue(), delta)
	if tmp.Sign() < 0 {
		panic(xerrors.Errorf("addNativeTokenBalanceDelta (id: %s, delta: %d): %v",
			nativeTokenID, delta, vm.ErrNotEnoughNativeAssetBalance))
	}
	if tmp.Cmp(abi.MaxUint256) > 0 {
		panic(xerrors.Errorf("addNativeTokenBalanceDelta: %v", vm.ErrOverflow))
	}
	nt.setOutValue(tmp)
	switch {
	case nt.identicalInOut():
		return 0
	case nt.storageDepositCharged && !nt.producesOutput():
		// this is an old token in the on-chain ledger. Now it disappears and storage deposit
		// is released and delta of anchor is positive
		nt.storageDepositCharged = false
		txb.addDeltaBaseTokensToTotal(txb.storageDepositAssumption.NativeTokenOutput)
		return int64(txb.storageDepositAssumption.NativeTokenOutput)
	case !nt.storageDepositCharged && nt.producesOutput():
		// this is a new token in the on-chain ledger
		// There's a need for additional storage deposit on the respective UTXO, so delta for the anchor is negative
		nt.storageDepositCharged = true
		if txb.storageDepositAssumption.NativeTokenOutput > txb.totalBaseTokensInL2Accounts {
			panic(vmexceptions.ErrNotEnoughFundsForInternalStorageDeposit)
		}
		txb.subDeltaBaseTokensFromTotal(txb.storageDepositAssumption.NativeTokenOutput)
		return -int64(txb.storageDepositAssumption.NativeTokenOutput)
	}
	return 0
}

// ensureNativeTokenBalance makes sure that cached output is in the builder
// if not, it asks for the in balance by calling the loader function
// Panics if the call results to exceeded limits
func (txb *AnchorTransactionBuilder) ensureNativeTokenBalance(nativeTokenID iotago.NativeTokenID) *nativeTokenBalance {
	if nativeTokenBalance, exists := txb.balanceNativeTokens[nativeTokenID]; exists {
		return nativeTokenBalance
	}

	basicOutputIn, outputID := txb.loadNativeTokenOutputFunc(nativeTokenID) // output will be nil if no such token id accounted yet
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
		basicOutputOut = txb.newInternalTokenOutput(txb.anchorOutput.AliasID, nativeTokenID)
	} else {
		basicOutputOut = cloneInternalBasicOutputOrNil(basicOutputIn)
	}

	nativeTokenBalance := &nativeTokenBalance{
		nativeTokenID:         basicOutputOut.NativeTokens[0].ID,
		outputID:              outputID,
		in:                    basicOutputIn,
		out:                   basicOutputOut,
		storageDepositCharged: basicOutputIn != nil,
	}
	txb.balanceNativeTokens[nativeTokenID] = nativeTokenBalance
	return nativeTokenBalance
}
