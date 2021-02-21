// package to build value transaction
package vtxbuilder

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/txutil"
)

type inputBalances struct {
	outputId valuetransaction.OutputID
	remain   []*balance.Balance
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
		ret.inputBalancesByOutput[i].remain = txutil.CloneBalances(orig.inputBalancesByOutput[i].remain)
		ret.inputBalancesByOutput[i].consumed = txutil.CloneBalances(orig.inputBalancesByOutput[i].consumed)
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
			remain:   txutil.CloneBalances(bals),
			consumed: make([]*balance.Balance, 0, len(bals)),
		}
		inb.remain, err = compressAndSortBalances(inb.remain)
		if err != nil {
			return nil, err
		}
		ret.inputBalancesByOutput = append(ret.inputBalancesByOutput, inb)
	}
	ret.sortInputBalancesById() // for determinism
	ret.SetConsumerPrioritySmallerBalances()
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
			remain:   txutil.CloneBalances(bals),
			consumed: make([]*balance.Balance, 0, len(bals)),
		}
		inb.remain, err = compressAndSortBalances(inb.remain)
		if err != nil {
			return nil, err
		}
		ret.inputBalancesByOutput = append(ret.inputBalancesByOutput, inb)
	}
	ret.sortInputBalancesById() // for determinism
	ret.SetConsumerPrioritySmallerBalances()
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
		if !consumer(&vtxb.inputBalancesByOutput[i].outputId, vtxb.inputBalancesByOutput[i].remain) {
			return
		}
	}
}

func (vtxb *Builder) sortInputBalancesById() {
	sort.Slice(vtxb.inputBalancesByOutput, func(i, j int) bool {
		return bytes.Compare(vtxb.inputBalancesByOutput[i].outputId[:], vtxb.inputBalancesByOutput[j].outputId[:]) < 0
	})
}

func (vtxb *Builder) SetConsumerPrioritySmallerBalances() {
	sort.Slice(vtxb.inputBalancesByOutput, func(i, j int) bool {
		si := txutil.BalancesSumTotal(vtxb.inputBalancesByOutput[i].remain)
		sj := txutil.BalancesSumTotal(vtxb.inputBalancesByOutput[j].remain)
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
		si := txutil.BalancesSumTotal(vtxb.inputBalancesByOutput[i].remain)
		sj := txutil.BalancesSumTotal(vtxb.inputBalancesByOutput[j].remain)
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
		ret += txutil.BalanceOfColor(inp.remain, col)
	}
	return ret
}

// Returns consumed and unconsumed total
func subtractAmount(bals []*balance.Balance, col balance.Color, amount int64) (int64, int64) {
	if amount == 0 {
		return 0, 0
	}
	for _, bal := range bals {
		if bal.Color == col {
			if bal.Value >= amount {
				bal.Value -= amount
				return amount, 0
			}
			bal.Value = 0
			return bal.Value, amount - bal.Value
		}
	}
	return 0, amount
}

func addAmount(bals []*balance.Balance, col balance.Color, amount int64) []*balance.Balance {
	if amount == 0 {
		return bals
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
func (vtxb *Builder) moveAmount(targetAddr address.Address, origColor, targetColor balance.Color, amountToConsume int64) {
	saveAmount := amountToConsume
	if amountToConsume == 0 {
		return
	}
	var consumedAmount int64
	for i := range vtxb.inputBalancesByOutput {
		consumedAmount, amountToConsume = subtractAmount(vtxb.inputBalancesByOutput[i].remain, origColor, amountToConsume)
		vtxb.inputBalancesByOutput[i].consumed = addAmount(vtxb.inputBalancesByOutput[i].consumed, origColor, consumedAmount)
		if amountToConsume == 0 {
			break
		}
	}
	if amountToConsume > 0 {
		panic(errorNotEnoughBalance)
	}
	vtxb.addToOutputs(targetAddr, targetColor, saveAmount)
}

func (vtxb *Builder) moveAmountFromTransaction(targetAddr address.Address, origColor, targetColor balance.Color, amountToConsume int64, txid valuetransaction.ID) {
	saveAmount := amountToConsume
	if amountToConsume == 0 {
		return
	}
	for i := range vtxb.inputBalancesByOutput {
		if vtxb.inputBalancesByOutput[i].outputId.TransactionID() != txid {
			continue
		}
		var consumedAmount int64
		consumedAmount, amountToConsume = subtractAmount(vtxb.inputBalancesByOutput[i].remain, origColor, amountToConsume)
		vtxb.inputBalancesByOutput[i].consumed = addAmount(vtxb.inputBalancesByOutput[i].consumed, origColor, consumedAmount)
		if amountToConsume == 0 {
			break
		}
	}
	if amountToConsume > 0 {
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

// MoveTokensToAddress move token without changing color
func (vtxb *Builder) MoveTokensToAddress(targetAddr address.Address, col balance.Color, amount int64) error {
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
		return fmt.Errorf("EraseColor: not enough balance: need %d, found %d, color %s",
			amount, actualBalance, col.String())
	}
	vtxb.moveAmount(targetAddr, col, balance.ColorIOTA, amount)
	return nil
}

// MintColoredTokens creates output of NewColor tokens out of inputs with specified color
func (vtxb *Builder) MintColoredTokens(targetAddr address.Address, sourceColor balance.Color, amount int64) error {
	if vtxb.finalized {
		panic("using finalized transaction builder")
	}
	if vtxb.GetInputBalance(sourceColor) < amount {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, sourceColor, balance.ColorNew, amount)
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
		for _, bal := range vtxb.inputBalancesByOutput[i].remain {
			if bal.Value > 0 {
				vtxb.addToOutputs(vtxb.inputBalancesByOutput[i].outputId.Address(), bal.Color, bal.Value)
				//bal.Value = 0
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
	ret := "inputs:\n"
	// remain
	for i := range vtxb.inputBalancesByOutput {
		ret += vtxb.inputBalancesByOutput[i].outputId.Address().String() + " - " +
			vtxb.inputBalancesByOutput[i].outputId.TransactionID().String() + "\n"
		for _, bal := range vtxb.inputBalancesByOutput[i].remain {
			ret += fmt.Sprintf("      remain %d %s\n", bal.Value, bal.Color.String())
		}
		for _, bal := range vtxb.inputBalancesByOutput[i].consumed {
			ret += fmt.Sprintf("      consumed %d %s\n", bal.Value, bal.Color.String())
		}
	}
	ret += "outputs:\n"
	for addr, balmap := range vtxb.outputBalances {
		ret += fmt.Sprintf("        %s\n", addr.String())
		for c, b := range balmap {
			ret += fmt.Sprintf("                         %s: %d\n", c.String(), b)
		}
	}
	return ret
}
