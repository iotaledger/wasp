package vm

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type TransactionBuilder struct {
	sctxbuilder    *sctransaction.TransactionBuilder
	balances       map[valuetransaction.ID][]*balance.Balance
	ownAddress     address.Address
	inputBalances  map[balance.Color]int64
	outputBalances map[address.Address]map[balance.Color]int64
	total          int64
}

// NewTxBuilder
type TransactionBuilderParams struct {
	Balances   map[valuetransaction.ID][]*balance.Balance
	OwnColor   balance.Color
	OwnAddress address.Address
	RequestIds []sctransaction.RequestId
}

func NewTxBuilder(par TransactionBuilderParams) (*TransactionBuilder, error) {
	ret := &TransactionBuilder{
		sctxbuilder:    sctransaction.NewTransactionBuilder(),
		balances:       par.Balances,
		inputBalances:  make(map[balance.Color]int64),
		outputBalances: make(map[address.Address]map[balance.Color]int64),
	}
	for _, lst := range par.Balances {
		for _, b := range lst {
			s, _ := ret.inputBalances[b.Color()]
			s = s + b.Value()
			ret.inputBalances[b.Color()] = s
			ret.total += s
		}
	}

	ret.ownAddress = par.OwnAddress
	// transfer smart contract token
	if err := ret.Transfer(par.OwnAddress, par.OwnColor, 1); err != nil {
		return nil, err
	}
	// destroy tokens corresponding to requests
	for i := range par.RequestIds {
		if err := ret.EraseColor(par.OwnAddress, (balance.Color)(*par.RequestIds[i].TransactionId()), 1); err != nil {
			return nil, err
		}
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

// SqueezeTransactionBuilder produces not finalized yet transaction builder
// without state and request blocks
func (acc *TransactionBuilder) Finalize(stateIndex uint32, stateHash hashing.HashValue) *sctransaction.Transaction {
	if !(acc.sumInputs() == 0) {
		panic("assertion failed: acc.sumInputs() == 0")
	}
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
		panic("wrong balance I")
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
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
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
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
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
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
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

func (acc *TransactionBuilder) AddReminder(reminderAddress *address.Address) {
	if acc.sumInputs() == 0 {
		return
	}
	if reminderAddress == nil {
		reminderAddress = &acc.ownAddress
	}

	for col, v := range acc.inputBalances {
		if v > 0 {
			err := acc.Transfer(*reminderAddress, col, v)
			if err != nil {
				panic(err)
			}
		}
	}
	if acc.sumInputs() != 0 {
		panic("Finalize: acc.sumInputs() != 0")
	}
}
