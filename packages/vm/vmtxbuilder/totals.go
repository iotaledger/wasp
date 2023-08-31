package vmtxbuilder

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
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

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() *TransactionTotals {
	anchorInputSD := parameters.L1().Protocol.RentStructure.MinRent(txb.anchorOutput)
	totals := &TransactionTotals{
		NativeTokenBalances:             make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies:        make(map[iotago.NativeTokenID]*big.Int),
		TotalBaseTokensInL2Accounts:     txb.anchorOutput.Deposit() - anchorInputSD,
		TotalBaseTokensInStorageDeposit: anchorInputSD,
	}
	// sum over native tokens which require inputs
	for id, ntb := range txb.balanceNativeTokens {
		if !ntb.requiresExistingAccountingUTXOAsInput() {
			continue
		}
		s, ok := totals.NativeTokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, ntb.accountingInput.NativeTokens[0].Amount)
		totals.NativeTokenBalances[id] = s
		// sum up storage deposit in inputs of internal UTXOs
		totals.TotalBaseTokensInStorageDeposit += ntb.accountingInput.Amount
	}
	// sum up all explicitly consumed outputs, except anchor output
	for _, out := range txb.consumed {
		a := out.Assets()
		totals.TotalBaseTokensInL2Accounts += a.BaseTokens
		for _, nativeToken := range a.NativeTokens {
			s, ok := totals.NativeTokenBalances[nativeToken.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nativeToken.Amount)
			totals.NativeTokenBalances[nativeToken.ID] = s
		}
	}
	for _, f := range txb.invokedFoundries {
		if f.requiresExistingAccountingUTXOAsInput() {
			totals.TotalBaseTokensInStorageDeposit += f.accountingInput.Amount
			simpleTokenScheme := util.MustTokenScheme(f.accountingInput.TokenScheme)
			totals.TokenCirculatingSupplies[f.accountingInput.MustNativeTokenID()] = new(big.Int).
				Sub(simpleTokenScheme.MintedTokens, simpleTokenScheme.MeltedTokens)
		}
	}

	for _, nft := range txb.nftsIncluded {
		if !isc.IsEmptyOutputID(nft.accountingInputID) {
			totals.TotalBaseTokensInStorageDeposit += nft.accountingInput.Amount
		}
	}

	return totals
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() *TransactionTotals {
	anchorOutputSD := parameters.L1().Protocol.RentStructure.MinRent(txb.resultAnchorOutput)

	totals := &TransactionTotals{
		NativeTokenBalances:             make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies:        make(map[iotago.NativeTokenID]*big.Int),
		TotalBaseTokensInL2Accounts:     txb.resultAnchorOutput.Amount - anchorOutputSD,
		TotalBaseTokensInStorageDeposit: anchorOutputSD,
		SentOutBaseTokens:               0,
		SentOutTokenBalances:            make(map[iotago.NativeTokenID]*big.Int),
	}
	// sum over native tokens which produce outputs
	for id, ntb := range txb.balanceNativeTokens {
		if !ntb.producesAccountingOutput() {
			continue
		}
		s, ok := totals.NativeTokenBalances[id]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, ntb.getOutValue())
		totals.NativeTokenBalances[id] = s
		// sum up storage deposit in inputs of internal UTXOs
		totals.TotalBaseTokensInStorageDeposit += ntb.accountingOutput.Amount
	}
	for _, f := range txb.invokedFoundries {
		if !f.producesAccountingOutput() {
			continue
		}
		totals.TotalBaseTokensInStorageDeposit += f.accountingOutput.Amount
		id := f.accountingOutput.MustNativeTokenID()
		totals.TokenCirculatingSupplies[id] = big.NewInt(0)
		simpleTokenScheme := util.MustTokenScheme(f.accountingOutput.TokenScheme)
		totals.TokenCirculatingSupplies[id].Sub(simpleTokenScheme.MintedTokens, simpleTokenScheme.MeltedTokens)
	}
	for _, o := range txb.postedOutputs {
		assets := transaction.AssetsFromOutput(o)
		totals.SentOutBaseTokens += assets.BaseTokens
		for _, nativeToken := range assets.NativeTokens {
			s, ok := totals.SentOutTokenBalances[nativeToken.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nativeToken.Amount)
			totals.SentOutTokenBalances[nativeToken.ID] = s
		}
	}
	for _, nft := range txb.nftsIncluded {
		if !nft.sentOutside {
			totals.TotalBaseTokensInStorageDeposit += nft.resultingOutput.Amount
		}
	}
	for _, nft := range txb.nftsMinted {
		totals.SentOutBaseTokens += nft.Deposit()
	}
	return totals
}

// TotalBaseTokensInOutputs returns (a) total base tokens owned by SCs and (b) total base tokens locked as storage deposit
func (txb *AnchorTransactionBuilder) TotalBaseTokensInOutputs() (uint64, uint64) {
	totals := txb.sumOutputs()
	return totals.TotalBaseTokensInL2Accounts, totals.TotalBaseTokensInStorageDeposit
}

// MustBalanced asserts that the txb is balanced (intputs/outputs) and is consistent with L2
// IMPORTANT: must be executed after `BuildTransactionEssence`, so that txb.resultAnchorOutput is calculated
func (txb *AnchorTransactionBuilder) MustBalanced() {
	// assert inputs/outpus are balanced
	totalsIN := txb.sumInputs()
	totalsOUT := txb.sumOutputs()

	if err := totalsIN.BalancedWith(totalsOUT); err != nil {
		fmt.Printf("================= MustBalanced: %v \ninTotals: %v\noutTotals: %v\n", err, totalsIN, totalsOUT)
		panic(fmt.Errorf("%v: %w ", vm.ErrFatalTxBuilderNotBalanced, err))
	}

	// assert the txbuilder is consistent with L2 accounting
	l2Totals := txb.accountsView.TotalFungibleTokens()
	if totalsOUT.TotalBaseTokensInL2Accounts != l2Totals.BaseTokens {
		panic(fmt.Errorf("base tokens L1 (%d) != base tokens L2 (%d): %v",
			totalsOUT.TotalBaseTokensInL2Accounts, l2Totals.BaseTokens, vm.ErrInconsistentL2LedgerWithL1TxBuilder))
	}
	for _, nativeToken := range l2Totals.NativeTokens {
		b1, ok := totalsOUT.NativeTokenBalances[nativeToken.ID]
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
