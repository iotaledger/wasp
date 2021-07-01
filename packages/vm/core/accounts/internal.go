package accounts

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	varStateAccounts    = "a"
	varStateTotalAssets = "t"
)

func getAccountsMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, varStateAccounts)
}

func getAccountsMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, varStateAccounts)
}

func getAccount(state kv.KVStore, agentID *coretypes.AgentID) *collections.Map {
	return collections.NewMap(state, string(agentID.Bytes()))
}

func getAccountR(state kv.KVStoreReader, agentID *coretypes.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, string(agentID.Bytes()))
}

func getTotalAssetsAccount(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, varStateTotalAssets)
}

func getTotalAssetsAccountR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, varStateTotalAssets)
}

// CreditToAccount brings new funds to the on chain ledger.
func CreditToAccount(state kv.KVStore, agentID *coretypes.AgentID, transfer *ledgerstate.ColoredBalances) {
	creditToAccount(state, getAccount(state, agentID), transfer)
	creditToAccount(state, getTotalAssetsAccount(state), transfer)
	mustCheckLedger(state, "CreditToAccount")
}

// creditToAccount internal
func creditToAccount(state kv.KVStore, account *collections.Map, transfer *ledgerstate.ColoredBalances) {
	if transfer == nil {
		return
	}
	defer touchAccount(state, account)

	transfer.ForEach(func(col ledgerstate.Color, bal uint64) bool {
		var currentBalance uint64
		v := account.MustGetAt(col[:])
		if v != nil {
			currentBalance = util.MustUint64From8Bytes(v)
		}
		account.MustSetAt(col[:], util.Uint64To8Bytes(currentBalance+bal))
		return true
	})
}

// DebitFromAccount removes funds from the chain ledger.
func DebitFromAccount(state kv.KVStore, agentID *coretypes.AgentID, transfer *ledgerstate.ColoredBalances) bool {
	if !debitFromAccount(state, getAccount(state, agentID), transfer) {
		return false
	}
	if !debitFromAccount(state, getTotalAssetsAccount(state), transfer) {
		panic("debitFromAccount: inconsistent accounts ledger state")
	}
	mustCheckLedger(state, "DebitFromAccount")
	return true
}

// debitFromAccount internal
func debitFromAccount(state kv.KVStore, account *collections.Map, transfer *ledgerstate.ColoredBalances) bool {
	if transfer == nil {
		return true
	}
	defer touchAccount(state, account)

	current := getAccountBalances(account.Immutable())

	ok := true
	transfer.ForEach(func(col ledgerstate.Color, transferAmount uint64) bool {
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
			account.MustSetAt(col[:], util.Uint64To8Bytes(rem))
		} else {
			account.MustDelAt(col[:])
		}
	}
	return true
}

func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID *coretypes.AgentID, transfer *ledgerstate.ColoredBalances) bool {
	if fromAgentID.Equals(toAgentID) {
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
	if account.Name() == varStateTotalAssets {
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

func GetBalance(state kv.KVStoreReader, agentID *coretypes.AgentID, color ledgerstate.Color) uint64 {
	b := getAccountR(state, agentID).MustGetAt(color[:])
	if b == nil {
		return 0
	}
	ret, _ := util.Uint64From8Bytes(b)
	return ret
}

func getAccountsIntern(state kv.KVStoreReader) dict.Dict {
	ret := dict.New()
	getAccountsMapR(state).MustIterate(func(agentID []byte, val []byte) bool {
		ret.Set(kv.Key(agentID), []byte{})
		return true
	})
	return ret
}

func getAccountBalances(account *collections.ImmutableMap) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	err := account.IterateBalances(func(col ledgerstate.Color, bal uint64) bool {
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
func GetAccountBalances(state kv.KVStoreReader, agentID *coretypes.AgentID) (map[ledgerstate.Color]uint64, bool) {
	account := getAccountR(state, agentID)
	if account.MustLen() == 0 {
		return nil, false
	}
	return getAccountBalances(account), true
}

func GetTotalAssets(state kv.KVStoreReader) *ledgerstate.ColoredBalances {
	return ledgerstate.NewColoredBalances(getAccountBalances(getTotalAssetsAccountR(state)))
}

func calcTotalAssets(state kv.KVStoreReader) *ledgerstate.ColoredBalances {
	ret := make(map[ledgerstate.Color]uint64)
	getAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := coretypes.NewAgentIDFromBytes(key)
		if err != nil {
			panic(err)
		}
		for col, b := range getAccountBalances(getAccountR(state, agentID)) {
			ret[col] += b
		}
		return true
	})
	return ledgerstate.NewColoredBalances(ret)
}

func mustCheckLedger(state kv.KVStore, checkpoint string) {
	a := GetTotalAssets(state)
	c := calcTotalAssets(state)
	if !coretypes.EqualColoredBalances(a, c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'", checkpoint))
	}
}

//nolint:unparam
func getAccountBalanceDict(ctx coretypes.SandboxView, account *collections.ImmutableMap, tag string) dict.Dict {
	balances := getAccountBalances(account)
	// ctx.Log().Debugf("%s. balance = %s\n", tag, balances.String())
	return EncodeBalances(balances)
}

func EncodeBalances(balances map[ledgerstate.Color]uint64) dict.Dict {
	ret := dict.New()
	for col, bal := range balances {
		ret.Set(kv.Key(col[:]), codec.EncodeUint64(bal))
	}
	return ret
}

func DecodeBalances(balances dict.Dict) (map[ledgerstate.Color]uint64, error) {
	ret := map[ledgerstate.Color]uint64{}
	for col, bal := range balances {
		c, _, err := codec.DecodeColor([]byte(col))
		if err != nil {
			return nil, err
		}
		b, _, err := codec.DecodeUint64(bal)
		if err != nil {
			return nil, err
		}
		ret[c] = b
	}
	return ret, nil
}

func GetOrder(state kv.KVStore, address ledgerstate.Address) uint64 {
	order, _, _ := codec.DecodeUint64(state.MustGet(kv.Key(address.Bytes()) + "ord"))
	return order
}

func SetOrder(state kv.KVStore, address ledgerstate.Address, order uint64) {
	state.Set(kv.Key(address.Bytes())+"ord", codec.EncodeUint64(order))
}
