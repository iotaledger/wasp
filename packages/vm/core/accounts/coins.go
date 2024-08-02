package accounts

import (
	"math/big"

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

func (s *StateReader) getCoinBalance(accountKey kv.Key, coinType isc.CoinType) *big.Int {
	r := new(big.Int)
	b := s.coinsMapR(accountKey).GetAt(coinType.Bytes())
	if len(b) > 0 {
		r.SetBytes(b)
	}
	return r
}

func (s *StateWriter) setCoinBalance(accountKey kv.Key, coinType isc.CoinType, n *big.Int) {
	if n.Sign() == 0 {
		s.coinsMap(accountKey).DelAt(coinType.Bytes())
	} else {
		s.coinsMap(accountKey).SetAt(coinType.Bytes(), codec.BigIntAbs.Encode(n))
	}
}

func (s *StateReader) GetCoinBalance(agentID isc.AgentID, coinID isc.CoinType, chainID isc.ChainID) *big.Int {
	return s.getCoinBalance(accountKey(agentID, chainID), coinID)
}

func (s *StateReader) GetCoinBalanceTotal(coinID isc.CoinType) *big.Int {
	return s.getCoinBalance(L2TotalsAccount, coinID)
}

func (s *StateReader) GetCoins(agentID isc.AgentID, chainID isc.ChainID) isc.CoinBalances {
	ret := isc.CoinBalances{}
	s.coinsMapR(accountKey(agentID, chainID)).Iterate(func(coinType []byte, val []byte) bool {
		ret.Add(
			codec.CoinType.MustDecode(coinType),
			codec.BigIntAbs.MustDecode(val),
		)
		return true
	})
	return ret
}
