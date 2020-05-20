package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"sort"
)

type sorteableOutput struct {
	id       valuetransaction.OutputID
	balances []*balance.Balance
}

type sorteableOutputs struct {
	lessComparator func(so1, so2 *sorteableOutput) bool
	outputs        []*sorteableOutput
}

func (souts sorteableOutputs) Len() int {
	return len(souts.outputs)
}

func (souts sorteableOutputs) Less(i, j int) bool {
	return souts.lessComparator(souts.outputs[i], souts.outputs[j])
}

func (souts sorteableOutputs) Swap(i, j int) {
	souts.outputs[i], souts.outputs[j] = souts.outputs[j], souts.outputs[i]
}

// return nil if outputs are not enough for the amount
func SelectMinimumOutputs(outputs map[valuetransaction.OutputID][]*balance.Balance, color balance.Color, amount int64) map[valuetransaction.OutputID][]*balance.Balance {
	sorted := sorteableOutputs{
		lessComparator: func(so1, so2 *sorteableOutput) bool {
			b1 := balanceOfColor(so1.balances, color)
			b2 := balanceOfColor(so2.balances, color)
			return b1 != 0 && b1 < b2
		},
		outputs: make([]*sorteableOutput, 0, len(outputs)),
	}
	for aoid, bals := range outputs {
		sorted.outputs = append(sorted.outputs, &sorteableOutput{
			id:       aoid,
			balances: bals,
		})
	}
	sort.Sort(sorted)

	ret := make(map[valuetransaction.OutputID][]*balance.Balance)
	sum := int64(0)
	for _, o := range sorted.outputs {
		ret[o.id] = o.balances
		sum += balanceOfColor(o.balances, color)
		if sum >= amount {
			return ret
		}
	}
	return nil
}

func balanceOfColor(bals []*balance.Balance, color balance.Color) int64 {
	sum := int64(0)
	for _, b := range bals {
		if b.Color() == color {
			sum += b.Value()
		}
	}
	return sum
}
