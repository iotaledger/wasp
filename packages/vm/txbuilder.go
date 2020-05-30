package vm

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type TransactionBuilder struct {
	invalid        bool
	sctxbuilder    *sctransaction.TransactionBuilder
	balances       map[valuetransaction.ID][]*balance.Balance
	ownAddress     address.Address
	ownColor       balance.Color
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

// the transacion builder assumes valid smart contract token and immediately pushes to outputs, to the same address
func NewTxBuilder(par TransactionBuilderParams) (*TransactionBuilder, error) {
	ret := &TransactionBuilder{
		sctxbuilder:    sctransaction.NewTransactionBuilder(),
		balances:       par.Balances,
		inputBalances:  make(map[balance.Color]int64),
		outputBalances: make(map[address.Address]map[balance.Color]int64),
		ownAddress:     par.OwnAddress,
		ownColor:       par.OwnColor,
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
	if err := ret.Transfer(ret.ownAddress, ret.ownColor, 1); err != nil {
		return nil, err
	}
	ret.MustValidate()
	return ret, nil
}

func (acc *TransactionBuilder) AddRequestBlock(reqBlk *sctransaction.RequestBlock) error {
	// create and transfer request token to the target SC address
	if err := acc.NewColor(reqBlk.Address(), balance.ColorIOTA, 1); err != nil {
		return err
	}
	acc.sctxbuilder.AddRequestBlock(reqBlk)
	return nil
}

// Finalize produce final transaction
func (acc *TransactionBuilder) Finalize(stateIndex uint32, stateHash hashing.HashValue, timestamp int64) *sctransaction.Transaction {
	acc.MustValidate()

	if acc.sumInputs() != 0 {
		// if reminder wasn't added explicitely, assume own address
		acc.AddReminder(acc.ownAddress)
	}
	acc.sctxbuilder.AddStateBlock(sctransaction.NewStateBlockParams{
		Color:      acc.ownColor,
		StateIndex: stateIndex,
		StateHash:  stateHash,
		Timestamp:  timestamp,
	})

	acc.MustValidate()

	oids := make([]valuetransaction.OutputID, 0, len(acc.balances))
	for txid := range acc.balances {
		oids = append(oids, valuetransaction.NewOutputID(acc.ownAddress, txid))
	}
	acc.sctxbuilder.AddInputs(oids...)

	for addr, lst := range acc.outputBalances {
		for col, v := range lst {
			acc.sctxbuilder.AddBalanceToOutput(addr, balance.New(col, v))
		}
	}
	ret, err := acc.sctxbuilder.Finalize()
	if err != nil {
		panic(err)
	}
	return ret
}

func (acc *TransactionBuilder) sumInputs() int64 {
	var ret int64
	for _, b := range acc.inputBalances {
		ret += b
	}
	return ret
}

func (acc *TransactionBuilder) sumOutputs() int64 {
	var ret int64
	for _, a := range acc.outputBalances {
		for _, s := range a {
			ret += s
		}
	}
	return ret
}

func (acc *TransactionBuilder) MustValidate() {
	if acc.sumInputs()+acc.sumOutputs() != acc.total {
		panic("inconsistency: unequal input and output balances")
	}
}

func (acc *TransactionBuilder) GetInputBalance(color balance.Color) (int64, bool) {
	ret, ok := acc.inputBalances[color]
	return ret, ok
}

func (acc *TransactionBuilder) GetOutputBalance(addr address.Address, color balance.Color) (int64, bool) {
	bals, ok := acc.outputBalances[addr]
	if !ok {
		return 0, false
	}
	ret, ok := bals[color]
	return ret, ok
}

func (acc *TransactionBuilder) InputColors() []balance.Color {
	ret := make([]balance.Color, 0, len(acc.inputBalances))
	for col := range acc.inputBalances {
		ret = append(ret, col)
	}
	return ret
}

// Transfer transfers tokens without changing color
func (acc *TransactionBuilder) Transfer(targetAddr address.Address, col balance.Color, amount int64) error {
	acc.MustValidate()
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return fmt.Errorf("transfer: wrong color %s", col.String())
	}
	if inpb < amount {
		return fmt.Errorf("transfer: not enough funds of color %s", col.String())
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount

	if _, ok := acc.outputBalances[targetAddr]; !ok {
		acc.outputBalances[targetAddr] = make(map[balance.Color]int64)
	}
	s, _ := acc.outputBalances[targetAddr][col]
	s += amount
	acc.outputBalances[targetAddr][col] = s

	acc.MustValidate()
	return nil
}

// NewColor repaints tokens to new color
func (acc *TransactionBuilder) NewColor(targetAddr address.Address, col balance.Color, amount int64) error {
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return fmt.Errorf("newColor: wrong color %s", col.String())
	}
	if inpb < amount {
		return fmt.Errorf("newColor: not enough funds of color %s", col.String())
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount

	if _, ok := acc.outputBalances[targetAddr]; !ok {
		acc.outputBalances[targetAddr] = make(map[balance.Color]int64)
	}
	s, _ := acc.outputBalances[targetAddr][balance.ColorNew]
	s += amount
	acc.outputBalances[targetAddr][balance.ColorNew] = s

	acc.MustValidate()
	return nil
}

// EraseColor repaints tokens to IOTA color
func (acc *TransactionBuilder) EraseColor(targetAddr address.Address, col balance.Color, amount int64) error {
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return fmt.Errorf("eraseColor: wrong color %s", col.String())
	}
	if inpb < amount {
		return fmt.Errorf("eraseColor: not enough funds of color %s", col.String())
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount

	if _, ok := acc.outputBalances[targetAddr]; !ok {
		acc.outputBalances[targetAddr] = make(map[balance.Color]int64)
	}
	s, _ := acc.outputBalances[targetAddr][balance.ColorIOTA]
	s += amount
	acc.outputBalances[targetAddr][balance.ColorIOTA] = s

	acc.MustValidate()
	return nil
}

func (acc *TransactionBuilder) AddReminder(reminderAddress address.Address) {
	acc.MustValidate()
	if acc.sumInputs() == 0 {
		return
	}
	for col, v := range acc.inputBalances {
		if v > 0 {
			err := acc.Transfer(reminderAddress, col, v)
			if err != nil {
				panic(err)
			}
		}
	}
	if acc.sumInputs() != 0 {
		panic("AddReminder: acc.sumInputs() != 0")
	}
	acc.MustValidate()
}
