package accounts

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
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
	varStateUtxoMapping = "u"
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
func creditToAccount(state kv.KVStore, account *collections.Map, assets *iscp.Assets) {
	if assets == nil || (assets.Iotas == 0 && len(assets.Tokens) == 0) {
		return
	}
	defer touchAccount(state, account)

	creditAsset(account, iscp.IotaAssetID, new(big.Int).SetUint64(assets.Iotas))
	for _, token := range assets.Tokens {
		creditAsset(account, token.ID[:], token.Amount)
	}
}

func creditAsset(account *collections.Map, assetID []byte, amount *big.Int) {
	if amount.Cmp(big.NewInt(0)) == -1 {
		return // cannot credit negative amounts. // TODO should it panic here?, or should this check be removed?
	}
	balance := big.NewInt(0)
	v := account.MustGetAt(assetID)
	if v != nil {
		balance.SetBytes(v)
	}
	balance.Add(balance, amount)
	account.MustSetAt(assetID, balance.Bytes())
}

func DebitFromAccount(state kv.KVStore, agentID *iscp.AgentID, assets *iscp.Assets) {
	mustCheckLedger(state, "DebitFromAccount IN")
	defer mustCheckLedger(state, "DebitFromAccount OUT")

	if !debitFromAccount(state, getAccount(state, agentID), assets, false) {
		panic(ErrNotEnoughFunds)
	}
	if !debitFromAccount(state, getTotalAssetsAccount(state), assets, true) {
		panic("debitFromAccount: inconsistent accounts ledger state")
	}
}

// debitFromAccount internal
func debitFromAccount(state kv.KVStore, account *collections.Map, assets *iscp.Assets, isTotalAssetsAccount bool) bool {
	if assets == nil || (assets.Iotas == 0 && len(assets.Tokens) == 0) {
		return true
	}
	defer touchAccount(state, account)

	// assert there is enough balance in the account
	if !hasEnoughBalance(account.ImmutableMap, iscp.IotaAssetID, new(big.Int).SetUint64(assets.Iotas)) {
		return false
	}
	for _, token := range assets.Tokens {
		if !hasEnoughBalance(account.ImmutableMap, token.ID[:], token.Amount) {
			return false
		}
	}

	// debit from the account
	debitAsset(account, iscp.IotaAssetID, new(big.Int).SetUint64(assets.Iotas))
	for _, token := range assets.Tokens {
		debitAsset(account, token.ID[:], token.Amount)

		// delete UTXO mapping when the chain no longer holds more of this asset
		if !isTotalAssetsAccount {
			continue
		}
		remainingChainAssetBalance := getAccountAssetBalance(account.ImmutableMap, token.ID[:])
		if util.IsZeroBigInt(remainingChainAssetBalance) {
			removeUTXOMapping(state, token.ID)
		}
	}
	return true
}

func hasEnoughBalance(account *collections.ImmutableMap, assetID []byte, amount *big.Int) bool {
	balance := getAccountAssetBalance(account, assetID)
	if balance == nil {
		return false
	}
	return balance.Cmp(amount) == 1 || balance.Cmp(amount) == 0
}

