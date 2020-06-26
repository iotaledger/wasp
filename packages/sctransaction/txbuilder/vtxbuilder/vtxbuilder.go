// package to build value transaction
package vtxbuilder

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"sort"
)

type inputBalances struct {
	outputId valuetransaction.OutputID
	reminder []*balance.Balance
	consumed []*balance.Balance
}

type Builder struct {
	finalized             bool
	inputBalancesByOutput []inputBalances
	outputBalances        map[address.Address]map[balance.Color]int64
}

func newVTBuilder(orig *Builder) *Builder {
	if orig == nil {
		return &Builder{
			inputBalancesByOutput: make([]inputBalances, 0),
			outputBalances:        make(map[address.Address]map[balance.Color]int64),
		}
	}
	ret := &Builder{
		inputBalancesByOutput: make([]inputBalances, len(orig.inputBalancesByOutput)),
		outputBalances:        make(map[address.Address]map[balance.Color]int64),
	}
	for i := range ret.inputBalancesByOutput {
		ret.inputBalancesByOutput[i].outputId = orig.inputBalancesByOutput[i].outputId
		ret.inputBalancesByOutput[i].reminder = util.CloneBalances(orig.inputBalancesByOutput[i].reminder)
		ret.inputBalancesByOutput[i].consumed = util.CloneBalances(orig.inputBalancesByOutput[i].consumed)
	}
	for addr, bals := range orig.outputBalances {
		ret.outputBalances[addr] = make(map[balance.Color]int64)
		for col, b := range bals {
			ret.outputBalances[addr][col] = b
		}
	}
	return ret
}

var (
	errorWrongInputs      = errors.New("wrong inputs")
	errorWrongColor       = errors.New("wrong color")
	errorNotEnoughBalance = errors.New("non existent or not enough colored balance")
)

func NewFromAddressBalances(addr *address.Address, addressBalances map[valuetransaction.ID][]*balance.Balance) (*Builder, error) {
	ret := newVTBuilder(nil)
	var err error
	for txid, bals := range addressBalances {
		if (balance.Color)(txid) == balance.ColorNew || (balance.Color)(txid) == balance.ColorIOTA {
			return nil, errorWrongInputs
		}
		inb := inputBalances{
			outputId: valuetransaction.NewOutputID(*addr, txid),
			reminder: util.CloneBalances(bals),
			consumed: make([]*balance.Balance, 0, len(bals)),
		}
		inb.reminder, err = compressAndSortBalances(inb.reminder)
		if err != nil {
			return nil, err
		}
		ret.inputBalancesByOutput = append(ret.inputBalancesByOutput, inb)
	}
	return ret, nil
}

