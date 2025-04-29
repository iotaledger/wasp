package accounts

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func accountCoinBalancesKey(accountKey kv.Key) string {
	return prefixAccountCoinBalances + string(accountKey)
}

func (s *StateReader) accountCoinBalancesMapR(accountKey kv.Key) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, accountCoinBalancesKey(accountKey))
}

func (s *StateWriter) accountCoinBalancesMap(accountKey kv.Key) *collections.Map {
	return collections.NewMap(s.state, accountCoinBalancesKey(accountKey))
}

func (s *StateReader) getCoinBalance(accountKey kv.Key, coinType coin.Type) coin.Value {
	b := s.accountCoinBalancesMapR(accountKey).GetAt(coinType.Bytes())
	return codec.MustDecode[coin.Value](b, 0)
}

func (s *StateWriter) setCoinBalance(accountKey kv.Key, coinType coin.Type, n coin.Value) {
	if n == 0 {
		s.accountCoinBalancesMap(accountKey).DelAt(coinType.Bytes())
	} else {
		s.accountCoinBalancesMap(accountKey).SetAt(coinType.Bytes(), codec.Encode(n))
	}
}

func (s *StateReader) GetCoinBalance(agentID isc.AgentID, coinID coin.Type) coin.Value {
	return s.getCoinBalance(accountKey(agentID), coinID)
}

func (s *StateReader) GetCoinBalanceTotal(coinID coin.Type) coin.Value {
	return s.getCoinBalance(L2TotalsAccount, coinID)
}

func (s *StateReader) GetCoins(agentID isc.AgentID) isc.CoinBalances {
	ret := isc.NewCoinBalances()
	s.accountCoinBalancesMapR(accountKey(agentID)).Iterate(func(coinType []byte, val []byte) bool {
		ret.Add(
			codec.MustDecode[coin.Type](coinType),
			codec.MustDecode[coin.Value](val),
		)
		return true
	})
	return ret
}
