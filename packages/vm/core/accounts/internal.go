package accounts

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	VarStateAccounts    = "a"
	VarStateAllowances  = "l"
	VarStateTotalAssets = "t"
)

func getAccountsMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, VarStateAccounts)
}

func getAccountsMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, VarStateAccounts)
}

func getAccount(state kv.KVStore, agentID coretypes.AgentID) *collections.Map {
	return collections.NewMap(state, string(agentID[:]))
}

func getAccountR(state kv.KVStoreReader, agentID coretypes.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, string(agentID[:]))
}

func getTotalAssetsAccount(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, VarStateTotalAssets)
}

func getTotalAssetsAccountR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, VarStateTotalAssets)
}

// CreditToAccount brings new funds to the on chain ledger.
// Alone it is called when new funds arrive with the request, otherwise it called from MoveBetweenAccounts
func CreditToAccount(state kv.KVStore, agentID coretypes.AgentID, transfer coretypes.ColoredBalances) {
	//fmt.Printf("CreditToAccount: %s -- %s\n", agentID.String(), cbalances.Str(transfer))
	creditToAccount(state, getAccount(state, agentID), transfer)
	creditToAccount(state, getTotalAssetsAccount(state), transfer)
	MustCheckLedger(state, "CreditToAccount")
}

// creditToAccount internal
func creditToAccount(state kv.KVStore, account *collections.Map, transfer coretypes.ColoredBalances) {
	if transfer == nil || transfer.Len() == 0 {
		return
	}
	defer touchAccount(state, account)

	transfer.Iterate(func(col balance.Color, bal int64) bool {
		var currentBalance int64
		v := account.MustGetAt(col[:])
		if v != nil {
			currentBalance = int64(util.MustUint64From8Bytes(v))
		}
		account.MustSetAt(col[:], util.Uint64To8Bytes(uint64(currentBalance+bal)))
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
	MustCheckLedger(state, "DebitFromAccount")
	return true
}

// debitFromAccount internal
func debitFromAccount(state kv.KVStore, account *collections.Map, transfer coretypes.ColoredBalances) bool {
	if transfer == nil || transfer.Len() == 0 {
		return true
	}
	defer touchAccount(state, account)

	current := getAccountBalances(account.Immutable())

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
			account.MustSetAt(col[:], util.Uint64To8Bytes(uint64(rem)))
		} else {
			account.MustDelAt(col[:])
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

func touchAccount(state kv.KVStore, account *collections.Map) {
	if account.Name() == VarStateTotalAssets {
		return
	}
	agentid := []byte(account.Name())
	accounts := getAccountsMap(state)
	if account.MustLen() == 0 {
		accounts.MustDelAt(agentid)
	} else {
		accounts.MustSetAt(agentid, []byte{0xFF})
	}
}

func GetBalance(state kv.KVStoreReader, agentID coretypes.AgentID, color balance.Color) int64 {
	b := getAccountR(state, agentID).MustGetAt(color[:])
	if b == nil {
		return 0
	}
	ret, _ := util.Int64From8Bytes(b)
	return ret
}

func GetAccounts(state kv.KVStoreReader) dict.Dict {
	ret := dict.New()
	getAccountsMapR(state).MustIterate(func(agentID []byte, val []byte) bool {
		ret.Set(kv.Key(agentID), []byte{})
		return true
	})
	return ret
}

func getAccountBalances(account *collections.ImmutableMap) map[balance.Color]int64 {
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
func GetAccountBalances(state kv.KVStoreReader, agentID coretypes.AgentID) (map[balance.Color]int64, bool) {
	account := getAccountR(state, agentID)
	if account.MustLen() == 0 {
		return nil, false
	}
	return getAccountBalances(account), true
}

func GetTotalAssets(state kv.KVStoreReader) coretypes.ColoredBalances {
	return cbalances.NewFromMap(getAccountBalances(getTotalAssetsAccountR(state)))
}

func CalcTotalAssets(state kv.KVStoreReader) coretypes.ColoredBalances {
	ret := make(map[balance.Color]int64)
	getAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := coretypes.NewAgentIDFromBytes([]byte(key))
		if err != nil {
			panic(err)
		}
		for col, b := range getAccountBalances(getAccountR(state, agentID)) {
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

func getAccountBalanceDict(ctx coretypes.SandboxView, account *collections.ImmutableMap, tag string) dict.Dict {
	balances := getAccountBalances(account)
	ctx.Log().Debugf("%s. balance = %s\n", tag, cbalances.NewFromMap(balances).String())
	return EncodeBalances(balances)
}

func EncodeBalances(balances map[balance.Color]int64) dict.Dict {
	ret := dict.New()
	for col, bal := range balances {
		ret.Set(kv.Key(col[:]), codec.EncodeInt64(bal))
	}
	return ret
}

func DecodeBalances(balances dict.Dict) (map[balance.Color]int64, error) {
	ret := map[balance.Color]int64{}
	for col, bal := range balances {
		c, _, err := codec.DecodeColor([]byte(col))
		if err != nil {
			return nil, err
		}
		b, _, err := codec.DecodeInt64(bal)
		if err != nil {
			return nil, err
		}
		ret[c] = b
	}
	return ret, nil
}