func getAccountAssetBalance(account *collections.ImmutableMap, assetID []byte) *big.Int {
	v := account.MustGetAt(assetID)
	if v == nil {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(account.MustGetAt(assetID))
}

func debitAsset(account *collections.Map, assetID []byte, amount *big.Int) {
	balance := new(big.Int).SetBytes(account.MustGetAt(assetID))
	balance.Sub(balance, amount)
	if util.IsZeroBigInt(balance) {
		account.MustDelAt(assetID) // remove asset entry if balance is empty
		return
	}
	account.MustSetAt(assetID, balance.Bytes())
}

func MoveBetweenAccounts(state kv.KVStore, fromAgentID, toAgentID *iscp.AgentID, transfer *iscp.Assets) bool {
	mustCheckLedger(state, "MoveBetweenAccounts.IN")
	defer mustCheckLedger(state, "MoveBetweenAccounts.OUT")
	if fromAgentID.Equals(toAgentID) {
		// no need to move
		return true
	}
	// total assets account doesn't change
	if !debitFromAccount(state, getAccount(state, fromAgentID), transfer, false) {
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
	acc := getAccountR(state, agentID)
	return getAccountAssetBalance(acc, iscp.IotaAssetID).Uint64()
}

// GetTokenBalance returns balance or nil if it does not exist
func GetTokenBalance(state kv.KVStoreReader, agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	acc := getAccountR(state, agentID)
	balance := getAccountAssetBalance(acc, tokenID[:])
	if balance == nil {
		return big.NewInt(0)
	}
	return balance
}

// GetTokenBalanceTotal return total of the native token on-chain
func GetTokenBalanceTotal(state kv.KVStoreReader, tokenID *iotago.NativeTokenID) *big.Int {
	panic("not implemented")
}

// GetAssets returns all assets owned by agentID. Returns nil if account does not exist
func GetAssets(state kv.KVStoreReader, agentID *iscp.AgentID) *iscp.Assets {
	acc := getAccountR(state, agentID)
	ret := iscp.NewEmptyAssets()
	acc.MustIterate(func(k []byte, v []byte) bool {
		if iscp.IsIota(k) {
			ret.Iotas = new(big.Int).SetBytes(v).Uint64()
			return true
		}
		token := iotago.NativeToken{
			ID:     iscp.TokenIDFromAssetID(k),
			Amount: new(big.Int).SetBytes(v),
		}
		ret.Tokens = append(ret.Tokens, &token)
		return true
	})
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

func getAccountAssets(account *collections.ImmutableMap) *iscp.Assets {
	ret := iscp.NewEmptyAssets()
	account.MustIterate(func(assetID []byte, val []byte) bool {
		if iscp.IsIota(assetID) {
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

// GetAccountAssets returns all assets belonging to the agentID on the state.
// Normally, the state is the partition of the 'accountsc'
func GetAccountAssets(state kv.KVStoreReader, agentID *iscp.AgentID) (*iscp.Assets, bool) {
	account := getAccountR(state, agentID)
	if account.MustLen() == 0 {
		return nil, false
	}
	return getAccountAssets(account), true
}

func GetTotalAssets(state kv.KVStoreReader) *iscp.Assets {
	return getAccountAssets(getTotalAssetsAccountR(state))
}

func calcTotalAssets(state kv.KVStoreReader) *iscp.Assets {
	ret := iscp.NewEmptyAssets()

	getAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		agentID, err := iscp.AgentIDFromBytes(key)
		if err != nil {
			panic(err)
		}
		accBalances := getAccountAssets(getAccountR(state, agentID))
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

func getAccountBalanceDict(account *collections.ImmutableMap) dict.Dict {
	return getAccountAssets(account).ToDict()
}

func DecodeBalances(balances dict.Dict) (*iscp.Assets, error) {
	return iscp.NewAssetsFromDict(balances)
}

const postfixMaxAssumedNonceKey = "non"

func GetMaxAssumedNonce(state kv.KVStoreReader, address iotago.Address) uint64 {
	nonce, err := codec.DecodeUint64(
		state.MustGet(kv.Key(iscp.BytesFromAddress(address))+postfixMaxAssumedNonceKey),
		0,
	)
	if err != nil {
		panic(err)
	}
	return nonce
}

func RecordMaxAssumedNonce(state kv.KVStore, address iotago.Address, nonce uint64) {
	next := GetMaxAssumedNonce(state, address) + 1
	if nonce > next {
		next = nonce
	}
	state.Set(kv.Key(iscp.BytesFromAddress(address))+postfixMaxAssumedNonceKey, codec.EncodeUint64(next))
}

func GetUtxoMapping(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, varStateUtxoMapping)
}

func SetAssetsUtxoIndices(state kv.KVStore, stateIndex uint32, tokenUtxoIndices []iotago.NativeTokenID) {
	mapping := GetUtxoMapping(state)
	for index, assetID := range tokenUtxoIndices {
		entry := codec.EncodeUint16(uint16(index))
		entry = append(entry, codec.EncodeUint32(stateIndex)...)
		mapping.MustSetAt(assetID[:], entry)
	}
}

func GetUtxoForAsset(state kv.KVStore, id iotago.NativeTokenID) (stateIndex uint32, outputIndex uint16, err error) {
	mapping := GetUtxoMapping(state)
	entry := mapping.MustGetAt(id[:])
	outputIndex, err = codec.DecodeUint16(entry[:2])
	if err != nil {
		return 0, 0, err
	}
	stateIndex, err = codec.DecodeUint32(entry[2:])
	if err != nil {
		return 0, 0, err
	}
	return stateIndex, outputIndex, nil
}

func removeUTXOMapping(state kv.KVStore, id iotago.NativeTokenID) {
	mapping := GetUtxoMapping(state)
	mapping.DelAt(id[:]) //nolint:errcheck // No need to check this error
}
