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
func CreditToAccount(state kv.KVStore, agentID isc.AgentID, assets *isc.Assets) {
	if assets == nil || assets.IsEmpty() {
		return
	}
	creditToAccount(state, accountKey(agentID), assets)
	creditToAccount(state, l2TotalsAccount, assets)
	touchAccount(state, agentID)
}

// creditToAccount adds assets to the internal account map
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

// DebitFromAccount takes out assets balance the on chain ledger. If not enough it panics
func DebitFromAccount(state kv.KVStore, agentID isc.AgentID, assets *isc.Assets) {
	if assets == nil || assets.IsEmpty() {
		return
	}
	if !debitFromAccount(state, accountKey(agentID), assets) {
		panic(fmt.Errorf("debit from %s: %v\nassets: %s", agentID, ErrNotEnoughFunds, assets))
	}
	if !debitFromAccount(state, l2TotalsAccount, assets) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	touchAccount(state, agentID)
}

// debitFromAccount debits assets from the internal accounts map
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

func getFungibleTokens(state kv.KVStoreReader, accountKey kv.Key) *isc.Assets {
	ret := isc.NewEmptyAssets()
	ret.AddBaseTokens(getBaseTokens(state, accountKey))
	nativeTokensMapR(state, accountKey).MustIterate(func(idBytes []byte, val []byte) bool {
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
	allAccountsMapR(state).MustIterateKeys(func(key []byte) bool {
		ret.Add(getFungibleTokens(state, kv.Key(key)))
		return true
	})
	return ret
}

// GetAccountFungibleTokens returns all fungible tokens belonging to the agentID on the state
func GetAccountFungibleTokens(state kv.KVStoreReader, agentID isc.AgentID) *isc.Assets {
	return getFungibleTokens(state, accountKey(agentID))
}

func GetTotalL2FungibleTokens(state kv.KVStoreReader) *isc.Assets {
	return getFungibleTokens(state, l2TotalsAccount)
}

func getAccountBalanceDict(state kv.KVStoreReader, accountKey kv.Key) dict.Dict {
	return getFungibleTokens(state, accountKey).ToDict()
}
