package vm

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

// TransactionBuilder is a wrapper around basic sc transaction builder
// it provides generalized token accounts for the VM sandbox
type TransactionBuilder struct {
	sctxbuilder    *sctransaction.TransactionBuilder
	balances       map[valuetransaction.ID][]*balance.Balance
	scAddress      address.Address
	scColor        balance.Color
	inputBalances  map[balance.Color]int64
	outputBalances map[address.Address]map[balance.Color]int64
	total          int64
}

// NewTxBuilder
type TransactionBuilderParams struct {
	Balances   map[valuetransaction.ID][]*balance.Balance
	OwnColor   balance.Color
	OwnAddress address.Address
}

func NewTxBuilder(par TransactionBuilderParams) (*TransactionBuilder, error) {
	ret := &TransactionBuilder{
		sctxbuilder:    sctransaction.NewTransactionBuilder(),
		balances:       par.Balances,
		inputBalances:  make(map[balance.Color]int64),
		outputBalances: make(map[address.Address]map[balance.Color]int64),
		scAddress:      par.OwnAddress,
		scColor:        par.OwnColor,
	}
	for _, lst := range par.Balances {
		for _, b := range lst {
			s, ok := ret.inputBalances[b.Color()]
			if ok {
				ret.inputBalances[b.Color()] = s + b.Value()
			} else {
				ret.inputBalances[b.Color()] = b.Value()
			}
			ret.total += b.Value()
		}
	}
	if ret.total == 0 {
		return nil, fmt.Errorf("empty input balances")
	}
	// move smart contract token from SC address to itself
	if err := ret.MoveTokens(ret.scAddress, ret.scColor, 1); err != nil {
		return nil, err
	}
	ret.MustValidate()
	return ret, nil
}

func (txb *TransactionBuilder) Clone() *TransactionBuilder {
	panic("implement me")
}

func (txb *TransactionBuilder) AddRequestBlock(reqBlk *sctransaction.RequestBlock) error {
	// create and transfer request token to the target SC address
	if err := txb.NewColor(reqBlk.Address(), balance.ColorIOTA, 1); err != nil {
		return err
	}
	txb.sctxbuilder.AddRequestBlock(reqBlk)
	return nil
}

// Finalize produce final transaction
func (txb *TransactionBuilder) Finalize(stateIndex uint32, stateHash hashing.HashValue, timestamp int64) *sctransaction.Transaction {
	txb.MustValidate()

	if txb.sumInputs() != 0 {
		// if reminder wasn't added explicitly, assume own address
		txb.AddReminder(txb.scAddress)
	}
	txb.sctxbuilder.AddStateBlock(sctransaction.NewStateBlockParams{
		Color:      txb.scColor,
		StateIndex: stateIndex,
		StateHash:  stateHash,
		Timestamp:  timestamp,
	})

	txb.MustValidate()

	oids := make([]valuetransaction.OutputID, 0, len(txb.balances))
	for txid := range txb.balances {
		oids = append(oids, valuetransaction.NewOutputID(txb.scAddress, txid))
	}
	txb.sctxbuilder.MustAddInputs(oids...)

	for addr, lst := range txb.outputBalances {
		for col, v := range lst {
			txb.sctxbuilder.AddBalanceToOutput(addr, balance.New(col, v))
		}
	}
	ret, err := txb.sctxbuilder.Finalize()
	if err != nil {
		panic(err)
	}
	return ret
}

func (txb *TransactionBuilder) sumInputs() int64 {
	var ret int64
	for _, b := range txb.inputBalances {
		ret += b
	}
	return ret
}

func (txb *TransactionBuilder) sumOutputs() int64 {
	var ret int64
	for _, a := range txb.outputBalances {
		for _, s := range a {
			ret += s
		}
	}
	return ret
}

func (txb *TransactionBuilder) MustValidate() {
	if txb.sumInputs()+txb.sumOutputs() != txb.total {
		panic("inconsistency: unequal input and output balances")
	}
}

func (txb *TransactionBuilder) GetInputBalance(color balance.Color) (int64, bool) {
	ret, ok := txb.inputBalances[color]
	return ret, ok
}

func (txb *TransactionBuilder) GetOutputBalance(addr address.Address, color balance.Color) (int64, bool) {
	bals, ok := txb.outputBalances[addr]
	if !ok {
		return 0, false
	}
	ret, ok := bals[color]
	return ret, ok
}

func (txb *TransactionBuilder) InputColors() []balance.Color {
	ret := make([]balance.Color, 0, len(txb.inputBalances))
	for col := range txb.inputBalances {
		ret = append(ret, col)
	}
	return ret
}

// MoveTokens transfers tokens without changing color
func (txb *TransactionBuilder) MoveTokens(targetAddr address.Address, col balance.Color, amount int64) error {
	txb.MustValidate()
	inpb, ok := txb.inputBalances[col]
	if !ok {
		return fmt.Errorf("transfer: wrong color %s", col.String())
	}
	if inpb < amount {
		return fmt.Errorf("transfer: not enough funds of color %s", col.String())
	}
	txb.inputBalances[col] = txb.inputBalances[col] - amount

	if _, ok := txb.outputBalances[targetAddr]; !ok {
		txb.outputBalances[targetAddr] = make(map[balance.Color]int64)
	}
	s, _ := txb.outputBalances[targetAddr][col]
	s += amount
	txb.outputBalances[targetAddr][col] = s

	txb.MustValidate()
	return nil
}

// NewColor repaints tokens to new color
func (txb *TransactionBuilder) NewColor(targetAddr address.Address, col balance.Color, amount int64) error {
	inpb, ok := txb.inputBalances[col]
	if !ok {
		return fmt.Errorf("newColor: wrong color %s", col.String())
	}
	if inpb < amount {
		return fmt.Errorf("newColor: not enough funds of color %s", col.String())
	}
	txb.inputBalances[col] = txb.inputBalances[col] - amount

	if _, ok := txb.outputBalances[targetAddr]; !ok {
		txb.outputBalances[targetAddr] = make(map[balance.Color]int64)
	}
	s, _ := txb.outputBalances[targetAddr][balance.ColorNew]
	s += amount
	txb.outputBalances[targetAddr][balance.ColorNew] = s

	txb.MustValidate()
	return nil
}

// EraseColor repaints tokens to IOTA color
func (txb *TransactionBuilder) EraseColor(targetAddr address.Address, col balance.Color, amount int64) error {
	inpb, ok := txb.inputBalances[col]
	if !ok {
		return fmt.Errorf("eraseColor: wrong color %s", col.String())
	}
	if inpb < amount {
		return fmt.Errorf("eraseColor: not enough funds of color %s", col.String())
	}
	txb.inputBalances[col] = txb.inputBalances[col] - amount

	if _, ok := txb.outputBalances[targetAddr]; !ok {
		txb.outputBalances[targetAddr] = make(map[balance.Color]int64)
	}
	s, _ := txb.outputBalances[targetAddr][balance.ColorIOTA]
	s += amount
	txb.outputBalances[targetAddr][balance.ColorIOTA] = s

	txb.MustValidate()
	return nil
}

func (txb *TransactionBuilder) AddReminder(reminderAddress address.Address) {
	txb.MustValidate()
	if txb.sumInputs() == 0 {
		return
	}
	for col, v := range txb.inputBalances {
		if v > 0 {
			err := txb.MoveTokens(reminderAddress, col, v)
			if err != nil {
				panic(err)
			}
		}
	}
	if txb.sumInputs() != 0 {
		panic("AddReminder: txb.sumInputs() != 0")
	}
	txb.MustValidate()
}