func NewFromOutputBalances(outputBalances map[valuetransaction.OutputID][]*balance.Balance) (*Builder, error) {
	ret := newVTBuilder(nil)
	var err error
	for oid, bals := range outputBalances {
		if (balance.Color)(oid.TransactionID()) == balance.ColorNew || (balance.Color)(oid.TransactionID()) == balance.ColorIOTA {
			return nil, errorWrongInputs
		}
		inb := inputBalances{
			outputId: oid,
			reminder: util.CloneBalances(bals),
			consumed: make([]*balance.Balance, 0, len(bals)),
		}
		inb.reminder, err = compressAndSortBalances(inb.reminder)
		if err != nil {
			return nil, err
		}
		ret.inputBalancesByOutput = append(ret.inputBalancesByOutput, inb)
	}
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

func (vtxb *Builder) Clone() *Builder {
	return newVTBuilder(vtxb)
}

// ForEachInputBalance iterates through reminders
func (vtxb *Builder) ForEachInputBalance(consumer func(oid *valuetransaction.OutputID, bals []*balance.Balance) bool) {
	for i := range vtxb.inputBalancesByOutput {
		if !consumer(&vtxb.inputBalancesByOutput[i].outputId, vtxb.inputBalancesByOutput[i].reminder) {
			return
		}
	}
}

func (vtxb *Builder) SetConsumerPrioritySmallerBalances() {
	sort.Slice(vtxb.inputBalancesByOutput, func(i, j int) bool {
		si := util.BalancesSumTotal(vtxb.inputBalancesByOutput[i].reminder)
		sj := util.BalancesSumTotal(vtxb.inputBalancesByOutput[j].reminder)
		if si == sj {
			return i < j
		}
		return si < sj
	})
}

func (vtxb *Builder) SetConsumerPriorityLargerBalances() {
	if vtxb.finalized {
		panic("using finalized transaction builder")
	}
	sort.Slice(vtxb.inputBalancesByOutput, func(i, j int) bool {
		si := util.BalancesSumTotal(vtxb.inputBalancesByOutput[i].reminder)
		sj := util.BalancesSumTotal(vtxb.inputBalancesByOutput[j].reminder)
		if si == sj {
			return i < j
		}
		return si > sj
	})
}

// GetInputBalance what is available in inputs
func (vtxb *Builder) GetInputBalance(col balance.Color) int64 {
	if vtxb.finalized {
		panic("using finalized transaction builder")
	}
	ret := int64(0)
	for _, inp := range vtxb.inputBalancesByOutput {
		ret += util.BalanceOfColor(inp.reminder, col)
	}
	return ret
}

// Returns unconsumed total
func subtractAmount(bals []*balance.Balance, col balance.Color, amount int64) int64 {
	if amount == 0 {
		return 0
	}
	for _, bal := range bals {
		if bal.Color == col {
			if bal.Value >= amount {
				bal.Value -= amount
				return 0
			}
			ret := amount - bal.Value
			bal.Value = 0
			return ret
		}
	}
	return amount
}

func addAmount(bals []*balance.Balance, col balance.Color, amount int64) []*balance.Balance {
	if amount == 0 {
		panic("addAmount: amount == 0")
	}
	for _, bal := range bals {
		if bal.Color == col {
			bal.Value += amount
			return bals
		}
	}
	return append(bals, balance.New(col, amount))
}

// don't do any validation, may panic
func (vtxb *Builder) moveAmount(targetAddr address.Address, origColor, targetColor balance.Color, amount int64) {
	saveAmount := amount
	if amount == 0 {
		return
	}
	for i := range vtxb.inputBalancesByOutput {
		amount = subtractAmount(vtxb.inputBalancesByOutput[i].reminder, origColor, amount)
		if amount == 0 {
			vtxb.inputBalancesByOutput[i].consumed = addAmount(vtxb.inputBalancesByOutput[i].consumed, origColor, saveAmount)
			break
		}
	}
	if amount > 0 {
		panic(errorNotEnoughBalance)
	}
	vtxb.addToOutputs(targetAddr, targetColor, saveAmount)
}

func (vtxb *Builder) addToOutputs(targetAddr address.Address, col balance.Color, amount int64) {
	cmap, ok := vtxb.outputBalances[targetAddr]
	if !ok {
		cmap = make(map[balance.Color]int64)
		vtxb.outputBalances[targetAddr] = cmap
	}
	b, _ := cmap[col]
	cmap[col] = b + amount
}

// move token without changing color
func (vtxb *Builder) MoveToAddress(targetAddr address.Address, col balance.Color, amount int64) error {
	if vtxb.finalized {
		panic("using finalized transaction builder")
	}
	if vtxb.GetInputBalance(col) < amount {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, col, col, amount)
	return nil
}

func (vtxb *Builder) EraseColor(targetAddr address.Address, col balance.Color, amount int64) error {
	if vtxb.finalized {
		panic("using finalized transaction builder")
	}
	actualBalance := vtxb.GetInputBalance(col)
	if actualBalance < amount {
		return fmt.Errorf("not enough balance: need %d, found %d, color %s",
			amount, actualBalance, col.String())
	}
	vtxb.moveAmount(targetAddr, col, balance.ColorIOTA, amount)
	return nil
}

func (vtxb *Builder) NewColor(targetAddr address.Address, col balance.Color, amount int64) error {
	if vtxb.finalized {
		panic("using finalized transaction builder")
	}
	if vtxb.GetInputBalance(col) < amount {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, col, balance.ColorNew, amount)
	return nil
}

// Build build the final value transaction: not signed and without data payload

func (vtxb *Builder) Build(useAllInputs bool) *valuetransaction.Transaction {
	if vtxb.finalized {
		panic("using finalized transaction builder")
	}
	defer func() {
		vtxb.finalized = true
	}()
	if !useAllInputs {
		// filter out unconsumed inputs
		finp := vtxb.inputBalancesByOutput[:0]
		for i := range vtxb.inputBalancesByOutput {
			if len(vtxb.inputBalancesByOutput[i].consumed) == 0 {
				continue
			}
			finp = append(finp, vtxb.inputBalancesByOutput[i])
		}
		vtxb.inputBalancesByOutput = finp
	}
	for i := range vtxb.inputBalancesByOutput {
		for _, bal := range vtxb.inputBalancesByOutput[i].reminder {
			if bal.Value > 0 {
				vtxb.addToOutputs(vtxb.inputBalancesByOutput[i].outputId.Address(), bal.Color, bal.Value)
				bal.Value = 0
			}
		}
	}
	inps := make([]valuetransaction.OutputID, len(vtxb.inputBalancesByOutput))
	for i := range inps {
		inps[i] = vtxb.inputBalancesByOutput[i].outputId
	}
	sort.Slice(inps, func(i, j int) bool {
		return bytes.Compare(inps[i][:], inps[j][:]) < 0
	})
	outmap := make(map[address.Address][]*balance.Balance)
	for addr, balmap := range vtxb.outputBalances {
		outmap[addr] = make([]*balance.Balance, 0, len(balmap))
		for col, b := range balmap {
			if b <= 0 {
				panic("internal inconsistency: balance value must be positive")
			}
			outmap[addr] = append(outmap[addr], balance.New(col, b))
		}
		sort.Slice(outmap[addr], func(i, j int) bool {
			return bytes.Compare(outmap[addr][i].Color[:], outmap[addr][j].Color[:]) < 0
		})
	}
	return valuetransaction.New(
		valuetransaction.NewInputs(inps...),
		valuetransaction.NewOutputs(outmap),
	)
}

func (vtxb *Builder) Dump() string {
	ret := ""
	// reminder
	for i := range vtxb.inputBalancesByOutput {
		ret += vtxb.inputBalancesByOutput[i].outputId.Address().String() + "-" +
			vtxb.inputBalancesByOutput[i].outputId.TransactionID().String() + "\n"
		for _, bal := range vtxb.inputBalancesByOutput[i].reminder {
			ret += fmt.Sprintf("      %s: %d\n", bal.Color.String(), bal.Value)
		}
	}
	// TODO the rest
	return ret
}
