package accounts

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/bigint"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// CreditToAccount brings new funds to the on chain ledger
func (s *StateWriter) CreditToAccount(agentID isc.AgentID, coins isc.CoinBalances) {
	if len(coins) == 0 {
		return
	}
	s.creditToAccount(AccountKey(agentID), coins)
	s.creditToAccount(L2TotalsAccount, coins)
	s.touchAccount(agentID)
}

// creditToAccount adds coins to the internal account map
func (s *StateWriter) creditToAccount(accountKey kv.Key, coins isc.CoinBalances) {
	if len(coins) == 0 {
		return
	}

	for coinType, amount := range coins.Iterate() {
		if amount == 0 {
			continue
		}
		balance := s.getCoinBalance(accountKey, coinType) + amount
		s.setCoinBalance(accountKey, coinType, balance)
	}
}

func (s *StateWriter) CreditToAccountFullDecimals(agentID isc.AgentID, wei *big.Int) {
	if !bigint.IsPositive(wei) {
		return
	}
	s.creditToAccountFullDecimals(AccountKey(agentID), wei)
	s.creditToAccountFullDecimals(L2TotalsAccount, wei)
	s.touchAccount(agentID)
}

// creditToAccountFullDecimals adds coins to the internal account map
func (s *StateWriter) creditToAccountFullDecimals(accountKey kv.Key, wei *big.Int) {
	s.setBaseTokensFullDecimals(accountKey, new(big.Int).Add(s.getBaseTokensFullDecimals(accountKey), wei))
}

// DebitFromAccount takes out coins balance the on chain ledger. If not enough it panics
func (s *StateWriter) DebitFromAccount(agentID isc.AgentID, coins isc.CoinBalances) {
	if len(coins) == 0 {
		return
	}
	if !s.debitFromAccount(AccountKey(agentID), coins) {
		panic(fmt.Errorf("cannot debit (%s) from %s: %w", coins, agentID, ErrNotEnoughFunds))
	}
	if !s.debitFromAccount(L2TotalsAccount, coins) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	s.touchAccount(agentID)
}

// debitFromAccount debits coins from the internal accounts map
func (s *StateWriter) debitFromAccount(accountKey kv.Key, coins isc.CoinBalances) bool {
	if len(coins) == 0 {
		return true
	}

	// first check, then mutate
	coinMutations := isc.NewCoinBalances()
	for coinType, amount := range coins.Iterate() {
		if amount == 0 {
			continue
		}
		balance := s.getCoinBalance(accountKey, coinType)
		if balance < amount {
			return false
		}
		coinMutations[coinType] = balance - amount
	}

	for coinType, amount := range coinMutations.Iterate() {
		s.setCoinBalance(accountKey, coinType, amount)
	}
	return true
}

// DebitFromAccountFullDecimals removes the amount from the chain ledger. If not enough it panics
func (s *StateWriter) DebitFromAccountFullDecimals(agentID isc.AgentID, amount *big.Int) {
	if !bigint.IsPositive(amount) {
		return
	}
	if !s.debitFromAccountFullDecimals(AccountKey(agentID), amount) {
		panic(fmt.Errorf("cannot debit (%s) from %s: %w", amount.String(), agentID, ErrNotEnoughFunds))
	}

	if !s.debitFromAccountFullDecimals(L2TotalsAccount, amount) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	s.touchAccount(agentID)
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
func (s *StateReader) getFungibleTokens(accountKey kv.Key) isc.CoinBalances {
	ret := isc.NewCoinBalances()
	s.accountCoinBalancesMapR(accountKey).Iterate(func(coinType []byte, val []byte) bool {
		ret.Add(
			codec.MustDecode[coin.Type](coinType),
			codec.MustDecode[coin.Value](val),
		)
		return true
	})
	return ret
}

// GetAccountFungibleTokens returns all fungible tokens belonging to the agentID on the state
func (s *StateReader) GetAccountFungibleTokens(agentID isc.AgentID) isc.CoinBalances {
	return s.getFungibleTokens(AccountKey(agentID))
}

func (s *StateReader) GetTotalL2FungibleTokens() isc.CoinBalances {
	return s.getFungibleTokens(L2TotalsAccount)
}
