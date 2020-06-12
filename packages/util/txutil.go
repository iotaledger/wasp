package util

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"sort"
)

// SelectOutputsForAmount selects outputs out of given outputs just enough to transfer the amount
// outputs are selected in the descending order of balances with specified color
// the logic is to use small inputs first
// returns nil if outputs are not enough for the amount
// the set of resulting outputs must be DETERMINISTIC, despite map input and output
func SelectOutputsForAmount(outputs map[valuetransaction.OutputID][]*balance.Balance, color balance.Color, amount int64) map[valuetransaction.OutputID][]*balance.Balance {
	oids := make([]valuetransaction.OutputID, 0, len(outputs))
	for k := range outputs {
		oids = append(oids, k)
	}
	sort.Slice(oids, func(i, j int) bool {
		balsi := outputs[oids[i]]
		balsj := outputs[oids[j]]

		bi := BalanceOfColor(balsi, color)
		bj := BalanceOfColor(balsj, color)
		if bi == 0 {
			// if doesn't have color -> to the end of the list
			return false
		}
		if bi == 0 || bj == 0 {
			panic("bi == 0 || bj == 0")
		}
		// opposite to normal "less"
		switch {
		case bi > bj:
			return true
		case bi < bj:
			return false
		case bi == bj:
			return bytes.Compare(oids[i][:], oids[j][:]) < 0
		}
		panic("can't be")
	})

	ret := make(map[valuetransaction.OutputID][]*balance.Balance)
	sum := int64(0)
	for _, o := range oids {
		ret[o] = outputs[o]
		sum += BalanceOfColor(outputs[o], color)
		if sum >= amount {
			return ret
		}
	}
	return nil
}

func BalancesToString(outs map[valuetransaction.ID][]*balance.Balance) string {
	ret := ""
	for txid, bals := range outs {
		ret += txid.String() + ":\n"
		for _, bal := range bals {
			ret += fmt.Sprintf("         %s: %d\n", bal.Color().String(), bal.Value())
		}
	}
	return ret
}

func BalancesByColor(outs map[valuetransaction.ID][]*balance.Balance) (map[balance.Color]int64, int64) {
	ret := make(map[balance.Color]int64)
	var total int64
	for _, bals := range outs {
		for _, b := range bals {
			if s, ok := ret[b.Color()]; !ok {
				ret[b.Color()] = b.Value()
			} else {
				ret[b.Color()] = s + b.Value()
			}
			total += b.Value()
		}
	}
	return ret, total
}

func BalanceOfColor(bals []*balance.Balance, color balance.Color) int64 {
	sum := int64(0)
	for _, b := range bals {
		if b.Color() == color {
			sum += b.Value()
		}
	}
	return sum
}

func BalancesSumTotal(bals []*balance.Balance) int64 {
	var ret int64
	for _, b := range bals {
		ret += b.Value()
	}
	return ret
}

// BalancesHash calculates deterministic hash of address balances
func BalancesHash(outs map[valuetransaction.ID][]*balance.Balance) *hashing.HashValue {
	ids := make([]valuetransaction.ID, 0, len(outs))
	for txid := range outs {
		ids = append(ids, txid)
	}
	sort.Slice(ids, func(i, j int) bool {
		return bytes.Compare(ids[i][:], ids[j][:]) < 0
	})
	var buf bytes.Buffer
	for _, txid := range ids {
		buf.Write(txid[:])
		for _, b := range outs[txid] {
			buf.Write(b.Color().Bytes())
			_ = WriteUint64(&buf, uint64(b.Value()))
		}
	}
	return hashing.HashData(buf.Bytes())
}
