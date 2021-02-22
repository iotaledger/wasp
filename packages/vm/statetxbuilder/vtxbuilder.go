// package to build value transaction for the anchor transaction
package statetxbuilder

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
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

type vtxBuilder struct {
	reminderAddr          address.Address
	inputBalancesByOutput []inputBalances
	outputBalances        map[address.Address]map[balance.Color]int64
	originalBalances      map[balance.Color]int64
}

var (
	errorWrongInputs      = errors.New("wrong inputs")
	errorWrongColor       = errors.New("wrong color")
	errorNotEnoughBalance = errors.New("non existent or not enough colored balance")
)

func newValueTxBuilder(addr address.Address, addressBalances map[valuetransaction.ID][]*balance.Balance) (*vtxBuilder, error) {
	ret := &vtxBuilder{
		reminderAddr:          addr,
		inputBalancesByOutput: make([]inputBalances, 0),
		outputBalances:        make(map[address.Address]map[balance.Color]int64),
		originalBalances:      make(map[balance.Color]int64),
	}
	var err error
	for txid, bals := range addressBalances {
		if balance.Color(txid) == balance.ColorNew || balance.Color(txid) == balance.ColorIOTA {
			return nil, errorWrongInputs
		}
		inb := inputBalances{
			outputId: valuetransaction.NewOutputID(addr, txid),
			remain:   txutil.CloneBalances(bals),
			consumed: make([]*balance.Balance, 0, len(bals)),
		}
		inb.remain, err = copyCompressAndSortBalances(inb.remain)
		if err != nil {
			return nil, err
		}
		ret.inputBalancesByOutput = append(ret.inputBalancesByOutput, inb)

		for _, bal := range bals {
			b, _ := ret.originalBalances[bal.Color]
			ret.originalBalances[bal.Color] = b + bal.Value
		}
	}
	ret.sortInputBalancesById() // for determinism
	return ret, nil
}

func (vtxb *vtxBuilder) clone() *vtxBuilder {
	ret := &vtxBuilder{
		reminderAddr:          vtxb.reminderAddr,
		inputBalancesByOutput: make([]inputBalances, 0),
		outputBalances:        make(map[address.Address]map[balance.Color]int64),
		originalBalances:      make(map[balance.Color]int64),
	}
	for _, b := range vtxb.inputBalancesByOutput {
		remain, err := copyCompressAndSortBalances(b.remain)
		if err != nil {
			panic(err)
		}
		consumed, err := copyCompressAndSortBalances(b.consumed)
		if err != nil {
			panic(err)
		}
		ret.inputBalancesByOutput = append(ret.inputBalancesByOutput, inputBalances{
			outputId: b.outputId,
			remain:   remain,
			consumed: consumed,
		})
	}
	for addr, balmap := range vtxb.outputBalances {
		m := make(map[balance.Color]int64)
		for k, v := range balmap {
			m[k] = v
		}
		ret.outputBalances[addr] = m
	}
	for col, b := range vtxb.originalBalances {
		ret.originalBalances[col] = b
	}
	return ret
}

// ForEachInputBalance iterates through reminders
func (vtxb *vtxBuilder) ForEachInputBalance(consumer func(oid *valuetransaction.OutputID, bals []*balance.Balance) bool) {
	for i := range vtxb.inputBalancesByOutput {
		if !consumer(&vtxb.inputBalancesByOutput[i].outputId, vtxb.inputBalancesByOutput[i].remain) {
			return
		}
	}
}

// makes each color unique, sums up balances of repeating colors. Sorts colors.
// Returns underlying array!!
func copyCompressAndSortBalances(bals []*balance.Balance) ([]*balance.Balance, error) {
	if len(bals) == 0 {
		return nil, nil
	}
	ret := make([]*balance.Balance, 0, len(bals))
	bmap := make(map[balance.Color]int64)
	for _, bal := range bals {
		if _, ok := bmap[bal.Color]; !ok {
			bmap[bal.Color] = 0
		}
		bmap[bal.Color] += bal.Value
	}
	for col, val := range bmap {
		ret = append(ret, balance.New(col, val))
	}
	sort.Slice(ret, func(i, j int) bool {
		return bytes.Compare(ret[i].Color[:], ret[j].Color[:]) < 0
	})
	return ret, nil
}

func (vtxb *vtxBuilder) sortInputBalancesById() {
	sort.Slice(vtxb.inputBalancesByOutput, func(i, j int) bool {
		return bytes.Compare(vtxb.inputBalancesByOutput[i].outputId[:], vtxb.inputBalancesByOutput[j].outputId[:]) < 0
	})
}

// GetInputBalance what remains available in inputs
func (vtxb *vtxBuilder) GetInputBalance(col balance.Color, filterTxid ...valuetransaction.ID) int64 {
	takeFromTransaction := len(filterTxid) > 0
	ret := int64(0)
	for _, inp := range vtxb.inputBalancesByOutput {
		if takeFromTransaction && inp.outputId.TransactionID() != filterTxid[0] {
			continue
		}
		ret += txutil.BalanceOfColor(inp.remain, col)
	}
	return ret
}

