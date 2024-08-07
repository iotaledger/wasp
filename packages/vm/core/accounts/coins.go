package accounts

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func coinsMapKey(accountKey kv.Key) string {
	return prefixCoins + string(accountKey)
}

func (s *StateReader) coinsMapR(accountKey kv.Key) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, coinsMapKey(accountKey))
}

func (s *StateWriter) coinsMap(accountKey kv.Key) *collections.Map {
	return collections.NewMap(s.state, coinsMapKey(accountKey))
}

func (s *StateReader) getCoinBalance(accountKey kv.Key, coinType coin.Type) coin.Value {
	b := s.coinsMapR(accountKey).GetAt(coinType.Bytes())
	return codec.CoinValue.MustDecode(b, 0)
}

func (s *StateWriter) setCoinBalance(accountKey kv.Key, coinType coin.Type, n coin.Value) {
	if n == 0 {
		s.coinsMap(accountKey).DelAt(coinType.Bytes())
	} else {
		s.coinsMap(accountKey).SetAt(coinType.Bytes(), codec.CoinValue.Encode(n))
	}
}

func (s *StateReader) GetCoinBalance(agentID isc.AgentID, coinID coin.Type, chainID isc.ChainID) coin.Value {
	return s.getCoinBalance(accountKey(agentID, chainID), coinID)
}

func (s *StateReader) GetCoinBalanceTotal(coinID coin.Type) coin.Value {
	return s.getCoinBalance(L2TotalsAccount, coinID)
}

func (s *StateReader) GetCoins(agentID isc.AgentID, chainID isc.ChainID) isc.CoinBalances {
	ret := isc.CoinBalances{}
	s.coinsMapR(accountKey(agentID, chainID)).Iterate(func(coinType []byte, val []byte) bool {
		ret.Add(
			codec.CoinType.MustDecode(coinType),
			codec.CoinValue.MustDecode(val),
		)
		return true
	})
	return ret
}
