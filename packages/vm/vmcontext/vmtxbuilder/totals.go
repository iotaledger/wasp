package vmtxbuilder

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
)

type TransactionTotals struct {
	// does not include internal storage deposits
	TotalBaseTokensInL2Accounts uint64
	// internal storage deposit
	TotalBaseTokensInStorageDeposit uint64
	// balances of native tokens (in all inputs/outputs). In the tx builder only loaded those which are needed
	NativeTokenBalances map[iotago.NativeTokenID]*big.Int
	// token supplies in foundries
	TokenCirculatingSupplies map[iotago.NativeTokenID]*big.Int
	// base tokens sent out by the transaction
	SentOutBaseTokens uint64
	// Sent out native tokens by the transaction
	SentOutTokenBalances map[iotago.NativeTokenID]*big.Int
}

func (txb *AnchorTransactionBuilder) sumNativeTokens(totals *TransactionTotals, filter func(ntb *nativeTokenBalance) *big.Int) {
	for id, ntb := range txb.balanceNativeTokens {
		value := filter(ntb)
		if value == nil {
			continue
		}
		s, ok := totals.NativeTokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, value)
		totals.NativeTokenBalances[id] = s
		// sum up storage deposit in inputs of internal UTXOs
		totals.TotalBaseTokensInStorageDeposit += txb.storageDepositAssumption.NativeTokenOutput
	}
}

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() *TransactionTotals {
	ret := &TransactionTotals{
		NativeTokenBalances:             make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies:        make(map[iotago.NativeTokenID]*big.Int),
		TotalBaseTokensInL2Accounts:     txb.anchorOutput.Deposit() - txb.storageDepositAssumption.AnchorOutput,
		TotalBaseTokensInStorageDeposit: txb.storageDepositAssumption.AnchorOutput,
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
		a := out.FungibleTokens()
		ret.TotalBaseTokensInL2Accounts += a.BaseTokens
		for _, nativeToken := range a.NativeTokens {
			s, ok := ret.NativeTokenBalances[nativeToken.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nativeToken.Amount)
			ret.NativeTokenBalances[nativeToken.ID] = s
		}
	}
	for _, f := range txb.invokedFoundries {
		if f.requiresInput() {
			ret.TotalBaseTokensInStorageDeposit += f.in.Amount
			simpleTokenScheme := util.MustTokenScheme(f.in.TokenScheme)
			ret.TokenCirculatingSupplies[f.in.MustNativeTokenID()] = new(big.Int).
				Sub(simpleTokenScheme.MintedTokens, simpleTokenScheme.MeltedTokens)
		}
	}

	for _, nft := range txb.nftsIncluded {
		if !isc.IsEmptyOutputID(nft.outputID) {
			ret.TotalBaseTokensInStorageDeposit += nft.in.Amount
		}
	}

	return ret
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() *TransactionTotals {
	ret := &TransactionTotals{
		NativeTokenBalances:             make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies:        make(map[iotago.NativeTokenID]*big.Int),
		TotalBaseTokensInL2Accounts:     txb.totalBaseTokensInL2Accounts,
		TotalBaseTokensInStorageDeposit: txb.storageDepositAssumption.AnchorOutput,
		SentOutBaseTokens:               0,
		SentOutTokenBalances:            make(map[iotago.NativeTokenID]*big.Int),
	}
	// sum over native tokens which produce outputs
	txb.sumNativeTokens(ret, func(ntb *nativeTokenBalance) *big.Int {
		if ntb.producesOutput() {
			return ntb.getOutValue()
		}
		return nil
	})
	for _, f := range txb.invokedFoundries {
		if !f.producesOutput() {
			continue
		}
		ret.TotalBaseTokensInStorageDeposit += f.out.Amount
		id := f.out.MustNativeTokenID()
		ret.TokenCirculatingSupplies[id] = big.NewInt(0)
		simpleTokenScheme := util.MustTokenScheme(f.out.TokenScheme)
		ret.TokenCirculatingSupplies[id].Sub(simpleTokenScheme.MintedTokens, simpleTokenScheme.MeltedTokens)
	}
	for _, o := range txb.postedOutputs {
		assets := transaction.AssetsFromOutput(o)
		ret.SentOutBaseTokens += assets.BaseTokens
		for _, nativeToken := range assets.NativeTokens {
			s, ok := ret.SentOutTokenBalances[nativeToken.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nativeToken.Amount)
			ret.SentOutTokenBalances[nativeToken.ID] = s
		}
	}
	for _, nft := range txb.nftsIncluded {
		if !nft.sentOutside {
			ret.TotalBaseTokensInStorageDeposit += nft.out.Amount
		}
	}
	return ret
}

// Totals check consistency. If input total equals with output totals, returns (base tokens total, native token totals, true)
// Otherwise returns (0, nil, false)
func (txb *AnchorTransactionBuilder) Totals() (*TransactionTotals, *TransactionTotals, error) {
	totalsIN := txb.sumInputs()
	totalsOUT := txb.sumOutputs()
	err := totalsIN.BalancedWith(totalsOUT)
	return totalsIN, totalsOUT, err
}

// TotalBaseTokensInOutputs returns (a) total base tokens owned by SCs and (b) total base tokens locked as storage deposit
func (txb *AnchorTransactionBuilder) TotalBaseTokensInOutputs() (uint64, uint64) {
	totals := txb.sumOutputs()
	return totals.TotalBaseTokensInL2Accounts, totals.TotalBaseTokensInStorageDeposit
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

var DebugTxBuilder = true

func (txb *AnchorTransactionBuilder) MustBalanced(checkpoint string) {
	if !DebugTxBuilder {
		return
	}
	ins, outs, err := txb.Totals()
	if err != nil {
		fmt.Printf("================= MustBalanced [%s]: %v \ninTotals: %v\noutTotals: %v\n", err, checkpoint, ins, outs)
		panic(fmt.Errorf("[%s] %v: %v ", checkpoint, vm.ErrFatalTxBuilderNotBalanced, err))
	}
}

func (txb *AnchorTransactionBuilder) AssertConsistentWithL2Totals(l2Totals *isc.FungibleTokens, checkpoint string) {
	_, outTotal, err := txb.Totals()
	if err != nil {
		panic(fmt.Errorf("%v: %v", vm.ErrFatalTxBuilderNotBalanced, err))
	}
	if outTotal.TotalBaseTokensInL2Accounts != l2Totals.BaseTokens {
		panic(fmt.Errorf("'%s': base tokens L1 (%d) != base tokens L2 (%d): %v",
			checkpoint, outTotal.TotalBaseTokensInL2Accounts, l2Totals.BaseTokens, vm.ErrInconsistentL2LedgerWithL1TxBuilder))
	}
	for _, nativeToken := range l2Totals.NativeTokens {
		b1, ok := outTotal.NativeTokenBalances[nativeToken.ID]
		if !ok {
			// checking only those which are in the tx builder
			continue
		}
		if nativeToken.Amount.Cmp(b1) != 0 {
			panic(fmt.Errorf("token %s L1 (%d) != L2 (%d): %v",
				nativeToken.ID.String(), nativeToken.Amount, b1, vm.ErrInconsistentL2LedgerWithL1TxBuilder))
		}
	}
}

func (t *TransactionTotals) BalancedWith(another *TransactionTotals) error {
	tIn := t.TotalBaseTokensInL2Accounts + t.TotalBaseTokensInStorageDeposit
	tOut := another.TotalBaseTokensInL2Accounts + another.TotalBaseTokensInStorageDeposit + another.SentOutBaseTokens
	if tIn != tOut {
		msgIn := fmt.Sprintf("in.TotalBaseTokensInL2Accounts: %d\n+ in.TotalBaseTokensInStorageDeposit: %d\n (%d)",
			t.TotalBaseTokensInL2Accounts, t.TotalBaseTokensInStorageDeposit, tIn)
		msgOut := fmt.Sprintf("out.TotalBaseTokensInL2Accounts: %d\n+ out.TotalBaseTokensInStorageDeposit: %d\n+ out.SentOutBaseToken: %d\n (%d)",
			another.TotalBaseTokensInL2Accounts, another.TotalBaseTokensInStorageDeposit, another.SentOutBaseTokens, tOut)
		return fmt.Errorf("%v:\n %s\n    !=\n%s", vm.ErrFatalTxBuilderNotBalanced, msgIn, msgOut)
	}
	nativeTokenIDs := make(map[iotago.NativeTokenID]bool)
	for id := range t.TokenCirculatingSupplies {
		nativeTokenIDs[id] = true
	}
	for id := range another.TokenCirculatingSupplies {
		nativeTokenIDs[id] = true
	}
	for id := range t.NativeTokenBalances {
		nativeTokenIDs[id] = true
	}
	for id := range t.SentOutTokenBalances {
		nativeTokenIDs[id] = true
	}

	tokenSupplyDeltas := make(map[iotago.NativeTokenID]*big.Int)
	for nativeTokenID := range nativeTokenIDs {
		inSupply, ok := t.TokenCirculatingSupplies[nativeTokenID]
		if !ok {
			inSupply = big.NewInt(0)
		}
		outSupply, ok := another.TokenCirculatingSupplies[nativeTokenID]
		if !ok {
			outSupply = big.NewInt(0)
		}
		tokenSupplyDeltas[nativeTokenID] = big.NewInt(0).Sub(outSupply, inSupply)
	}
	for nativeTokenIDs, delta := range tokenSupplyDeltas {
		begin, ok := t.NativeTokenBalances[nativeTokenIDs]
		if !ok {
			begin = big.NewInt(0)
		} else {
			begin = new(big.Int).Set(begin) // clone
		}
		end, ok := another.NativeTokenBalances[nativeTokenIDs]
		if !ok {
			end = big.NewInt(0)
		} else {
			end = new(big.Int).Set(end) // clone
		}
		sent, ok := another.SentOutTokenBalances[nativeTokenIDs]
		if !ok {
			sent = big.NewInt(0)
		} else {
			sent = new(big.Int).Set(sent) // clone
		}

		end.Add(end, sent)
		begin.Add(begin, delta)
		if begin.Cmp(end) != 0 {
			return fmt.Errorf("%v: token %s not balanced: in (%d) != out (%d)", vm.ErrFatalTxBuilderNotBalanced, nativeTokenIDs, begin, end)
		}
	}
	return nil
}
