package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/sui-go/sui"
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

func (s *StateReader) getNativeTokenAmount(accountKey kv.Key, tokenID sui.ObjectID) *big.Int {
	r := new(big.Int)
	b := s.nativeTokensMapR(accountKey).GetAt(tokenID[:])
	if len(b) > 0 {
		r.SetBytes(b)
	}
	return r
}

func (s *StateWriter) setNativeTokenAmount(accountKey kv.Key, tokenID sui.ObjectID, n *big.Int) {
	if n.Sign() == 0 {
		s.nativeTokensMap(accountKey).DelAt(tokenID[:])
	} else {
		s.nativeTokensMap(accountKey).SetAt(tokenID[:], codec.BigIntAbs.Encode(n))
	}
}

func (s *StateReader) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID sui.ObjectID, chainID isc.ChainID) *big.Int {
	return s.getNativeTokenAmount(accountKey(agentID, chainID), nativeTokenID)
}

func (s *StateReader) GetNativeTokenBalanceTotal(nativeTokenID sui.ObjectID) *big.Int {
	return s.getNativeTokenAmount(L2TotalsAccount, nativeTokenID)
}

func (s *StateReader) GetNativeTokens(agentID isc.AgentID, chainID isc.ChainID) isc.NativeTokens {
	ret := isc.NativeTokens{}
	s.nativeTokensMapR(accountKey(agentID, chainID)).Iterate(func(idBytes []byte, val []byte) bool {
		ret = append(ret, &isc.NativeToken{
			ID:     isc.MustNativeTokenIDFromBytes(idBytes),
			Amount: new(big.Int).SetBytes(val),
		})
		return true
	})
	return ret
}
