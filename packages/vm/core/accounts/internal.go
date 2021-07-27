package accounts

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
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

func getAccount(state kv.KVStore, agentID *iscp.AgentID) *collections.Map {
	return collections.NewMap(state, string(agentID.Bytes()))
}

func getAccountR(state kv.KVStoreReader, agentID *iscp.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, string(agentID.Bytes()))
}

func getTotalAssetsAccount(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, varStateTotalAssets)
}

func getTotalAssetsAccountR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, varStateTotalAssets)
}

// CreditToAccount brings new funds to the on chain ledger.
func CreditToAccount(state kv.KVStore, agentID *iscp.AgentID, transfer colored.Balances) {
	creditToAccount(state, getAccount(state, agentID), transfer)
	creditToAccount(state, getTotalAssetsAccount(state), transfer)
	mustCheckLedger(state, "CreditToAccount")
}

// creditToAccount internal
func creditToAccount(state kv.KVStore, account *collections.Map, transfer colored.Balances) {
	defer touchAccount(state, account)

	// deterministic order of iteration is not important here
	transfer.ForEachRandomly(func(col colored.Color, bal uint64) bool {
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
func DebitFromAccount(state kv.KVStore, agentID *iscp.AgentID, transfer colored.Balances) bool {
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
func debitFromAccount(state kv.KVStore, account *collections.Map, transfer colored.Balances) bool {
	defer touchAccount(state, account)

	current := getAccountBalances(account.Immutable())
	ok := true
	// deterministic order of iteration is not important here
	transfer.ForEachRandomly(func(col colored.Color, transferAmount uint64) bool {
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

func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID *iscp.AgentID, transfer colored.Balances) bool {
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

func GetBalance(state kv.KVStoreReader, agentID *iscp.AgentID, col colored.Color) uint64 {
	b := getAccountR(state, agentID).MustGetAt(col[:])
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

func getAccountBalances(account *collections.ImmutableMap) colored.Balances {
	ret := colored.NewBalances()
	err := account.IterateBalances(func(col colored.Color, bal uint64) bool {
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
func GetAccountBalances(state kv.KVStoreReader, agentID *iscp.AgentID) (colored.Balances, bool) {
	account := getAccountR(state, agentID)
	if account.MustLen() == 0 {
		return nil, false
	}
	return getAccountBalances(account), true
}

func GetTotalAssets(state kv.KVStoreReader) colored.Balances {
	return getAccountBalances(getTotalAssetsAccountR(state))
}

func calcTotalAssets(state kv.KVStoreReader) colored.Balances {
	ret := colored.NewBalances()
	getAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := iscp.NewAgentIDFromBytes(key)
		if err != nil {
			panic(err)
		}
		for col, b := range getAccountBalances(getAccountR(state, agentID)) {
			ret.Add(col, b)
		}
		return true
	})
	return ret
}

func mustCheckLedger(state kv.KVStore, checkpoint string) {
	a := GetTotalAssets(state)
	c := calcTotalAssets(state)
	if !a.Equals(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'", checkpoint))
	}
}

//nolint:unparam
func getAccountBalanceDict(ctx iscp.SandboxView, account *collections.ImmutableMap, tag string) dict.Dict {
	balances := getAccountBalances(account)
	// ctx.Log().Debugf("%s. balance = %s\n", tag, balances.String())
	return EncodeBalances(balances)
}

func EncodeBalances(balances colored.Balances) dict.Dict {
	ret := dict.New()
	for col, bal := range balances {
		ret.Set(kv.Key(col[:]), codec.EncodeUint64(bal))
	}
	return ret
}

func DecodeBalances(balances dict.Dict) (colored.Balances, error) {
	ret := colored.NewBalances()
	for col, bal := range balances {
		c, _, err := codec.DecodeColor([]byte(col))
		if err != nil {
			return nil, err
		}
		b, _, err := codec.DecodeUint64(bal)
		if err != nil {
			return nil, err
		}
		ret.Set(c, b)
	}
	return ret, nil
}

const postfixMaxAssumedNonceKey = "non"

func GetMaxAssumedNonce(state kv.KVStore, address ledgerstate.Address) uint64 {
	nonce, _, _ := codec.DecodeUint64(state.MustGet(kv.Key(address.Bytes()) + postfixMaxAssumedNonceKey))
	return nonce
}

func RecordMaxAssumedNonce(state kv.KVStore, address ledgerstate.Address, nonce uint64) {
	next := GetMaxAssumedNonce(state, address) + 1
	if nonce > next {
		next = nonce
	}
	state.Set(kv.Key(address.Bytes())+postfixMaxAssumedNonceKey, codec.EncodeUint64(next))
}
