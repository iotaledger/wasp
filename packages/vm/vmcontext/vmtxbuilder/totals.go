package vmtxbuilder

import (
	"fmt"
	"math/big"

	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
)

type TransactionTotals struct {
	// does not include internal dust deposits
	TotalIotasOnChain uint64
	// internal dust deposit
	TotalIotasInDustDeposit uint64
	// balances of native tokens (in all inputs/outputs)
	TokenBalances map[iotago.NativeTokenID]*big.Int
	// token supplies in foundries
	TokenCirculatingSupplies map[iotago.NativeTokenID]*big.Int
	// sent out iotas
	SentOutIotas uint64
	// Sent out native tokens
	SentOutTokenBalances map[iotago.NativeTokenID]*big.Int
}

func (txb *AnchorTransactionBuilder) sumNativeTokens(totals *TransactionTotals, filter func(ntb *nativeTokenBalance) *big.Int) {
	for id, ntb := range txb.balanceNativeTokens {
		value := filter(ntb)
		if value == nil {
			continue
		}
		s, ok := totals.TokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, value)
		totals.TokenBalances[id] = s
		// sum up dust deposit in inputs of internal UTXOs
		totals.TotalIotasInDustDeposit += txb.dustDepositOnInternalTokenOutput
	}
}

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() *TransactionTotals {
	ret := &TransactionTotals{
		TokenBalances:            make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies: make(map[iotago.NativeTokenID]*big.Int),
		TotalIotasOnChain:        txb.anchorOutput.Deposit() - txb.dustDepositOnAnchor,
		TotalIotasInDustDeposit:  txb.dustDepositOnAnchor,
	}
	// sum over native tokens which require inputs
	txb.sumNativeTokens(ret, func(ntb *nativeTokenBalance) *big.Int {
		if ntb.requiresInput() {
			return ntb.in.NativeTokens[0].Amount
		}
		return nil
	})
	// sum up all explicitly consumed outputs, except anchor output
	for _, out := range txb.consumed {
		a := out.Assets()
		ret.TotalIotasOnChain += a.Iotas
		for _, nt := range a.Tokens {
			s, ok := ret.TokenBalances[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			ret.TokenBalances[nt.ID] = s
		}
	}
	for _, f := range txb.invokedFoundries {
		if f.requiresInput() {
			ret.TotalIotasInDustDeposit += f.in.Amount
			ret.TokenCirculatingSupplies[f.in.MustNativeTokenID()] = new(big.Int).Set(f.in.CirculatingSupply)
		}
	}
	return ret
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() *TransactionTotals {
	ret := &TransactionTotals{
		TokenBalances:            make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies: make(map[iotago.NativeTokenID]*big.Int),
		TotalIotasOnChain:        txb.totalIotasOnChain,
		TotalIotasInDustDeposit:  txb.dustDepositOnAnchor,
		SentOutIotas:             0,
		SentOutTokenBalances:     make(map[iotago.NativeTokenID]*big.Int),
	}
	// sum over native tokens which produce outputs
	txb.sumNativeTokens(ret, func(ntb *nativeTokenBalance) *big.Int {
		if ntb.producesOutput() {
			return ntb.getOutValue()
		}
		return nil
	})
	for _, f := range txb.invokedFoundries {
		if f.producesOutput() {
			ret.TotalIotasInDustDeposit += f.out.Amount
			id := f.out.MustNativeTokenID()
			ret.TokenCirculatingSupplies[id] = big.NewInt(0)
			ret.TokenCirculatingSupplies[id].Set(f.out.CirculatingSupply)
		}
	}
	for _, o := range txb.postedOutputs {
		assets := AssetsFromOutput(o)
		ret.TotalIotasOnChain += assets.Iotas
		for _, nt := range assets.Tokens {
			s, ok := ret.TokenBalances[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			ret.TokenBalances[nt.ID] = s

			ret.SentOutIotas += assets.Iotas
			s, ok = ret.SentOutTokenBalances[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			ret.SentOutTokenBalances[nt.ID] = s
		}
	}
	return ret
}

// Totals check consistency. If input total equals with output totals, returns (iota total, native token totals, true)
// Otherwise returns (0, nil, false)
func (txb *AnchorTransactionBuilder) Totals() (*TransactionTotals, *TransactionTotals, bool) {
	totalsIN := txb.sumInputs()
	totalsOUT := txb.sumOutputs()
	balanced := totalsIN.BalancedWith(totalsOUT)
	return totalsIN, totalsOUT, balanced
}

// InternalNativeTokenBalances returns internally maintained balances of native tokens in inputs and
func (txb *AnchorTransactionBuilder) InternalNativeTokenBalances() (map[iotago.NativeTokenID]*big.Int, map[iotago.NativeTokenID]*big.Int) {
	before := make(map[iotago.NativeTokenID]*big.Int)
	after := make(map[iotago.NativeTokenID]*big.Int)

	for id, ntb := range txb.balanceNativeTokens {
		if ntb.requiresInput() {
			before[id] = ntb.in.NativeTokens[0].Amount
		}
		if ntb.producesOutput() {
			after[id] = ntb.getOutValue()
		}
	}
	return before, after
}

func (t *TransactionTotals) BalancedWith(another *TransactionTotals) bool {
	if t.TotalIotasOnChain+t.TotalIotasInDustDeposit != another.TotalIotasOnChain+another.TotalIotasInDustDeposit {
		return false
	}
	tokenIDs := make(map[iotago.NativeTokenID]bool)
	for id := range t.TokenCirculatingSupplies {
		tokenIDs[id] = true
	}
	for id := range another.TokenCirculatingSupplies {
		tokenIDs[id] = true
	}
	for id := range t.TokenBalances {
		tokenIDs[id] = true
	}
	tokenSupplyDeltas := make(map[iotago.NativeTokenID]*big.Int)
	for id := range tokenIDs {
		inSupply, ok := t.TokenCirculatingSupplies[id]
		if !ok {
			inSupply = big.NewInt(0)
		}
		outSupply, ok := another.TokenCirculatingSupplies[id]
		if !ok {
			outSupply = big.NewInt(0)
		}
		tokenSupplyDeltas[id] = big.NewInt(0).Sub(outSupply, inSupply)
	}
	for id, delta := range tokenSupplyDeltas {
		begin, ok := t.TokenBalances[id]
		if !ok {
			begin = big.NewInt(0)
		}
		end, ok := another.TokenBalances[id]
		if !ok {
			end = big.NewInt(0)
		}
		begin.Add(begin, delta)
		if begin.Cmp(end) != 0 {
			return false
		}
	}
	return true
}

var DebugTxBuilder = func() bool { return true }() // trick linter

func (txb *AnchorTransactionBuilder) MustBalanced(checkpoint string) {
	if DebugTxBuilder {
		ins, outs, balanced := txb.Totals()
		if !balanced {
			fmt.Printf("================= MustBalanced [%s] \ninTotals: %v\noutTotals: %v\n", checkpoint, ins, outs)
			panic(xerrors.Errorf("internal: tx builder is not balanced [%s]", checkpoint))
		}
	}

}
