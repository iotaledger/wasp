package vmtxbuilder

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"golang.org/x/xerrors"
)

type TransactionTotals struct {
	// does not include internal dust deposits
	TotalIotasInL2Accounts uint64
	// internal dust deposit
	TotalIotasInDustDeposit uint64
	// balances of native tokens (in all inputs/outputs). In the tx builder only loaded those which are needed
	NativeTokenBalances map[iotago.NativeTokenID]*big.Int
	// token supplies in foundries
	TokenCirculatingSupplies map[iotago.NativeTokenID]*big.Int
	// sent out iotas by the transaction
	SentOutIotas uint64
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
		// sum up dust deposit in inputs of internal UTXOs
		totals.TotalIotasInDustDeposit += txb.dustDepositAssumption.NativeTokenOutput
	}
}

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() *TransactionTotals {
	ret := &TransactionTotals{
		NativeTokenBalances:      make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies: make(map[iotago.NativeTokenID]*big.Int),
		TotalIotasInL2Accounts:   txb.anchorOutput.Deposit() - txb.dustDepositAssumption.AnchorOutput,
		TotalIotasInDustDeposit:  txb.dustDepositAssumption.AnchorOutput,
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
		ret.TotalIotasInL2Accounts += a.Iotas
		for _, nt := range a.Tokens {
			s, ok := ret.NativeTokenBalances[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			ret.NativeTokenBalances[nt.ID] = s
		}
	}
	for _, f := range txb.invokedFoundries {
		if f.requiresInput() {
			ret.TotalIotasInDustDeposit += f.in.Amount
			ret.TokenCirculatingSupplies[f.in.MustNativeTokenID()] = new(big.Int).Set(f.in.MintedTokens)
		}
	}
	for _, nft := range txb.nftsIncluded {
		if nft.input != nil {
			ret.TotalIotasInDustDeposit += nft.in.Amount
		}
	}

	return ret
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() *TransactionTotals {
	ret := &TransactionTotals{
		NativeTokenBalances:      make(map[iotago.NativeTokenID]*big.Int),
		TokenCirculatingSupplies: make(map[iotago.NativeTokenID]*big.Int),
		TotalIotasInL2Accounts:   txb.totalIotasInL2Accounts,
		TotalIotasInDustDeposit:  txb.dustDepositAssumption.AnchorOutput,
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
			ret.TokenCirculatingSupplies[id].Set(f.out.MintedTokens)
		}
	}
	for _, o := range txb.postedOutputs {
		assets := transaction.AssetsFromOutput(o)
		ret.SentOutIotas += assets.Iotas
		for _, nt := range assets.Tokens {
			s, ok := ret.SentOutTokenBalances[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			ret.SentOutTokenBalances[nt.ID] = s
		}
	}
	for _, nft := range txb.nftsIncluded {
		if !nft.sentOutside {
			ret.TotalIotasInDustDeposit += nft.out.Amount
		}
	}
	return ret
}

// Totals check consistency. If input total equals with output totals, returns (iota total, native token totals, true)
// Otherwise returns (0, nil, false)
func (txb *AnchorTransactionBuilder) Totals() (*TransactionTotals, *TransactionTotals, error) {
	totalsIN := txb.sumInputs()
	totalsOUT := txb.sumOutputs()
	err := totalsIN.BalancedWith(totalsOUT)
	return totalsIN, totalsOUT, err
}

// TotalIotasInOutputs returns (a) total iotas owned by SCs and (b) total iotas locked as dust deposit
func (txb *AnchorTransactionBuilder) TotalIotasInOutputs() (uint64, uint64) {
	totals := txb.sumOutputs()
	return totals.TotalIotasInL2Accounts, totals.TotalIotasInDustDeposit
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
		panic(xerrors.Errorf("[%s] %v: %v ", checkpoint, vm.ErrFatalTxBuilderNotBalanced, err))
	}
}

func (txb *AnchorTransactionBuilder) AssertConsistentWithL2Totals(l2Totals *iscp.FungibleTokens, checkpoint string) {
	_, outTotal, err := txb.Totals()
	if err != nil {
		panic(xerrors.Errorf("%v: %v", vm.ErrFatalTxBuilderNotBalanced, err))
	}
	if outTotal.TotalIotasInL2Accounts != l2Totals.Iotas {
		panic(xerrors.Errorf("'%s': iotas L1 (%d) != iotas L2 (%d): %v",
			checkpoint, outTotal.TotalIotasInL2Accounts, l2Totals.Iotas, vm.ErrInconsistentL2LedgerWithL1TxBuilder))
	}
	for _, nt := range l2Totals.Tokens {
		b1, ok := outTotal.NativeTokenBalances[nt.ID]
		if !ok {
			// checking only those which are in the tx builder
			continue
		}
		if nt.Amount.Cmp(b1) != 0 {
			panic(xerrors.Errorf("token %s L1 (%d) != L2 (%d): %v",
				nt.ID.String(), nt.Amount, b1, vm.ErrInconsistentL2LedgerWithL1TxBuilder))
		}
	}
}

func (t *TransactionTotals) BalancedWith(another *TransactionTotals) error {
	tIn := t.TotalIotasInL2Accounts + t.TotalIotasInDustDeposit
	tOut := another.TotalIotasInL2Accounts + another.TotalIotasInDustDeposit + another.SentOutIotas
	if tIn != tOut {
		msgIn := fmt.Sprintf("in.TotalIotasInL2Accounts: %d\n+ in.TotalIotasInDustDeposit: %d\n (%d)",
			t.TotalIotasInL2Accounts, t.TotalIotasInDustDeposit, tIn)
		msgOut := fmt.Sprintf("out.TotalIotasInL2Accounts: %d\n+ out.TotalIotasInDustDeposit: %d\n+ out.SentOutIotas: %d\n (%d)",
			another.TotalIotasInL2Accounts, another.TotalIotasInDustDeposit, another.SentOutIotas, tOut)
		return xerrors.Errorf("%v:\n %s\n    !=\n%s", vm.ErrFatalTxBuilderNotBalanced, msgIn, msgOut)
	}
	tokenIDs := make(map[iotago.NativeTokenID]bool)
	for id := range t.TokenCirculatingSupplies {
		tokenIDs[id] = true
	}
	for id := range another.TokenCirculatingSupplies {
		tokenIDs[id] = true
	}
	for id := range t.NativeTokenBalances {
		tokenIDs[id] = true
	}
	for id := range t.SentOutTokenBalances {
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
		begin, ok := t.NativeTokenBalances[id]
		if !ok {
			begin = big.NewInt(0)
		} else {
			begin = new(big.Int).Set(begin) // clone
		}
		end, ok := another.NativeTokenBalances[id]
		if !ok {
			end = big.NewInt(0)
		} else {
			end = new(big.Int).Set(end) // clone
		}
		sent, ok := another.SentOutTokenBalances[id]
		if !ok {
			sent = big.NewInt(0)
		} else {
			sent = new(big.Int).Set(sent) // clone
		}

		end.Add(end, sent)
		begin.Add(begin, delta)
		if begin.Cmp(end) != 0 {
			return xerrors.Errorf("%v: token %s not balanced: in (%d) != out (%d)", vm.ErrFatalTxBuilderNotBalanced, id, begin, end)
		}
	}
	return nil
}
