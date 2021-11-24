package accounts

import (
	"bytes"
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
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
	mustCheckLedger(state, "CreditToAccount IN")
	defer mustCheckLedger(state, "CreditToAccount OUT")

	creditToAccount(state, getAccount(state, agentID), assets)
	creditToAccount(state, getTotalAssetsAccount(state), assets)
}

// creditToAccount internal
func creditToAccount(state kv.KVStore, account *collections.Map, transfer *iscp.Assets) {
	if transfer == nil {
		return
	}
	defer touchAccount(state, account)

	creditAsset(account, iscp.IOTA_TOKEN_ID, new(big.Int).SetUint64(transfer.Iotas))
	for _, token := range transfer.Tokens {
		creditAsset(account, token.ID[:], token.Amount)
	}
}

func creditAsset(account *collections.Map, assetID []byte, amount *big.Int) {
	if amount.Cmp(big.NewInt(0)) == -1 {
		return // cannot credit negative amounts. // TODO should it panic here?, or should this check be removed?
	}
	balance := big.NewInt(0)
	v := account.MustGetAt(assetID[:])
	if v != nil {
		balance.SetBytes(v)
	}
	balance.Add(balance, amount)
	account.MustSetAt(assetID[:], balance.Bytes())
}

func getAccountAssets(state kv.KVStore, agentID *iscp.AgentID) *iscp.Assets {
	return nil
}

func DebitFromAccount(state kv.KVStore, agentID *iscp.AgentID, assets *iscp.Assets) {
	mustCheckLedger(state, "DebitFromAccount IN")
	defer mustCheckLedger(state, "DebitFromAccount OUT")

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

	// assert there is enough balance in the account
	if !hasEnoughBalance(account, iscp.IOTA_TOKEN_ID, new(big.Int).SetUint64(assets.Iotas)) {
		return false
	}
	for _, token := range assets.Tokens {
		if !hasEnoughBalance(account, token.ID[:], token.Amount) {
			return false
		}
	}

	// debit from the account
	debitAsset(account, iscp.IOTA_TOKEN_ID, new(big.Int).SetUint64(assets.Iotas))
	for _, token := range assets.Tokens {
		debitAsset(account, token.ID[:], token.Amount)
	}
	return true
}

func hasEnoughBalance(account *collections.Map, assetID []byte, amount *big.Int) bool {
	v := account.MustGetAt(assetID[:])
	if v == nil {
		return false
	}
	balance := new(big.Int).SetBytes(account.MustGetAt(assetID[:]))
	return balance.Cmp(amount) == 1 || balance.Cmp(amount) == 0
}

func debitAsset(account *collections.Map, assetID []byte, amount *big.Int) {
	balance := new(big.Int).SetBytes(account.MustGetAt(assetID[:]))
	balance.Sub(balance, amount)
	if balance.Cmp(big.NewInt(0)) == 0 {
		account.MustDelAt(assetID[:]) // remove asset entry if balance is empty
	}
	account.MustSetAt(assetID[:], balance.Bytes())
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

func getAccountBalances(account *collections.ImmutableMap) *iscp.Assets {
	ret := iscp.NewEmptyAssets()
	account.MustIterate(func(assetID []byte, val []byte) bool {
		if bytes.Compare(assetID, iscp.IOTA_TOKEN_ID) == 0 {
			ret.Iotas = new(big.Int).SetBytes(val).Uint64()
			return true
		}
		var tokenID iotago.NativeTokenID
		copy(tokenID[:], assetID)
		token := iotago.NativeToken{
			ID:     tokenID,
			Amount: new(big.Int).SetBytes(val),
		}
		ret.Tokens = append(ret.Tokens, &token)
		return true
	})
	return ret
}

// GetAccountBalances returns all assets belonging to the agentID on the state.
// Normally, the state is the partition of the 'accountsc'
func GetAccountBalances(state kv.KVStoreReader, agentID *iscp.AgentID) (*iscp.Assets, bool) {
	account := getAccountR(state, agentID)
	if account.MustLen() == 0 {
		return nil, false
	}
	return getAccountBalances(account), true
}

func GetTotalAssets(state kv.KVStoreReader) *iscp.Assets {
	return getAccountBalances(getTotalAssetsAccountR(state))
}

func calcTotalAssets(state kv.KVStoreReader) *iscp.Assets {
	ret := iscp.NewEmptyAssets()

	getAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := iscp.AgentIDFromBytes(key)
		if err != nil {
			panic(err)
		}
		accBalances := getAccountBalances(getAccountR(state, agentID))
		ret.Add(accBalances)
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

func getAccountBalanceDict(ctx iscp.SandboxView, account *collections.ImmutableMap, tag string) dict.Dict {
	return getAccountBalances(account).ToDict()
}

func DecodeBalances(balances dict.Dict) (*iscp.Assets, error) {
	return iscp.NewAssetsFromDict(balances)
}

const postfixMaxAssumedNonceKey = "non"

func GetMaxAssumedNonce(state kv.KVStoreReader, address iotago.Address) uint64 {
	return 0
	// TODO refactor with BytesFromAddress util func
	// nonce, err := codec.DecodeUint64(state.MustGet(kv.Key(address.Bytes())+postfixMaxAssumedNonceKey), 0)
	// if err != nil {
	// 	panic(err)
	// }
	// return nonce
}

func RecordMaxAssumedNonce(state kv.KVStore, address iotago.Address, nonce uint64) {
	panic("not implemented")
}
