package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
)

func (s *StateReader) getBaseTokens(accountKey kv.Key) (baseTokens coin.Value, remainderWei *big.Int) {
	if s.v < allmigrations.SchemaVersionMigratedRebased {
		panic("unsupported schema version")
	}
	baseTokens = s.getCoinBalance(accountKey, coin.BaseTokenType)
	remainderWei = s.getWeiRemainder(accountKey)
	return
}

func (s *StateReader) getBaseTokensFullDecimals(accountKey kv.Key) *big.Int {
	baseTokens, remainderWei := s.getBaseTokens(accountKey)
	wei := util.BaseTokensDecimalsToEthereumDecimals(baseTokens, parameters.BaseTokenDecimals)
	wei.Add(wei, remainderWei)
	return wei
}

func (s *StateWriter) setBaseTokens(accountKey kv.Key, baseTokens coin.Value, remainderWei *big.Int) {
	if s.v < allmigrations.SchemaVersionMigratedRebased {
		panic("unsupported schema version")
	}
	s.setCoinBalance(accountKey, coin.BaseTokenType, baseTokens)
	s.setWeiRemainder(accountKey, remainderWei)
}

func (s *StateWriter) setBaseTokensFullDecimals(accountKey kv.Key, wei *big.Int) {
	baseTokens, remainderWei := util.EthereumDecimalsToBaseTokenDecimals(wei, parameters.BaseTokenDecimals)
	s.setBaseTokens(accountKey, baseTokens, remainderWei)
}

func (s *StateWriter) AdjustAccountBaseTokens(account isc.AgentID, adjustment coin.Value) {
	b := isc.NewCoinBalances().AddBaseTokens(adjustment)
	if adjustment > 0 {
		s.CreditToAccount(account, b)
	}
}

func (s *StateReader) GetBaseTokensBalance(agentID isc.AgentID) (bts coin.Value, remainder *big.Int) {
	return s.getBaseTokens(accountKey(agentID))
}

func (s *StateReader) GetBaseTokensBalanceFullDecimals(agentID isc.AgentID) *big.Int {
	return s.getBaseTokensFullDecimals(accountKey(agentID))
}

func (s *StateReader) GetBaseTokensBalanceDiscardExtraDecimals(agentID isc.AgentID) coin.Value {
	bts, _ := s.getBaseTokens(accountKey(agentID))
	return bts
}

func accountWeiRemainderKey(accountKey kv.Key) kv.Key {
	return prefixAccountWeiRemainder + accountKey
}

func (s *StateReader) getWeiRemainder(accountKey kv.Key) *big.Int {
	b := s.state.Get(accountWeiRemainderKey(accountKey))
	if b == nil {
		return new(big.Int)
	}
	return codec.MustDecode[*big.Int](b)
}

func (s *StateWriter) setWeiRemainder(accountKey kv.Key, v *big.Int) {
	if v.Sign() == 0 {
		s.state.Del(accountWeiRemainderKey(accountKey))
	} else {
		s.state.Set(accountWeiRemainderKey(accountKey), codec.Encode(v))
	}
}
