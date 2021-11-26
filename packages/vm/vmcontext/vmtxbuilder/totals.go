package vmtxbuilder

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
)

type TransactionTotals struct {
	// includes also internal dust deposits
	TotalIotas uint64
	// internal dust deposit
	InternalDustDeposit uint64
	// balances of native tokens
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
		totals.TotalIotas += txb.vByteCostOfNativeTokenBalance()
		totals.InternalDustDeposit += txb.vByteCostOfNativeTokenBalance()
	}
}

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() *TransactionTotals {
	ret := &TransactionTotals{
		TokenBalances:       make(map[iotago.NativeTokenID]*big.Int),
		TotalIotas:          txb.anchorOutput.Amount,
		InternalDustDeposit: 0,
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
		ret.TotalIotas += a.Iotas
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
		TokenBalances:       make(map[iotago.NativeTokenID]*big.Int),
		TotalIotas:          txb.currentBalanceIotasOnAnchor,
		InternalDustDeposit: 0,
	}
	// sum over native tokens which produce outputs
	txb.sumNativeTokens(ret, func(ntb *nativeTokenBalance) *big.Int {
		if ntb.producesOutput() {
			return ntb.balance
		}
		return nil
	})

	for _, o := range txb.postedOutputs {
		assets := assetsFromOutput(o)
		ret.TotalIotas += assets.Iotas
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

	if totalsIN.TotalIotas != totalsOUT.TotalIotas {
		return nil, nil, false
	}
	if len(totalsIN.TokenBalances) != len(totalsOUT.TokenBalances) {
		return nil, nil, false
	}
	for id, bIN := range totalsIN.TokenBalances {
		bOUT, ok := totalsOUT.TokenBalances[id]
		if !ok {
			return nil, nil, false
		}
		if bIN.Cmp(bOUT) != 0 {
			return nil, nil, false
		}
	}
	return totalsIN, totalsOUT, true
}