// subtractAmount returns consumed and remaining (unconsumed) balances
func subtractAmount(bals []*balance.Balance, col balance.Color, amount int64) (int64, int64) {
	if amount == 0 {
		return 0, 0
	}
	var consumed, unconsumed int64
	for _, bal := range bals {
		if bal.Color == col {
			if bal.Value >= amount {
				consumed = amount
				unconsumed = 0
			} else {
				consumed = bal.Value
				unconsumed = amount - bal.Value
			}
			bal.Value -= consumed
			return consumed, unconsumed
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
func (vtxb *vtxBuilder) moveAmount(targetAddr address.Address, origColor, targetColor balance.Color, amountToConsume int64, filterTxid ...valuetransaction.ID) {
	takeFromTransaction := len(filterTxid) > 0
	saveAmount := amountToConsume
	if amountToConsume == 0 {
		return
	}
	var consumedAmount int64
	for i := range vtxb.inputBalancesByOutput {
		if takeFromTransaction && vtxb.inputBalancesByOutput[i].outputId.TransactionID() != filterTxid[0] {
			continue
		}
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

func (vtxb *vtxBuilder) addToOutputs(targetAddr address.Address, col balance.Color, amount int64) {
	cmap, ok := vtxb.outputBalances[targetAddr]
	if !ok {
		cmap = make(map[balance.Color]int64)
		vtxb.outputBalances[targetAddr] = cmap
	}
	b, _ := cmap[col]
	cmap[col] = b + amount
}

// MoveTokens move token without changing color
func (vtxb *vtxBuilder) MoveTokens(targetAddr address.Address, col balance.Color, amount int64, filterTxid ...valuetransaction.ID) error {
	if vtxb.GetInputBalance(col, filterTxid...) < amount {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, col, col, amount, filterTxid...)
	return nil
}

func (vtxb *vtxBuilder) EraseColor(targetAddr address.Address, col balance.Color, amount int64, filterTxid ...valuetransaction.ID) error {
	actualBalance := vtxb.GetInputBalance(col, filterTxid...)
	if actualBalance < amount {
		return fmt.Errorf("EraseColor: not enough balance: need %d, found %d, color %s",
			amount, actualBalance, col.String())
	}
	vtxb.moveAmount(targetAddr, col, balance.ColorIOTA, amount, filterTxid...)
	return nil
}

// MintColor creates output of NewColor tokens out of inputs with specified color
func (vtxb *vtxBuilder) MintColor(targetAddr address.Address, sourceColor balance.Color, amount int64, filterTxid ...valuetransaction.ID) error {
	if vtxb.GetInputBalance(sourceColor, filterTxid...) < amount {
		return errorNotEnoughBalance
	}
	vtxb.moveAmount(targetAddr, sourceColor, balance.ColorNew, amount, filterTxid...)
	return nil
}

// Build build the final value transaction: not signed and without data payload
func (vtxb *vtxBuilder) build() *valuetransaction.Transaction {
	// send all remaining balances to the main address
	for i := range vtxb.inputBalancesByOutput {
		for _, bal := range vtxb.inputBalancesByOutput[i].remain {
			if bal.Value > 0 {
				vtxb.addToOutputs(vtxb.reminderAddr, bal.Color, bal.Value)
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
		var err error
		outmap[addr], err = copyCompressAndSortBalances(outmap[addr])
		if err != nil {
			panic(err)
		}
	}
	inputs := valuetransaction.NewInputs(inps...)
	outputs := valuetransaction.NewOutputs(outmap)
	//fmt.Printf("-- building: inputs %s\n", inputs.String())
	//fmt.Printf("-- building: outputs %s\n", outputs.String())
	return valuetransaction.New(inputs, outputs)
}

func (vtxb *vtxBuilder) Dump() string {
	ret := fmt.Sprintf("remainder address; %s\ninputs:\n", vtxb.reminderAddr.String())
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

func (vtxb *vtxBuilder) validate() {
	remaining := make(map[balance.Color]int64)
	consumed := make(map[balance.Color]int64)
	inOutputs := make(map[balance.Color]int64)

	for _, ib := range vtxb.inputBalancesByOutput {
		for _, b := range ib.remain {
			s, _ := remaining[b.Color]
			remaining[b.Color] = s + b.Value
		}
		for _, b := range ib.consumed {
			s, _ := consumed[b.Color]
			consumed[b.Color] = s + b.Value
		}
	}
	for _, balmap := range vtxb.outputBalances {
		for col, val := range balmap {
			s, _ := inOutputs[col]
			inOutputs[col] = s + val
		}
	}
	orig := cbalances.NewFromMap(vtxb.originalBalances)
	rem := cbalances.NewFromMap(remaining)
	cons := cbalances.NewFromMap(consumed)

	sumMap := make(map[balance.Color]int64)
	rem.AddToMap(sumMap)
	cons.AddToMap(sumMap)
	sum := cbalances.NewFromMap(sumMap)
	if !sum.Equal(orig) {
		panic("invalid sums")
	}
}
