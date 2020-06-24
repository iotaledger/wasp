// package to build smart contract transaction (including UTXO value transation part)
package txbuilder

import (
	"bytes"
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"sort"
)

type ValueTransactionBuilder struct {
	inputBalancesByOutput  map[valuetransaction.OutputID][]*balance.Balance
	inputBalancesByAddress map[address.Address][]*balance.Balance
	inputBalancesByColor   map[balance.Color]int64
	outputBalances         map[address.Address]map[balance.Color]int64
}

func newVTBuilder() *ValueTransactionBuilder {
	return &ValueTransactionBuilder{
		inputBalancesByOutput:  make(map[valuetransaction.OutputID][]*balance.Balance),
		inputBalancesByAddress: make(map[address.Address][]*balance.Balance),
		inputBalancesByColor:   make(map[balance.Color]int64),
		outputBalances:         make(map[address.Address]map[balance.Color]int64),
	}
}

var (
	errorWrongInputs      = errors.New("wrong inputs")
	errorWrongColor       = errors.New("wrong color")
	errorNotEnoughBalance = errors.New("non existent or not enough colored balance")
)

func NewFromAddressBalances(addr *address.Address, addressBalances map[valuetransaction.ID][]*balance.Balance) (*ValueTransactionBuilder, error) {
	ret := newVTBuilder()

	for txid, bals := range addressBalances {
		if (balance.Color)(txid) == balance.ColorNew || (balance.Color)(txid) == balance.ColorIOTA {
			return nil, errorWrongInputs
		}
		ret.inputBalancesByOutput[valuetransaction.NewOutputID(*addr, txid)] = bals
	}
	ret.collectInputBalancesByColor()
	return ret, nil
}

func NewFromOutputBalances(outputBalances map[valuetransaction.OutputID][]*balance.Balance) (*ValueTransactionBuilder, error) {
	ret := newVTBuilder()
	for oid, bals := range outputBalances {
		if (balance.Color)(oid.TransactionID()) == balance.ColorNew || (balance.Color)(oid.TransactionID()) == balance.ColorIOTA {
			return nil, errorWrongInputs
		}
		lst, ok := ret.inputBalancesByAddress[oid.Address()]
		if !ok {
			lst = make([]*balance.Balance, 0)
		}
		lst = append(lst, bals...)
		ret.inputBalancesByAddress[oid.Address()] = lst
	}
	for addr, bals := range ret.inputBalancesByAddress {
		c, err := compressAndSortBalances(bals)
		if err != nil {
			return nil, err
		}
		ret.inputBalancesByAddress[addr] = c
	}
	ret.collectInputBalancesByColor()
	return ret, nil
}

// makes each color unique, sums up balances of repeating colors. Sorts colors.
// Returns underlying array!!
func compressAndSortBalances(bals []*balance.Balance) ([]*balance.Balance, error) {
	if len(bals) == 0 {
		return nil, nil
	}
	bmap := make(map[balance.Color]int64)
	for _, bal := range bals {
		if bal.Color == balance.ColorNew {
			return nil, errorWrongColor
		}
		if _, ok := bmap[bal.Color]; !ok {
			bmap[bal.Color] = 0
		}
		bmap[bal.Color] += bal.Value
	}
	bals = bals[:0]
	for col, val := range bmap {
		bals = append(bals, balance.New(col, val))
	}
	sort.Slice(bals, func(i, j int) bool {
		return bytes.Compare(bals[i].Color[:], bals[j].Color[:]) < 0
	})
	return bals, nil
}

func (vtxb *ValueTransactionBuilder) collectInputBalancesByColor() {
	for _, bals := range vtxb.inputBalancesByAddress {
		for _, bal := range bals {
			if _, ok := vtxb.inputBalancesByColor[bal.Color]; !ok {
				vtxb.inputBalancesByColor[bal.Color] = 0
			}
			vtxb.inputBalancesByColor[bal.Color] += bal.Value
		}
	}
}

func (vtxb *ValueTransactionBuilder) Clone() *ValueTransactionBuilder {
	ret := &ValueTransactionBuilder{
		inputBalancesByOutput:  vtxb.inputBalancesByOutput,    // immutable
		inputBalancesByAddress: vtxb.inputBalancesByAddress,   // immutable
		inputBalancesByColor:   make(map[balance.Color]int64), // changes in progress
		outputBalances:         make(map[address.Address]map[balance.Color]int64),
	}
	for col, b := range vtxb.inputBalancesByColor {
		ret.inputBalancesByColor[col] = b
	}

	for addr, bals := range vtxb.outputBalances {
		ret.outputBalances[addr] = make(map[balance.Color]int64)
		for col, b := range bals {
			ret.outputBalances[addr][col] = b
		}
	}
	return ret
}

// GetInputBalance what is available in inputs
func (vtxb *ValueTransactionBuilder) GetInputBalance(col balance.Color) (int64, bool) {
	ret, ok := vtxb.inputBalancesByColor[col]
	return ret, ok
}

// don't do any validation, can panic
func (vtxb *ValueTransactionBuilder) moveAmount(targetAddr address.Address, origColor, targetColor balance.Color, amount int64) {
	cmap, ok := vtxb.outputBalances[targetAddr]
	if !ok {
		cmap = make(map[balance.Color]int64)
	}
	b, _ := cmap[targetColor]
	cmap[targetColor] = b + amount
	vtxb.inputBalancesByColor[origColor] -= amount
}

// move token without changing color
func (vtxb *ValueTransactionBuilder) MoveToAddress(targetAddr address.Address, col balance.Color, amount int64) error {
	bal, ok := vtxb.inputBalancesByColor[col]
	if !ok || amount < bal {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, col, col, amount)
	return nil
}

func (vtxb *ValueTransactionBuilder) EraseColor(targetAddr address.Address, col balance.Color, amount int64) error {
	bal, ok := vtxb.inputBalancesByColor[col]
	if !ok || amount < bal {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, col, balance.ColorIOTA, amount)
	return nil
}

func (vtxb *ValueTransactionBuilder) NewColor(targetAddr address.Address, col balance.Color, amount int64) error {
	bal, ok := vtxb.inputBalancesByColor[col]
	if !ok || amount < bal {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, col, balance.ColorNew, amount)
	return nil
}

// Build build the final value transaction: not signed and without data payload

const (
	buildModeConsumeSmallerFirst = 0
	buildModeConsumeBiggerFirst  = 1
	buildConsumeAllInputs        = 2
)

func (vtxb *ValueTransactionBuilder) build(buildMode int) *valuetransaction.Transaction {
	switch buildMode {
	case buildModeConsumeSmallerFirst:

	case buildModeConsumeBiggerFirst:
	case buildConsumeAllInputs:

	}
	panic("wrong")
}

func (vtxb *ValueTransactionBuilder) BuildConsumeAllInputs(reminderAddress address.Address) *valuetransaction.Transaction {
	for col, b := range vtxb.inputBalancesByColor {
		vtxb.moveAmount(reminderAddress, col, col, b)
	}
	inps := make([]valuetransaction.OutputID, 0, len(vtxb.inputBalancesByOutput))
	for oid := range vtxb.inputBalancesByOutput {
		inps = append(inps, oid)
	}
	sort.Slice(inps, func(i, j int) bool {
		return bytes.Compare(inps[i][:], inps[j][:]) < 0
	})
	inputs := valuetransaction.NewInputs(inps...)

	ret := valuetransaction.New()

}
