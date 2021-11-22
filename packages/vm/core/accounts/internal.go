package accounts

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

const (
	varStateAccounts    = "a"
	varStateTotalAssets = "t"
)

var ErrNotEnoughFunds = xerrors.New("not enough funds")

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

// CreditToAccount brings new funds to the on chain ledger
func CreditToAccount(state kv.KVStore, agentID *iscp.AgentID, assets *iscp.Assets) {
	mustCheckLedger(state, "CreditToAccountOld IN")
	defer mustCheckLedger(state, "CreditToAccountOld OUT")

	creditToAccount(state, getAccount(state, agentID), assets)
	creditToAccount(state, getTotalAssetsAccount(state), assets)
}

// creditToAccount internal
func creditToAccount(state kv.KVStore, account *collections.Map, transfer *iscp.Assets) {
	if transfer == nil {
		return
	}
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

func DebitFromAccount(state kv.KVStore, agentID *iscp.AgentID, assets *iscp.Assets) {
	mustCheckLedger(state, "DebitFromAccountOld IN")
	defer mustCheckLedger(state, "DebitFromAccountOld OUT")

	if !debitFromAccount(state, getAccount(state, agentID), assets) {
		panic(ErrNotEnoughFunds)
	}
	if !debitFromAccount(state, getTotalAssetsAccount(state), assets) {
		panic("debitFromAccount: inconsistent accounts ledger state")
	}
}

// debitFromAccount internal
func debitFromAccount(state kv.KVStore, account *collections.Map, assets *iscp.Assets) bool {
	if assets == nil {
		return true
	}
	defer touchAccount(state, account)

	current := getAccountBalances(account.Immutable())
	ok := true
	// deterministic order of iteration is not important here
	assets.ForEachRandomly(func(col colored.Color, transferAmount uint64) bool {
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
	current.ForEachRandomly(func(col colored.Color, bal uint64) bool {
		if bal > 0 {
			account.MustSetAt(col[:], util.Uint64To8Bytes(bal))
		} else {
			account.MustDelAt(col[:])
		}
		return true
	})
	return true
}

func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID *iscp.AgentID, transfer *iscp.Assets) bool {
	mustCheckLedger(state, "MoveBetweenAccounts.IN")
	defer mustCheckLedger(state, "MoveBetweenAccounts.OUT")
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

func MustMoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID *iscp.AgentID, transfer *iscp.Assets) {
	if !MoveBetweenAccounts(state, fromAgentID, toAgentID, transfer) {
		panic(ErrNotEnoughFunds)
	}
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

// GetIotaBalance return iota balance. 0 mean it does not exist
func GetIotaBalance(state kv.KVStoreReader, agentID *iscp.AgentID) uint64 {
	panic("not implemented")
}

// GetTokenBalance returns balance or nil if it does not exist
func GetTokenBalance(state kv.KVStoreReader, agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	panic("not implemented")
}

// GetAssets returns all assets owned by agentID. Returns nil if account does not exist
func GetAssets(state kv.KVStoreReader, agentID *iscp.AgentID) *iscp.Assets {
	panic("not implemented")
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
	account.MustIterateBalances(func(col colored.Color, bal uint64) bool {
		ret[col] = bal
		return true
	})
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
		agentID, err := iscp.AgentIDFromBytes(key)
		if err != nil {
			panic(err)
		}
		ret.AddAll(getAccountBalances(getAccountR(state, agentID)))
		return true
	})
	return ret
}

func mustCheckLedger(state kv.KVStore, checkpoint string) {
	a := GetTotalAssets(state)
	c := calcTotalAssets(state)
	if !a.Equals(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n total assets: %s\ncalc total: %s\n",
			checkpoint, a.String(), c.String()))
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

// TODO
func DecodeBalances(balances dict.Dict) (*iscp.Assets, error) {
	ret := colored.NewBalances()
	for col, bal := range balances {
		c, err := codec.DecodeColor([]byte(col))
		if err != nil {
			return nil, err
		}
		b, err := codec.DecodeUint64(bal)
		if err != nil {
			return nil, err
		}
		ret.Set(c, b)
	}
	return ret, nil
}

const postfixMaxAssumedNonceKey = "non"

func GetMaxAssumedNonce(state kv.KVStoreReader, address iotago.Address) uint64 {
	nonce, err := codec.DecodeUint64(state.MustGet(kv.Key(address.Bytes())+postfixMaxAssumedNonceKey), 0)
	if err != nil {
		panic(err)
	}
	return nonce
}

func RecordMaxAssumedNonce(state kv.KVStore, address iotago.Address, nonce uint64) {
	panic("not implemented")
}
