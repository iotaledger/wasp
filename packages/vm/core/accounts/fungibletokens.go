package accounts

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/bigint"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
)

// CreditToAccount brings new funds to the on chain ledger
func (s *StateWriter) CreditToAccount(agentID isc.AgentID, coins isc.CoinBalances, chainID isc.ChainID) {
	if len(coins) == 0 {
		return
	}
	s.creditToAccount(accountKey(agentID, chainID), coins)
	s.creditToAccount(L2TotalsAccount, coins)
	s.touchAccount(agentID, chainID)
}

// creditToAccount adds coins to the internal account map
func (s *StateWriter) creditToAccount(accountKey kv.Key, coins isc.CoinBalances) {
	if len(coins) == 0 {
		return
	}

	for coinType, amount := range coins {
		if amount.Sign() == 0 {
			continue
		}
		if amount.Sign() < 0 {
			panic(ErrBadAmount)
		}
		balance := s.getCoinBalance(accountKey, coinType)
		balance.Add(balance, amount)
		if balance.Cmp(util.MaxUint64) > 0 {
			panic(ErrOverflow)
		}
		s.setCoinBalance(accountKey, coinType, balance)
	}
}

func (s *StateWriter) CreditToAccountFullDecimals(agentID isc.AgentID, wei *big.Int, chainID isc.ChainID) {
	if !bigint.IsPositive(wei) {
		return
	}
	s.creditToAccountFullDecimals(accountKey(agentID, chainID), wei)
	s.creditToAccountFullDecimals(L2TotalsAccount, wei)
	s.touchAccount(agentID, chainID)
}

// creditToAccountFullDecimals adds coins to the internal account map
func (s *StateWriter) creditToAccountFullDecimals(accountKey kv.Key, wei *big.Int) {
	s.setBaseTokensFullDecimals(accountKey, new(big.Int).Add(s.getBaseTokensFullDecimals(accountKey), wei))
}

// DebitFromAccount takes out coins balance the on chain ledger. If not enough it panics
func (s *StateWriter) DebitFromAccount(agentID isc.AgentID, coins isc.CoinBalances, chainID isc.ChainID) {
	if len(coins) == 0 {
		return
	}
	if !s.debitFromAccount(accountKey(agentID, chainID), coins) {
		panic(fmt.Errorf("cannot debit (%s) from %s: %w", coins, agentID, ErrNotEnoughFunds))
	}
	if !s.debitFromAccount(L2TotalsAccount, coins) {
		panic("debitFromAccount: inconsistent ledger state")
	}
	s.touchAccount(agentID, chainID)
}

// debitFromAccount debits coins from the internal accounts map
// NOTE: this function does not take NFTs into account
func (s *StateWriter) debitFromAccount(accountKey kv.Key, coins isc.CoinBalances) bool {
	if len(coins) == 0 {
		return true
	}

	// first check, then mutate
	coinMutations := isc.NewCoinBalances()
	for coinType, amount := range coins {
		if amount.Sign() == 0 {
			continue
		}
		if amount.Sign() < 0 {
			panic(ErrBadAmount)
		}
		balance := s.getCoinBalance(accountKey, coinType)
		balance = balance.Sub(balance, amount)
		if balance.Sign() < 0 {
			return false
		}
		coinMutations.Add(coinType, balance)
	}

	for coinType, amount := range coinMutations {
		s.setCoinBalance(accountKey, coinType, amount)
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
func (s *StateReader) getFungibleTokens(accountKey kv.Key) isc.CoinBalances {
	ret := isc.NewCoinBalances()
	s.coinsMapR(accountKey).Iterate(func(coinType []byte, val []byte) bool {
		ret.Add(
			codec.CoinType.MustDecode(coinType),
			codec.BigIntAbs.MustDecode(val),
		)
		return true
	})
	return ret
}

// GetAccountFungibleTokens returns all fungible tokens belonging to the agentID on the state
func (s *StateReader) GetAccountFungibleTokens(agentID isc.AgentID, chainID isc.ChainID) isc.CoinBalances {
	return s.getFungibleTokens(accountKey(agentID, chainID))
}

func (s *StateReader) GetTotalL2FungibleTokens() isc.CoinBalances {
	return s.getFungibleTokens(L2TotalsAccount)
}
