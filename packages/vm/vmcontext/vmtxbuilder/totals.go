package vmtxbuilder

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
)

type TransactionTotals struct {
	// includes also internal dust deposits
	totalIotas uint64
	// internal dust deposit
	internalDustDeposit uint64
	// balances of native tokens
	tokenBalances map[iotago.NativeTokenID]*big.Int
}

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() (ret TransactionTotals) {
	ret.tokenBalances = make(map[iotago.NativeTokenID]*big.Int)
	ret.totalIotas = txb.anchorOutput.Amount
	sumIotasInternalDustDeposit := uint64(0)

	// sum up all initial values of internal accounts
	for id, ntb := range txb.balanceNativeTokens {
		if !ntb.requiresInput() {
			continue
		}
		sumIotasInternalDustDeposit += txb.vByteCostOfNativeTokenBalance()
		s, ok := ret.tokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, ntb.initial)
		ret.tokenBalances[id] = s
		// sum up dust deposit in inputs of internal UTXOs
		ret.totalIotas += txb.vByteCostOfNativeTokenBalance()
		ret.internalDustDeposit += txb.vByteCostOfNativeTokenBalance()
	}
	// sum up all explicitly consumed outputs, except anchor output
	for _, out := range txb.consumed {
		a := out.Assets()
		ret.totalIotas += a.Iotas
		for _, nt := range a.Tokens {
			s, ok := ret.tokenBalances[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			ret.tokenBalances[nt.ID] = s
		}
	}
	return
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() (ret TransactionTotals) {
	ret.tokenBalances = make(map[iotago.NativeTokenID]*big.Int)
	ret.totalIotas = txb.currentBalanceIotasOnAnchor

	// sum up all initial values of internal accounts
	for id, ntb := range txb.balanceNativeTokens {
		if !ntb.producesOutput() {
			continue
		}
		s, ok := ret.tokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, ntb.balance)
		ret.tokenBalances[id] = s
		// sum up dust deposit in outputs of internal UTXOs
		ret.totalIotas += txb.vByteCostOfNativeTokenBalance()
		ret.internalDustDeposit += txb.vByteCostOfNativeTokenBalance()
	}
	for _, o := range txb.postedOutputs {
		assets := assetsFromOutput(o)
		ret.totalIotas += assets.Iotas
		for _, nt := range assets.Tokens {
			s, ok := ret.tokenBalances[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			ret.tokenBalances[nt.ID] = s
		}
	}
	return
}

// Totals check consistency. If input total equals with output totals, returns (iota total, native token totals, true)
// Otherwise returns (0, nil, false)
func (txb *AnchorTransactionBuilder) Totals() (TransactionTotals, TransactionTotals, bool) {
	totalsIN := txb.sumInputs()
	totalsOUT := txb.sumOutputs()

	if totalsIN.totalIotas != totalsOUT.totalIotas {
		return TransactionTotals{}, TransactionTotals{}, false
	}
	if len(totalsIN.tokenBalances) != len(totalsOUT.tokenBalances) {
		return TransactionTotals{}, TransactionTotals{}, false
	}
	for id, bIN := range totalsIN.tokenBalances {
		bOUT, ok := totalsOUT.tokenBalances[id]
		if !ok {
			return TransactionTotals{}, TransactionTotals{}, false
		}
		if bIN.Cmp(bOUT) != 0 {
			return TransactionTotals{}, TransactionTotals{}, false
		}
	}
	return totalsIN, totalsOUT, true
}
