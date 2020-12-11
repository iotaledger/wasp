package accountsc

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	VarStateAccounts    = "a"
	VarStateAllowances  = "l"
	VarStateTotalAssets = "t"
)

func getAccountsMap(state kv.KVStore) *datatypes.MustMap {
	return datatypes.NewMustMap(state, VarStateAccounts)
}

func getAccount(state kv.KVStore, agentID coretypes.AgentID) *datatypes.MustMap {
	return datatypes.NewMustMap(state, string(agentID[:]))
}

func getTotalAssetsAccount(state kv.KVStore) *datatypes.MustMap {
	return datatypes.NewMustMap(state, VarStateTotalAssets)
}

// CreditToAccount brings new funds to the on chain ledger.
// Alone it is called when new funds arrive with the request, otherwise it called from MoveBetweenAccounts
func CreditToAccount(state kv.KVStore, agentID coretypes.AgentID, transfer coretypes.ColoredBalances) {
	//fmt.Printf("CreditToAccount: %s -- %s\n", agentID.String(), cbalances.Str(transfer))
	creditToAccount(state, getAccount(state, agentID), transfer)
	creditToAccount(state, getTotalAssetsAccount(state), transfer)
}

// creditToAccount internal
func creditToAccount(state kv.KVStore, account *datatypes.MustMap, transfer coretypes.ColoredBalances) {
	if transfer == nil || transfer.Len() == 0 {
		return
	}
	defer touchAccount(state, account)

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
func DebitFromAccount(state kv.KVStore, agentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	//fmt.Printf("DebitFromAccount: %s -- %s\n", agentID.String(), cbalances.Str(transfer))

	if !debitFromAccount(state, getAccount(state, agentID), transfer) {
		return false
	}
	if !debitFromAccount(state, getTotalAssetsAccount(state), transfer) {
		panic("debitFromAccount: inconsistent accounts ledger state")
	}
	return true
}

// debitFromAccount internal
func debitFromAccount(state kv.KVStore, account *datatypes.MustMap, transfer coretypes.ColoredBalances) bool {
	if transfer == nil || transfer.Len() == 0 {
		return true
	}
	defer touchAccount(state, account)

	current := getAccountBalances(account)

	ok := true
	transfer.Iterate(func(col balance.Color, transferAmount int64) bool {
		bal := current[col]
		if bal < transferAmount {
			ok = false
			return false
		}
		current[col] = bal - transferAmount
		return true
	})
	if !ok {
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

func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID coretypes.AgentID, transfer coretypes.ColoredBalances) bool {
	//fmt.Printf("MoveBetweenAccounts: %s -> %s -- %s\n", fromAgentID.String(), toAgentID.String(), cbalances.Str(transfer))

	if fromAgentID == toAgentID {
		// no need to move
		return true
	}
	// total assets account doesn't change
	if !debitFromAccount(state, getAccount(state, fromAgentID), transfer) {
		return false
	}
	creditToAccount(state, getAccount(state, toAgentID), transfer)
	return true
}

func touchAccount(state kv.KVStore, account *datatypes.MustMap) {
	if account.Name() == VarStateTotalAssets {
		return
	}
	agentid := []byte(account.Name())
	accounts := getAccountsMap(state)
	if account.Len() == 0 {
		accounts.DelAt(agentid)
	} else {
		accounts.SetAt(agentid, []byte{0xFF})
	}
}

func GetBalance(state kv.KVStore, agentID coretypes.AgentID, color balance.Color) int64 {
	b := getAccount(state, agentID).GetAt(color[:])
	if b == nil {
		return 0
	}
	ret, _ := util.Int64From8Bytes(b)
	return ret
}

func GetAccounts(state kv.KVStore) dict.Dict {
	ret := dict.New()
	getAccountsMap(state).Iterate(func(agentID []byte, val []byte) bool {
		ret.Set(kv.Key(agentID), []byte{})
		return true
	})
	return ret
}

func getAccountBalances(account *datatypes.MustMap) map[balance.Color]int64 {
	ret := make(map[balance.Color]int64)
	err := account.IterateBalances(func(col balance.Color, bal int64) bool {
		ret[col] = bal
		return true
	})
	if err != nil {
		panic(err)
	}
	return ret
}

// GetAccountBalances returns all colored balances belonging to the agentID on the state.
// Normally, the state is the partition of the 'accountsc'
func GetAccountBalances(state kv.KVStore, agentID coretypes.AgentID) (map[balance.Color]int64, bool) {
	account := getAccount(state, agentID)
	if account.Len() == 0 {
		return nil, false
	}
	return getAccountBalances(account), true
}

func GetTotalAssets(state kv.KVStore) coretypes.ColoredBalances {
	return cbalances.NewFromMap(getAccountBalances(getTotalAssetsAccount(state)))
}

func CalcTotalAssets(state kv.KVStore) coretypes.ColoredBalances {
	ret := make(map[balance.Color]int64)
	getAccountsMap(state).IterateKeys(func(key []byte) bool {
		agentID, err := coretypes.NewAgentIDFromBytes([]byte(key))
		if err != nil {
			panic(err)
		}
		for col, b := range getAccountBalances(getAccount(state, agentID)) {
			ret[col] = ret[col] + b
		}
		return true
	})
	return cbalances.NewFromMap(ret)
}

func MustCheckLedger(state kv.KVStore, checkpoint string) {
	//fmt.Printf("--------------- check ledger checkpoint: '%s'\n", checkpoint)
	a := GetTotalAssets(state)
	c := CalcTotalAssets(state)
	if !a.Equal(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'", checkpoint))
	}
}

func getAccountBalanceDict(ctx vmtypes.SandboxView, account *datatypes.MustMap, event string) dict.Dict {
	balances := getAccountBalances(account)
	ctx.Log().Debugf("%s. balance = %s\n", event, cbalances.NewFromMap(balances).String())
	ret := dict.New()
	for col, bal := range balances {
		ret.Set(kv.Key(col[:]), codec.EncodeInt64(bal))
	}
	return ret
}
