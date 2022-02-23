package vmtxbuilder

import (
	"bytes"
	"math/big"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
)

// nativeTokenBalance represents on-chain account of the specific native token
type nativeTokenBalance struct {
	tokenID            iotago.NativeTokenID
	input              iotago.UTXOInput // if in != nil
	dustDepositCharged bool
	in                 *iotago.BasicOutput // if nil it means output does not exist, this is new account for the token_id
	out                *iotago.BasicOutput // current balance of the token_id on the chain
}

func (n *nativeTokenBalance) clone() *nativeTokenBalance {
	return &nativeTokenBalance{
		tokenID:            n.tokenID,
		input:              n.input,
		dustDepositCharged: n.dustDepositCharged,
		in:                 cloneInternalBasicOutputOrNil(n.in),
		out:                cloneInternalBasicOutputOrNil(n.out),
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
	case !n.in.Blocks.Equal(n.out.Blocks):
		return false
	case len(n.in.NativeTokens) != 1:
		panic("identicalBasicOutputs: internal inconsistency 2")
	case len(n.out.NativeTokens) != 1:
		panic("identicalBasicOutputs: internal inconsistency 3")
	case n.in.NativeTokens[0].ID != n.tokenID:
		panic("identicalBasicOutputs: internal inconsistency 4")
	case n.out.NativeTokens[0].ID != n.tokenID:
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
		Amount: txb.dustDepositAssumption.NativeTokenOutput,
		NativeTokens: iotago.NativeTokens{{
			ID:     nativeTokenID,
			Amount: big.NewInt(0),
		}},
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: aliasID.ToAddress()},
		},
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
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
		return bytes.Compare(ret[i].tokenID[:], ret[j].tokenID[:]) < 0
	})
	return ret
}

func (txb *AnchorTransactionBuilder) NativeTokenRecordsToBeUpdated() ([]iotago.NativeTokenID, []iotago.NativeTokenID) {
	toBeUpdated := make([]iotago.NativeTokenID, 0, len(txb.balanceNativeTokens))
	toBeRemoved := make([]iotago.NativeTokenID, 0, len(txb.balanceNativeTokens))
	for _, nt := range txb.nativeTokenOutputsSorted() {
		if nt.producesOutput() {
			toBeUpdated = append(toBeUpdated, nt.tokenID)
		} else if nt.requiresInput() {
			toBeRemoved = append(toBeRemoved, nt.tokenID)
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
