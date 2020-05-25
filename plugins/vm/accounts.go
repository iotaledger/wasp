package vm

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type accounts struct {
	inputBalances  map[balance.Color]int64
	moved          int64
	outputBalances map[balance.Color]int64
	reminder       int64
}

func accountsFromBalancesPlain(bals map[valuetransaction.ID][]*balance.Balance) *accounts {
	ret := &accounts{
		inputBalances:  make(map[balance.Color]int64),
		outputBalances: make(map[balance.Color]int64),
	}
	total := int64(0)
	for _, lst := range bals {
		for _, b := range lst {
			s, _ := ret.inputBalances[b.Color()]
			s = s + b.Value()
			ret.inputBalances[b.Color()] = s
			total += s
			if _, ok := ret.outputBalances[b.Color()]; !ok {
				ret.outputBalances[b.Color()] = 0
			}
		}
	}
	ret.reminder = total
	if _, ok := ret.outputBalances[balance.ColorIOTA]; !ok {
		ret.outputBalances[balance.ColorIOTA] = 0
	}
	ret.outputBalances[balance.ColorNew] = 0

	return ret
}

// AccountsFromBalances
func AccountsFromBalances(bals map[valuetransaction.ID][]*balance.Balance, scColor balance.Color, reqids []sctransaction.RequestId) (*accounts, error) {
	ret := accountsFromBalancesPlain(bals)
	// transfer smart contract token
	if err := ret.Transfer(scColor, 1); err != nil {
		return nil, err
	}
	// destroy tokens corresponding to requests
	for i := range reqids {
		if err := ret.DestroyColor((balance.Color)(*reqids[i].TransactionId()), 1); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (acc *accounts) GetInputBalance(color balance.Color) (int64, bool) {
	ret, ok := acc.inputBalances[color]
	return ret, ok
}

func (acc *accounts) GetOutputBalance(color balance.Color) (int64, bool) {
	ret, ok := acc.outputBalances[color]
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
func (acc *accounts) Transfer(col balance.Color, amount int64) error {
	if col == balance.ColorNew {
		return errors.New("can't use new color")
	}
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount
	acc.moved = acc.moved + amount

	acc.outputBalances[col] = acc.outputBalances[col] + amount
	acc.reminder = acc.reminder - amount

	return nil
}

// NewColor repaints tokens to new color
func (acc *accounts) NewColor(col balance.Color, amount int64) error {
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount
	acc.moved = acc.moved + amount

	acc.outputBalances[balance.ColorNew] = acc.outputBalances[balance.ColorNew] + amount
	acc.reminder = acc.reminder - amount

	return nil
}

// DestroyColor repaints tokens to iotas
func (acc *accounts) DestroyColor(col balance.Color, amount int64) error {
	inpb, ok := acc.inputBalances[col]
	if !ok {
		return errors.New("wrong color")
	}
	if inpb < amount {
		return errors.New("not enough funds")
	}
	acc.inputBalances[col] = acc.inputBalances[col] - amount
	acc.moved = acc.moved + amount

	acc.outputBalances[balance.ColorIOTA] = acc.outputBalances[balance.ColorNew] + amount
	acc.reminder = acc.reminder - amount

	return nil
}
