package util

import (
	"bytes"
	"fmt"
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

// selects minimum output ot of given outputs to be enough to transfer the amount
// outputs are selected in the ascending order of balances with specified color
// returns nil if outputs are not enough for the amount
func SelectMinimumOutputs(outputs map[valuetransaction.OutputID][]*balance.Balance, color balance.Color, amount int64) map[valuetransaction.OutputID][]*balance.Balance {
	sorted := sorteableOutputs{
		lessComparator: func(so1, so2 *sorteableOutput) bool {
			b1 := BalanceOfColor(so1.balances, color)
			b2 := BalanceOfColor(so2.balances, color)
			if b1 == 0 {
				return false
			}
			if b1 == 0 || b2 == 0 {
				panic("b1 == 0 || b2 == 0")
			}
			switch {
			case b1 < b2:
				return true
			case b1 > b2:
				return false
			case b1 == b2:
				return bytes.Compare(so1.id[:], so2.id[:]) < 0
			}
			panic("can't be")
		},
		outputs: make([]*sorteableOutput, 0, len(outputs)),
	}
	for aoid, bals := range outputs {
		if BalanceOfColor(bals, color) != 0 {
			sorted.outputs = append(sorted.outputs, &sorteableOutput{
				id:       aoid,
				balances: bals,
			})
		}
	}
	sort.Sort(sorted)

	ret := make(map[valuetransaction.OutputID][]*balance.Balance)
	sum := int64(0)
	for _, o := range sorted.outputs {
		ret[o.id] = o.balances
		sum += BalanceOfColor(o.balances, color)
		if sum >= amount {
			return ret
		}
	}
	return nil
}

func BalancesToString(outs map[valuetransaction.ID][]*balance.Balance) string {
	ret := "{"
	for txid, bals := range outs {
		ret += txid.String() + ":\n"
		for _, bal := range bals {
			ret += fmt.Sprintf("         %s: %d\n", bal.Color().String(), bal.Value())
		}
	}
	return ret
}
