package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func nativeTokensMapKey(accountKey kv.Key) string {
	return prefixNativeTokens + string(accountKey)
}

func (s *StateReader) nativeTokensMapR(accountKey kv.Key) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, nativeTokensMapKey(accountKey))
}

func (s *StateWriter) nativeTokensMap(accountKey kv.Key) *collections.Map {
	return collections.NewMap(s.state, nativeTokensMapKey(accountKey))
}

func (s *StateReader) getNativeTokenAmount(accountKey kv.Key, tokenID isc.CoinType) *big.Int {
	r := new(big.Int)
	b := s.nativeTokensMapR(accountKey).GetAt(tokenID.Bytes())
	if len(b) > 0 {
		r.SetBytes(b)
	}
	return r
}

func (s *StateWriter) setNativeTokenAmount(accountKey kv.Key, tokenID isc.CoinType, n *big.Int) {
	if n.Sign() == 0 {
		s.nativeTokensMap(accountKey).DelAt(tokenID.Bytes())
	} else {
		s.nativeTokensMap(accountKey).SetAt(tokenID.Bytes(), codec.BigIntAbs.Encode(n))
	}
}

func (s *StateReader) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID isc.CoinType, chainID isc.ChainID) *big.Int {
	return s.getNativeTokenAmount(accountKey(agentID, chainID), nativeTokenID)
}

func (s *StateReader) GetNativeTokenBalanceTotal(nativeTokenID isc.CoinType) *big.Int {
	return s.getNativeTokenAmount(L2TotalsAccount, nativeTokenID)
}

func (s *StateReader) GetNativeTokens(agentID isc.AgentID, chainID isc.ChainID) isc.CoinBalances {
	ret := isc.CoinBalances{}
	s.nativeTokensMapR(accountKey(agentID, chainID)).Iterate(func(idBytes []byte, val []byte) bool {
		// TODO: Adapt to map structure
		ret = append(ret, &isc.Coin{
			CoinType: isc.MustNativeTokenIDFromBytes(idBytes),
			Amount:   new(big.Int).SetBytes(val),
		})
		return true
	})
	return ret
}
