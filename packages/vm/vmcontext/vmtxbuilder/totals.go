package vmtxbuilder

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
)

type TransactionTotals struct {
	// does not include internal dust deposits
	TotalIotasOnChain uint64
	// internal dust deposit
	TotalIotasInDustDeposit uint64
	// balances of native tokens (in all inputs/outputs)
	TokenBalances map[iotago.NativeTokenID]*big.Int
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
		totals.TotalIotasInDustDeposit += txb.vByteCostOfNativeTokenBalance()
	}
}

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() *TransactionTotals {
	ret := &TransactionTotals{
		TokenBalances:           make(map[iotago.NativeTokenID]*big.Int),
		TotalIotasOnChain:       txb.TotalAvailableIotas(),
		TotalIotasInDustDeposit: txb.minDustDepositOnAnchor,
	}
	// sum over native tokens which require inputs
	txb.sumNativeTokens(ret, func(ntb *nativeTokenBalance) *big.Int {
		if ntb.requiresInput() {
			return ntb.initial
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
	return ret
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() *TransactionTotals {
	ret := &TransactionTotals{
		TokenBalances:           make(map[iotago.NativeTokenID]*big.Int),
		TotalIotasOnChain:       txb.currentBalanceIotasOnAnchor,
		TotalIotasInDustDeposit: 0,
	}
	// sum over native tokens which produce outputs
	txb.sumNativeTokens(ret, func(ntb *nativeTokenBalance) *big.Int {
		if ntb.producesOutput() {
			return ntb.balance
		}
		return nil
	})

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
			before[id] = ntb.initial
		}
		if ntb.producesOutput() {
			after[id] = ntb.balance
		}
	}
	return before, after
}

func (t *TransactionTotals) BalancedWith(another *TransactionTotals) bool {
	if t.TotalIotasOnChain+t.TotalIotasInDustDeposit != another.TotalIotasOnChain+another.TotalIotasInDustDeposit {
		return false
	}
	if len(t.TokenBalances) != len(another.TokenBalances) {
		return false
	}
	for id, bT := range t.TokenBalances {
		bAnother, ok := another.TokenBalances[id]
		if !ok {
			return false
		}
		if bT.Cmp(bAnother) != 0 {
			return false
		}
	}
	return true
}
