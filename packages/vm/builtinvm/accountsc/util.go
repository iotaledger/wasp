package accountsc

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

func CreditToAccount(state codec.MutableMustCodec, agentID coretypes.AgentID, transfer coretypes.ColoredBalances) {
	account := state.GetMap(kv.Key(agentID[:]))
	defer touchAccount(state, agentID)

	transfer.Iterate(func(col balance.Color, bal int64) bool {
		var currentBalance int64
		v := account.GetAt(col[:])
		if v != nil {
			currentBalance = int64(util.MustUint64From8Bytes(v))
		}
		account.SetAt(col[:], util.Uint64To8Bytes(uint64(currentBalance+bal)))
		return true
	})
}

func DebitFromAccount(state codec.MutableMustCodec, agentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	account := state.GetMap(kv.Key(agentID[:]))
	defer touchAccount(state, agentID)

	bals := make(map[balance.Color]int64)
	success := true
	account.Iterate(func(elemKey []byte, value []byte) bool {
		col, _, err := balance.ColorFromBytes(elemKey)
		if err != nil {
			success = false
			return false
		}
		bal, _ := util.Int64From8Bytes(value)
		b := transfer.Balance(col)
		if bal < b {
			success = false
			return false
		}
		bals[col] = bal - b
		return true
	})
	if !success {
		return false
	}
	for col, rem := range bals {
		if rem > 0 {
			account.SetAt(col[:], util.Uint64To8Bytes(uint64(rem)))
		} else {
			account.DelAt(col[:])
		}
	}
	return true
}

func MoveBetweenAccounts(state codec.MutableMustCodec, fromAgentID, toAgentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	if !DebitFromAccount(state, fromAgentID, transfer) {
		return false
	}
	CreditToAccount(state, toAgentID, transfer)
	return true
}

func touchAccount(state codec.MutableMustCodec, agentID coretypes.AgentID) {
	account := state.GetMap(kv.Key(agentID[:]))
	accounts := state.GetMap(VarStateAllAccounts)
	if account.Len() == 0 {
		accounts.DelAt(agentID[:])
	} else {
		accounts.SetAt(agentID[:], []byte{0xFF})
	}
}

func GetBalance(state codec.ImmutableMustCodec, agentID coretypes.AgentID, color balance.Color) int64 {
	b := state.GetMap(kv.Key(agentID[:])).GetAt(color[:])
	if b == nil {
		return 0
	}
	ret, _ := util.Int64From8Bytes(b)
	return ret
}

func GetAccounts(state codec.ImmutableMustCodec) codec.ImmutableCodec {
	ret := codec.NewCodec(dict.New())
	state.GetMap(VarStateAllAccounts).Iterate(func(elemKey []byte, val []byte) bool {
		ret.Set(kv.Key(elemKey), val)
		return true
	})
	return ret
}

func GetAccountBalances(state codec.ImmutableMustCodec, agentID coretypes.AgentID) (map[balance.Color]int64, bool) {
	s, ok := state.GetString("tmptest")
	fmt.Printf("account.GetString %s -- %v\n", s, ok)

	ret := make(map[balance.Color]int64)
	account := state.GetMap(kv.Key(agentID[:]))
	if account.Len() == 0 {
		return nil, false
	}
	account.Iterate(func(elemKey []byte, value []byte) bool {
		fmt.Printf("account.IterateBalances1 %v -- %v\n", elemKey, value)
		return true
	})
	err := account.IterateBalances(func(col balance.Color, bal int64) bool {
		fmt.Printf("account.IterateBalances2 col %s bal %d\n", col.String(), bal)
		ret[col] = bal
		return true
	})
	if err != nil {
		return nil, false
	}
	return ret, true
}
