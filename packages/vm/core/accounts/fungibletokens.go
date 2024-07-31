package accounts

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/bigint"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

// CreditToAccount brings new funds to the on chain ledger
// NOTE: this function does not take NFTs into account
func (s *StateWriter) CreditToAccount(agentID isc.AgentID, assets *isc.Assets, chainID isc.ChainID) {
	if assets == nil || assets.IsEmpty() {
		return
	}
	s.creditToAccount(accountKey(agentID, chainID), assets)
	s.creditToAccount(L2TotalsAccount, assets)
	s.touchAccount(agentID, chainID)
}

// creditToAccount adds assets to the internal account map
// NOTE: this function does not take NFTs into account
func (s *StateWriter) creditToAccount(accountKey kv.Key, assets *isc.Assets) {
	if assets == nil || assets.IsEmpty() {
		return
	}

	if assets.BaseTokens > 0 {
		incomingTokensFullDecimals := util.BaseTokensDecimalsToEthereumDecimals(assets.BaseTokens, parameters.Decimals)
		s.creditToAccountFullDecimals(accountKey, incomingTokensFullDecimals)
	}
	for _, nt := range assets.NativeTokens {
		if nt.Amount.Sign() == 0 {
			continue
		}
		if nt.Amount.Sign() < 0 {
			panic(ErrBadAmount)
		}
		balance := s.getNativeTokenAmount(accountKey, nt.ID)
		balance.Add(balance, nt.Amount)
		if balance.Cmp(util.MaxUint256) > 0 {
			panic(ErrOverflow)
		}
		s.setNativeTokenAmount(accountKey, nt.ID, balance)
	}
}

func (s *StateWriter) CreditToAccountFullDecimals(agentID isc.AgentID, amount *big.Int, chainID isc.ChainID) {
	if !bigint.IsPositive(amount) {
		return
	}
	s.creditToAccountFullDecimals(accountKey(agentID, chainID), amount)
	s.creditToAccountFullDecimals(L2TotalsAccount, amount)
	s.touchAccount(agentID, chainID)
}

// creditToAccountFullDecimals adds assets to the internal account map
func (s *StateWriter) creditToAccountFullDecimals(accountKey kv.Key, amount *big.Int) {
	s.setBaseTokensFullDecimals(accountKey, new(big.Int).Add(s.getBaseTokensFullDecimals(accountKey), amount))
}

// DebitFromAccount takes out assets balance the on chain ledger. If not enough it panics
// NOTE: this function does not take NFTs into account
func (s *StateWriter) DebitFromAccount(agentID isc.AgentID, assets *isc.Assets, chainID isc.ChainID) {
	if assets == nil || assets.IsEmpty() {
		return
	}
	if !s.debitFromAccount(accountKey(agentID, chainID), assets) {
		panic(fmt.Errorf("cannot debit (%s) from %s: %w", assets, agentID, ErrNotEnoughFunds))
	}
	if !s.debitFromAccount(L2TotalsAccount, assets) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	s.touchAccount(agentID, chainID)
}

// debitFromAccount debits assets from the internal accounts map
// NOTE: this function does not take NFTs into account
func (s *StateWriter) debitFromAccount(accountKey kv.Key, assets *isc.Assets) bool {
	if assets == nil || assets.IsEmpty() {
		return true
	}

	// first check, then mutate
	mutateBaseTokens := false

	baseTokensToDebit := util.BaseTokensDecimalsToEthereumDecimals(assets.BaseTokens, parameters.Decimals)
	var baseTokensToSet *big.Int
	if assets.BaseTokens > 0 {
		balance := s.getBaseTokensFullDecimals(accountKey)
		if baseTokensToDebit.Cmp(balance) > 0 {
			return false
		}
		mutateBaseTokens = true
		baseTokensToSet = new(big.Int).Sub(balance, baseTokensToDebit)
	}

	nativeTokensMutations := isc.NewEmptyAssets()
	for _, nt := range assets.NativeTokens {
		if nt.Amount.Sign() == 0 {
			continue
		}
		if nt.Amount.Sign() < 0 {
			panic(ErrBadAmount)
		}
		balance := s.getNativeTokenAmount(accountKey, nt.ID)
		balance = balance.Sub(balance, nt.Amount)
		if balance.Sign() < 0 {
			return false
		}
		nativeTokensMutations.AddNativeTokens(nt.ID, balance)
	}

	if mutateBaseTokens {
		s.setBaseTokensFullDecimals(accountKey, baseTokensToSet)
	}
	for _, nt := range nativeTokensMutations.NativeTokens {
		s.setNativeTokenAmount(accountKey, nt.ID, nt.Amount)
	}
	return true
}

// DebitFromAccountFullDecimals removes the amount from the chain ledger. If not enough it panics
func (s *StateWriter) DebitFromAccountFullDecimals(agentID isc.AgentID, amount *big.Int, chainID isc.ChainID) {
	if !bigint.IsPositive(amount) {
		return
	}
	if !s.debitFromAccountFullDecimals(accountKey(agentID, chainID), amount) {
		panic(fmt.Errorf("cannot debit (%s) from %s: %w", amount.String(), agentID, ErrNotEnoughFunds))
	}

	if !s.debitFromAccountFullDecimals(L2TotalsAccount, amount) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	s.touchAccount(agentID, chainID)
}

// debitFromAccountFullDecimals debits the amount from the internal accounts map
func (s *StateWriter) debitFromAccountFullDecimals(accountKey kv.Key, amount *big.Int) bool {
	balance := s.getBaseTokensFullDecimals(accountKey)
	if balance.Cmp(amount) < 0 {
		return false
	}
	s.setBaseTokensFullDecimals(accountKey, new(big.Int).Sub(balance, amount))
	return true
}

// getFungibleTokens returns the fungible tokens owned by an account (base tokens extra decimals will be discarded)
func (s *StateReader) getFungibleTokens(accountKey kv.Key) *isc.Assets {
	ret := isc.NewEmptyAssets()
	bts, _ := s.getBaseTokens(accountKey)
	ret.AddBaseTokens(bts)
	s.nativeTokensMapR(accountKey).Iterate(func(idBytes []byte, val []byte) bool {
		ret.AddNativeTokens(
			isc.MustNativeTokenIDFromBytes(idBytes),
			new(big.Int).SetBytes(val),
		)
		return true
	})
	return ret
}

// GetAccountFungibleTokens returns all fungible tokens belonging to the agentID on the state
func (s *StateReader) GetAccountFungibleTokens(agentID isc.AgentID, chainID isc.ChainID) *isc.Assets {
	return s.getFungibleTokens(accountKey(agentID, chainID))
}

func (s *StateReader) GetTotalL2FungibleTokens() *isc.Assets {
	return s.getFungibleTokens(L2TotalsAccount)
}
