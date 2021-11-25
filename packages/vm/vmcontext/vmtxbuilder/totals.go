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

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() (ret TransactionTotals) {
	ret.TokenBalances = make(map[iotago.NativeTokenID]*big.Int)
	ret.TotalIotas = txb.anchorOutput.Amount
	sumIotasInternalDustDeposit := uint64(0)

	// sum up all initial values of internal accounts
	for id, ntb := range txb.balanceNativeTokens {
		if !ntb.requiresInput() {
			continue
		}
		sumIotasInternalDustDeposit += txb.vByteCostOfNativeTokenBalance()
		s, ok := ret.TokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, ntb.initial)
		ret.TokenBalances[id] = s
		// sum up dust deposit in inputs of internal UTXOs
		ret.TotalIotas += txb.vByteCostOfNativeTokenBalance()
		ret.InternalDustDeposit += txb.vByteCostOfNativeTokenBalance()
	}
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
	return
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() (ret TransactionTotals) {
	ret.TokenBalances = make(map[iotago.NativeTokenID]*big.Int)
	ret.TotalIotas = txb.currentBalanceIotasOnAnchor

	// sum up all initial values of internal accounts
	for id, ntb := range txb.balanceNativeTokens {
		if !ntb.producesOutput() {
			continue
		}
		s, ok := ret.TokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, ntb.balance)
		ret.TokenBalances[id] = s
		// sum up dust deposit in outputs of internal UTXOs
		ret.TotalIotas += txb.vByteCostOfNativeTokenBalance()
		ret.InternalDustDeposit += txb.vByteCostOfNativeTokenBalance()
	}
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
	return
}

// Totals check consistency. If input total equals with output totals, returns (iota total, native token totals, true)
// Otherwise returns (0, nil, false)
func (txb *AnchorTransactionBuilder) Totals() (TransactionTotals, TransactionTotals, bool) {
	totalsIN := txb.sumInputs()
	totalsOUT := txb.sumOutputs()

	if totalsIN.TotalIotas != totalsOUT.TotalIotas {
		return TransactionTotals{}, TransactionTotals{}, false
	}
	if len(totalsIN.TokenBalances) != len(totalsOUT.TokenBalances) {
		return TransactionTotals{}, TransactionTotals{}, false
	}
	for id, bIN := range totalsIN.TokenBalances {
		bOUT, ok := totalsOUT.TokenBalances[id]
		if !ok {
			return TransactionTotals{}, TransactionTotals{}, false
		}
		if bIN.Cmp(bOUT) != 0 {
			return TransactionTotals{}, TransactionTotals{}, false
		}
	}
	return totalsIN, totalsOUT, true
}
