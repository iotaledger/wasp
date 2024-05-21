package accounts

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
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

func (s *StateReader) getNativeTokenAmount(accountKey kv.Key, tokenID iotago.NativeTokenID) *big.Int {
	r := new(big.Int)
	b := s.nativeTokensMapR(accountKey).GetAt(tokenID[:])
	if len(b) > 0 {
		r.SetBytes(b)
	}
	return r
}

func (s *StateWriter) setNativeTokenAmount(accountKey kv.Key, tokenID iotago.NativeTokenID, n *big.Int) {
	if n.Sign() == 0 {
		s.nativeTokensMap(accountKey).DelAt(tokenID[:])
	} else {
		s.nativeTokensMap(accountKey).SetAt(tokenID[:], codec.BigIntAbs.Encode(n))
	}
}

func (s *StateReader) GetNativeTokenBalance(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID, chainID isc.ChainID) *big.Int {
	return s.getNativeTokenAmount(accountKey(agentID, chainID), nativeTokenID)
}

func (s *StateReader) GetNativeTokenBalanceTotal(nativeTokenID iotago.NativeTokenID) *big.Int {
	return s.getNativeTokenAmount(L2TotalsAccount, nativeTokenID)
}

func (s *StateReader) GetNativeTokens(agentID isc.AgentID, chainID isc.ChainID) iotago.NativeTokens {
	ret := iotago.NativeTokens{}
	s.nativeTokensMapR(accountKey(agentID, chainID)).Iterate(func(idBytes []byte, val []byte) bool {
		ret = append(ret, &iotago.NativeToken{
			ID:     isc.MustNativeTokenIDFromBytes(idBytes),
			Amount: new(big.Int).SetBytes(val),
		})
		return true
	})
	return ret
}
