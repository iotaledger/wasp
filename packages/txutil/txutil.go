package txutil

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

func TotalBalanceOfInputs(outs map[valuetransaction.OutputID][]*balance.Balance) int64 {
	ret := int64(0)
	for _, bals := range outs {
		for _, b := range bals {
			ret += b.Value
		}
	}
	return ret
}

func BalancesToString(outs map[valuetransaction.ID][]*balance.Balance) string {
	if outs == nil {
		return "empty balances"
	}

	txids := make([]valuetransaction.ID, 0, len(outs))
	for txid := range outs {
		txids = append(txids, txid)
	}
	sort.Slice(txids, func(i, j int) bool {
		return bytes.Compare(txids[i][:], txids[j][:]) < 0
	})

	ret := ""
	for _, txid := range txids {
		bals := outs[txid]
		ret += txid.String() + ":\n"
		for _, bal := range bals {
			ret += fmt.Sprintf("         %s: %d\n", bal.Color.String(), bal.Value)
		}
	}
	return ret
}

func InputBalancesToString(outs map[valuetransaction.OutputID][]*balance.Balance) string {
	if outs == nil {
		return "empty balances"
	}

	oids := make([]valuetransaction.OutputID, 0, len(outs))
	for oid := range outs {
		oids = append(oids, oid)
	}
	sort.Slice(oids, func(i, j int) bool {
		return bytes.Compare(oids[i][:], oids[j][:]) < 0
	})

	ret := ""
	for _, oid := range oids {
		bals := outs[oid]
		ret += oid.Address().String() + "-" + oid.TransactionID().String() + ":\n"
		for _, bal := range bals {
			ret += fmt.Sprintf("         %s: %d\n", bal.Color.String(), bal.Value)
		}
	}
	return ret
}

func BalancesByColor(outs map[valuetransaction.ID][]*balance.Balance) (map[balance.Color]int64, int64) {
	ret := make(map[balance.Color]int64)
	var total int64
	for _, bals := range outs {
		for _, b := range bals {
			if s, ok := ret[b.Color]; !ok {
				ret[b.Color] = b.Value
			} else {
				ret[b.Color] = s + b.Value
			}
			total += b.Value
		}
	}
	return ret, total
}

func OutputBalancesByColor(outs map[valuetransaction.OutputID][]*balance.Balance) (map[balance.Color]int64, int64) {
	ret := make(map[balance.Color]int64)
	var total int64
	for _, bals := range outs {
		for _, b := range bals {
			if s, ok := ret[b.Color]; !ok {
				ret[b.Color] = b.Value
			} else {
				ret[b.Color] = s + b.Value
			}
			total += b.Value
		}
	}
	return ret, total
}

func BalanceOfColor(bals []*balance.Balance, color balance.Color) int64 {
	sum := int64(0)
	for _, b := range bals {
		if b.Color == color {
			sum += b.Value
		}
	}
	return sum
}

func CloneBalances(bals []*balance.Balance) []*balance.Balance {
	ret := make([]*balance.Balance, len(bals))
	for i := range ret {
		ret[i] = balance.New(bals[i].Color, bals[i].Value)
	}
	return ret
}

func BalancesSumTotal(bals []*balance.Balance) int64 {
	var ret int64
	for _, b := range bals {
		ret += b.Value
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
			buf.Write(b.Color.Bytes())
			_ = util.WriteUint64(&buf, uint64(b.Value))
		}
	}
	return hashing.HashData(buf.Bytes())
}

func InputsToStringByAddress(inputs *valuetransaction.Inputs) string {
	imap := make(map[string][]string)
	inputs.ForEach(func(oid valuetransaction.OutputID) bool {
		a := oid.Address().String()
		m, ok := imap[a]
		if !ok {
			m = make([]string, 0)
		}
		m = append(m, oid.TransactionID().String())
		imap[a] = m
		return true
	})
	ret := ""
	for a, m := range imap {
		ret += a + ":\n"
		for _, t := range m {
			ret += "    " + t + "\n"
		}
	}
	return ret
}

func AddressesToStrings(addrs []address.Address) []string {
	return addressesToStrings(addrs, false)
}

func AddressesToStringsShort(addrs []address.Address) []string {
	return addressesToStrings(addrs, true)
}

func addressesToStrings(addrs []address.Address, short bool) []string {
	ret := make([]string, len(addrs))
	for i := range ret {
		if short {
			ret[i] = addrs[i].String()[:6] + ".."
		} else {
			ret[i] = addrs[i].String()
		}
	}
	return ret
}
