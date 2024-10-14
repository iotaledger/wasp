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
	if s.v < allmigrations.SchemaVersionIotaRebased {
		panic("unsupported schema version")
	}
	baseTokens = s.getCoinBalance(accountKey, coin.BaseTokenType)
	remainderWei = s.getWeiRemainder(accountKey)
	return
}

func (s *StateReader) getBaseTokensFullDecimals(accountKey kv.Key) *big.Int {
	baseTokens, remainderWei := s.getBaseTokens(accountKey)
	wei := util.BaseTokensDecimalsToEthereumDecimals(baseTokens, parameters.Decimals)
	wei.Add(wei, remainderWei)
	return wei
}

func (s *StateWriter) setBaseTokens(accountKey kv.Key, baseTokens coin.Value, remainderWei *big.Int) {
	if s.v < allmigrations.SchemaVersionIotaRebased {
		panic("unsupported schema version")
	}
	s.setCoinBalance(accountKey, coin.BaseTokenType, baseTokens)
	s.setWeiRemainder(accountKey, remainderWei)
}

func (s *StateWriter) setBaseTokensFullDecimals(accountKey kv.Key, wei *big.Int) {
	baseTokens, remainderWei := util.EthereumDecimalsToBaseTokenDecimals(wei, parameters.Decimals)
	s.setBaseTokens(accountKey, baseTokens, remainderWei)
}

func (s *StateWriter) AdjustAccountBaseTokens(account isc.AgentID, adjustment coin.Value, chainID isc.ChainID) {
	b := isc.NewCoinBalances().AddBaseTokens(adjustment)
	switch {
	case adjustment > 0:
		s.CreditToAccount(account, b, chainID)
	case adjustment < 0:
		s.DebitFromAccount(account, b, chainID)
	}
}

func (s *StateReader) GetBaseTokensBalance(agentID isc.AgentID, chainID isc.ChainID) (bts coin.Value, remainder *big.Int) {
	return s.getBaseTokens(accountKey(agentID, chainID))
}

func (s *StateReader) GetBaseTokensBalanceFullDecimals(agentID isc.AgentID, chainID isc.ChainID) *big.Int {
	return s.getBaseTokensFullDecimals(accountKey(agentID, chainID))
}

func (s *StateReader) GetBaseTokensBalanceDiscardExtraDecimals(agentID isc.AgentID, chainID isc.ChainID) coin.Value {
	bts, _ := s.getBaseTokens(accountKey(agentID, chainID))
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
	s.state.Set(accountWeiRemainderKey(accountKey), codec.Encode[*big.Int](v))
}
