package accounts

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// CreditToAccount brings new funds to the on chain ledger
// NOTE: this function does not take NFTs into account
func CreditToAccount(state kv.KVStore, agentID isc.AgentID, assets *isc.Assets, chainID isc.ChainID) {
	if assets == nil || assets.IsEmpty() {
		return
	}
	creditToAccount(state, accountKey(agentID, chainID), assets)
	creditToAccount(state, L2TotalsAccount, assets)
	touchAccount(state, agentID, chainID)
}

// creditToAccount adds assets to the internal account map
// NOTE: this function does not take NFTs into account
func creditToAccount(state kv.KVStore, accountKey kv.Key, assets *isc.Assets) {
	if assets == nil || assets.IsEmpty() {
		return
	}

	if assets.BaseTokens > 0 {
		setBaseTokens(state, accountKey, getBaseTokens(state, accountKey)+assets.BaseTokens)
	}
	for _, nt := range assets.NativeTokens {
		if nt.Amount.Sign() == 0 {
			continue
		}
		if nt.Amount.Sign() < 0 {
			panic(ErrBadAmount)
		}
		balance := getNativeTokenAmount(state, accountKey, nt.ID)
		balance.Add(balance, nt.Amount)
		if balance.Cmp(util.MaxUint256) > 0 {
			panic(ErrOverflow)
		}
		setNativeTokenAmount(state, accountKey, nt.ID, balance)
	}
}

func CreditToAccountFullDecimals(state kv.KVStore, agentID isc.AgentID, amount *big.Int, chainID isc.ChainID) {
	if !util.IsPositiveBigInt(amount) {
		return
	}
	creditToAccountFullDecimals(state, accountKey(agentID, chainID), amount)
	creditToAccountFullDecimals(state, L2TotalsAccount, amount)
	touchAccount(state, agentID, chainID)
}

// creditToAccountFullDecimals adds assets to the internal account map
// NOTE: this function does not take NFTs into account
func creditToAccountFullDecimals(state kv.KVStore, accountKey kv.Key, amount *big.Int) {
	setBaseTokensFullDecimals(state, accountKey, new(big.Int).Add(getBaseTokensFullDecimals(state, accountKey), amount))
}

// DebitFromAccount takes out assets balance the on chain ledger. If not enough it panics
// NOTE: this function does not take NFTs into account
func DebitFromAccount(state kv.KVStore, agentID isc.AgentID, assets *isc.Assets, chainID isc.ChainID) {
	if assets == nil || assets.IsEmpty() {
		return
	}
	if !debitFromAccount(state, accountKey(agentID, chainID), assets) {
		panic(fmt.Errorf("cannot debit (%s) from %s: %w", assets, agentID, ErrNotEnoughFunds))
	}
	if !debitFromAccount(state, L2TotalsAccount, assets) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	touchAccount(state, agentID, chainID)
}

// debitFromAccount debits assets from the internal accounts map
// NOTE: this function does not take NFTs into account
func debitFromAccount(state kv.KVStore, accountKey kv.Key, assets *isc.Assets) bool {
	if assets == nil || assets.IsEmpty() {
		return true
	}

	// first check, then mutate
	mutateBaseTokens := false
	mutations := isc.NewEmptyAssets()

	if assets.BaseTokens > 0 {
		balance := getBaseTokens(state, accountKey)
		if assets.BaseTokens > balance {
			return false
		}
		mutateBaseTokens = true
		mutations.BaseTokens = balance - assets.BaseTokens
	}
	for _, nt := range assets.NativeTokens {
		if nt.Amount.Sign() == 0 {
			continue
		}
		if nt.Amount.Sign() < 0 {
			panic(ErrBadAmount)
		}
		balance := getNativeTokenAmount(state, accountKey, nt.ID)
		balance = balance.Sub(balance, nt.Amount)
		if balance.Sign() < 0 {
			return false
		}
		mutations.AddNativeTokens(nt.ID, balance)
	}

	if mutateBaseTokens {
		setBaseTokens(state, accountKey, mutations.BaseTokens)
	}
	for _, nt := range mutations.NativeTokens {
		setNativeTokenAmount(state, accountKey, nt.ID, nt.Amount)
	}
	return true
}

// DebitFromAccountFullDecimals removes the amount from the chain ledger. If not enough it panics
func DebitFromAccountFullDecimals(state kv.KVStore, agentID isc.AgentID, amount *big.Int, chainID isc.ChainID) {
	if !util.IsPositiveBigInt(amount) {
		return
	}
	if !debitFromAccountFullDecimals(state, accountKey(agentID, chainID), amount) {
		panic(fmt.Errorf("cannot debit (%s) from %s: %w", amount.String(), agentID, ErrNotEnoughFunds))
	}

	if !debitFromAccountFullDecimals(state, L2TotalsAccount, amount) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	touchAccount(state, agentID, chainID)
}

// debitFromAccountFullDecimals debits the amount from the internal accounts map
func debitFromAccountFullDecimals(state kv.KVStore, accountKey kv.Key, amount *big.Int) bool {
	balance := getBaseTokensFullDecimals(state, accountKey)
	if balance.Cmp(amount) < 0 {
		return false
	}
	setBaseTokensFullDecimals(state, accountKey, new(big.Int).Sub(balance, amount))
	return true
}

func getFungibleTokens(state kv.KVStoreReader, accountKey kv.Key) *isc.Assets {
	ret := isc.NewEmptyAssets()
	ret.AddBaseTokens(getBaseTokens(state, accountKey))
	nativeTokensMapR(state, accountKey).Iterate(func(idBytes []byte, val []byte) bool {
		ret.AddNativeTokens(
			isc.MustNativeTokenIDFromBytes(idBytes),
			new(big.Int).SetBytes(val),
		)
		return true
	})
	return ret
}

func calcL2TotalFungibleTokens(state kv.KVStoreReader) *isc.Assets {
	ret := isc.NewEmptyAssets()
	allAccountsMapR(state).IterateKeys(func(key []byte) bool {
		ret.Add(getFungibleTokens(state, kv.Key(key)))
		return true
	})
	return ret
}

func calcL2TotalBaseTokensFullDecimals(state kv.KVStoreReader) *big.Int {
	ret := big.NewInt(0)
	allAccountsMapR(state).IterateKeys(func(key []byte) bool {
		amount := getBaseTokensFullDecimals(state, kv.Key(key))
		ret = new(big.Int).Add(ret, amount)
		return true
	})
	return ret
}

// GetAccountFungibleTokens returns all fungible tokens belonging to the agentID on the state
func GetAccountFungibleTokens(state kv.KVStoreReader, agentID isc.AgentID, chainID isc.ChainID) *isc.Assets {
	return getFungibleTokens(state, accountKey(agentID, chainID))
}

func GetTotalL2FungibleTokens(state kv.KVStoreReader) *isc.Assets {
	return getFungibleTokens(state, L2TotalsAccount)
}

func getAccountBalanceDict(state kv.KVStoreReader, accountKey kv.Key) dict.Dict {
	return getFungibleTokens(state, accountKey).ToDict()
}
