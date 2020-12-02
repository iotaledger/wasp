package accountsc

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/coret/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// CreditToAccount brings new funds to the on chain ledger.
// Alone it is called when new funds arrive with the request, otherwise it called from MoveBetweenAccounts
func CreditToAccount(state codec.MutableMustCodec, agentID coret.AgentID, transfer coret.ColoredBalances) {
	//fmt.Printf("CreditToAccount: %s -- %s\n", agentID.String(), cbalances.Str(transfer))

	if agentID == TotalAssetsAccountID {
		// wrong account IDs
		return
	}
	creditToAccount(state, agentID, transfer)
	creditToAccount(state, TotalAssetsAccountID, transfer)
}

// creditToAccount internal
func creditToAccount(state codec.MutableMustCodec, agentID coret.AgentID, transfer coret.ColoredBalances) {
	if transfer == nil || transfer.Len() == 0 {
		return
	}
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

// DebitFromAccount removes funds from the chain ledger.
// Alone it is called when posting a request, otherwise it called from MoveBetweenAccounts
func DebitFromAccount(state codec.MutableMustCodec, agentID coret.AgentID, transfer coret.ColoredBalances) bool {
	//fmt.Printf("DebitFromAccount: %s -- %s\n", agentID.String(), cbalances.Str(transfer))

	if agentID == TotalAssetsAccountID {
		// wrong account IDs
		return false
	}
	if !debitFromAccount(state, agentID, transfer) {
		return false
	}
	if !debitFromAccount(state, TotalAssetsAccountID, transfer) {
		panic("debitFromAccount: inconsistent accounts ledger state")
	}
	return true
}

// debitFromAccount internal
func debitFromAccount(state codec.MutableMustCodec, agentID coret.AgentID, transfer coret.ColoredBalances) bool {
	if transfer == nil || transfer.Len() == 0 {
		return true
	}
	account := state.GetMap(kv.Key(agentID[:]))
	defer touchAccount(state, agentID)

	var err error
	var col balance.Color
	var bal int64
	current := make(map[balance.Color]int64)
	account.Iterate(func(elemKey []byte, value []byte) bool {
		if col, _, err = balance.ColorFromBytes(elemKey); err != nil {
			return false
		}
		if bal, err = util.Int64From8Bytes(value); err != nil {
			return false
		}
		current[col] = bal
		return true
	})
	if err != nil {
		return false
	}
	succ := true
	transfer.Iterate(func(col balance.Color, bal int64) bool {
		cur, _ := current[col]
		if cur < bal {
			succ = false
			return false
		}
		current[col] = cur - bal
		return true
	})
	if !succ {
		return false
	}
	for col, rem := range current {
		if rem > 0 {
			account.SetAt(col[:], util.Uint64To8Bytes(uint64(rem)))
		} else {
			account.DelAt(col[:])
		}
	}
	return true
}

func MoveBetweenAccounts(state codec.MutableMustCodec, fromAgentID, toAgentID coret.AgentID, transfer coret.ColoredBalances) bool {
	//fmt.Printf("MoveBetweenAccounts: %s -> %s -- %s\n", fromAgentID.String(), toAgentID.String(), cbalances.Str(transfer))

	if fromAgentID == toAgentID {
		// no need to move
		return true
	}
	// total assets account doesn't change
	if !debitFromAccount(state, fromAgentID, transfer) {
		return false
	}
	creditToAccount(state, toAgentID, transfer)
	return true
}

func touchAccount(state codec.MutableMustCodec, agentID coret.AgentID) {
	account := state.GetMap(kv.Key(agentID[:]))
	accounts := state.GetMap(VarStateAllAccounts)
	if account.Len() == 0 {
		accounts.DelAt(agentID[:])
	} else {
		accounts.SetAt(agentID[:], []byte{0xFF})
	}
}

func GetBalance(state codec.ImmutableMustCodec, agentID coret.AgentID, color balance.Color) int64 {
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

// GetAccountBalances returns all colored balances belonging to the agentID on the state.
// Normally, the state is the partition of the 'accountsc'
func GetAccountBalances(state codec.ImmutableMustCodec, agentID coret.AgentID) (map[balance.Color]int64, bool) {
	ret := make(map[balance.Color]int64)
	account := state.GetMap(kv.Key(agentID[:]))
	if account.Len() == 0 {
		return nil, false
	}
	err := account.IterateBalances(func(col balance.Color, bal int64) bool {
		ret[col] = bal
		return true
	})
	if err != nil {
		return nil, false
	}
	return ret, true
}

func GetTotalAssets(state codec.ImmutableMustCodec) coret.ColoredBalances {
	bals, ok := GetAccountBalances(state, TotalAssetsAccountID)
	if !ok {
		return cbalances.Nil
	}
	return cbalances.NewFromMap(bals)
}

func CalcTotalAssets(state codec.ImmutableMustCodec) coret.ColoredBalances {
	accounts := GetAccounts(state)
	retMap := make(map[balance.Color]int64)
	var agentID coret.AgentID
	var err error
	err = accounts.IterateKeys("", func(key kv.Key) bool {
		agentID, err = coret.NewAgentIDFromBytes([]byte(key))
		if err != nil {
			return false
		}
		if agentID == TotalAssetsAccountID {
			return true
		}
		balMap, ok := GetAccountBalances(state, agentID)
		if !ok {
			return true
		}
		for col, b := range balMap {
			s, _ := retMap[col]
			retMap[col] = s + b
		}
		return true
	})
	if err != nil {
		panic(err)
	}
	return cbalances.NewFromMap(retMap)
}

func MustCheckLedger(state codec.ImmutableMustCodec, checkpoint string) {
	//fmt.Printf("--------------- check ledger checkpoint: '%s'\n", checkpoint)
	a := GetTotalAssets(state)
	c := CalcTotalAssets(state)
	if !a.Equal(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'", checkpoint))
	}
}
