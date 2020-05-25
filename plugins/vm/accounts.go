package vm

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type accounts struct {
	ownAddress     address.Address
	inputBalances  map[balance.Color]int64
	outputBalances map[address.Address]map[balance.Color]int64
	reminder       int64
	total          int64
}

// AccountsFromBalances
type AccountsFromBalancesParams struct {
	Balances   map[valuetransaction.ID][]*balance.Balance
	OwnColor   balance.Color
	OwnAddress address.Address
	RequestIds []sctransaction.RequestId
}

func AccountsFromBalances(par AccountsFromBalancesParams) (*accounts, error) {
	ret := &accounts{
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

	ret.reminder = ret.total

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

func (acc *accounts) MustValidate() {
	var sumIn, sumOut int64

	for _, b := range acc.inputBalances {
		sumIn += b
	}
	for _, a := range acc.outputBalances {
		for _, s := range a {
			sumOut += s
		}
	}
	if sumOut+sumIn != acc.total {
		panic("wrong balance I")
	}
	if sumIn != acc.reminder {
		panic("wrong balance II")
	}
}

func (acc *accounts) GetInputBalance(color balance.Color) (int64, bool) {
	ret, ok := acc.inputBalances[color]
	return ret, ok
}

func (acc *accounts) GetOutputBalance(addr address.Address, color balance.Color) (int64, bool) {
	bals, ok := acc.outputBalances[addr]
	if !ok {
		return 0, false
	}
	ret, ok := bals[color]
	return ret, ok
}

func (acc *accounts) InputColors() []balance.Color {
	ret := make([]balance.Color, 0, len(acc.inputBalances))
	for col := range acc.inputBalances {
		ret = append(ret, col)
	}
	return ret
}

func (acc *accounts) Reminder() int64 {
	return acc.reminder
}

// Transfer transfers tokens without changing color
func (acc *accounts) Transfer(targetAddr address.Address, col balance.Color, amount int64) error {
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount
	acc.reminder = acc.reminder - amount

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
func (acc *accounts) NewColor(targetAddr address.Address, col balance.Color, amount int64) error {
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount
	acc.reminder = acc.reminder - amount

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
func (acc *accounts) EraseColor(targetAddr address.Address, col balance.Color, amount int64) error {
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount
	acc.reminder = acc.reminder - amount

	if _, ok := acc.outputBalances[targetAddr]; !ok {
		acc.outputBalances[targetAddr] = make(map[balance.Color]int64)
	}
	s, _ := acc.outputBalances[targetAddr][balance.ColorIOTA]
	s += amount
	acc.outputBalances[targetAddr][balance.ColorIOTA] = s

	acc.MustValidate()
	return nil
}
