package accounts

import (
	"math/big"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

func (s *StateReader) getBaseTokens(accountKey kv.Key) (tokens *big.Int, remainder *big.Int) {
	switch s.v {
	case 0:
		return lo.Must(codec.BigIntAbs.Decode(s.state.Get(BaseTokensKey(accountKey)), big.NewInt(0))), big.NewInt(0)
	default:
		amount := s.getBaseTokensFullDecimals(accountKey)
		// convert from 18 decimals, discard the remainder
		return util.EthereumDecimalsToBaseTokenDecimals(amount, parameters.Decimals)
	}
}

const v0BaseTokenDecimals = 6 // all v0 state was saved with 6 decimals

func (s *StateReader) getBaseTokensFullDecimals(accountKey kv.Key) *big.Int {
	switch s.v {
	case 0:
		baseTokens, _ := s.getBaseTokens(accountKey)
		return util.BaseTokensDecimalsToEthereumDecimals(baseTokens, v0BaseTokenDecimals)
	default:
		return lo.Must(codec.BigIntAbs.Decode(s.state.Get(BaseTokensKey(accountKey)), big.NewInt(0)))
	}
}

func (s *StateWriter) setBaseTokens(accountKey kv.Key, amount *big.Int) {
	switch s.v {
	case 0:
		s.state.Set(BaseTokensKey(accountKey), codec.BigIntAbs.Encode(amount))
	default:
		fullDecimals := util.BaseTokensDecimalsToEthereumDecimals(amount, parameters.Decimals)
		s.setBaseTokensFullDecimals(accountKey, fullDecimals)
	}
}

func (s *StateWriter) setBaseTokensFullDecimals(accountKey kv.Key, amount *big.Int) {
	switch s.v {
	case 0:
		baseTokens := util.MustEthereumDecimalsToBaseTokenDecimalsExact(amount, v0BaseTokenDecimals)
		s.setBaseTokens(accountKey, baseTokens)
	default:
		s.state.Set(BaseTokensKey(accountKey), codec.BigIntAbs.Encode(amount))
	}
}

func BaseTokensKey(accountKey kv.Key) kv.Key {
	return prefixBaseTokens + accountKey
}

func (s *StateWriter) AdjustAccountBaseTokens(account isc.AgentID, adjustment *big.Int, chainID isc.ChainID) {
	switch adjustment.Cmp(big.NewInt(0)) {
	case 1:
		// Greater than 0
		s.CreditToAccount(account, isc.NewAssets(adjustment), chainID)
	case -1:
		// Smaller than 0
		s.DebitFromAccount(account, isc.NewAssets(new(big.Int).Neg(adjustment)), chainID)
	}
}

func (s *StateReader) GetBaseTokensBalance(agentID isc.AgentID, chainID isc.ChainID) (bts *big.Int, remainder *big.Int) {
	return s.getBaseTokens(accountKey(agentID, chainID))
}

func (s *StateReader) GetBaseTokensBalanceFullDecimals(agentID isc.AgentID, chainID isc.ChainID) *big.Int {
	return s.getBaseTokensFullDecimals(accountKey(agentID, chainID))
}

func (s *StateReader) GetBaseTokensBalanceDiscardExtraDecimals(agentID isc.AgentID, chainID isc.ChainID) *big.Int {
	bts, _ := s.getBaseTokens(accountKey(agentID, chainID))
	return bts
}
