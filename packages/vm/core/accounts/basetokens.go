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

func (s *StateReader) getBaseTokens(accountKey kv.Key) (tokens uint64, remainder *big.Int) {
	switch s.v {
	case 0:
		return lo.Must(codec.Uint64.Decode(s.state.Get(BaseTokensKey(accountKey)), 0)), big.NewInt(0)
	default:
		amount := s.getBaseTokensFullDecimals(accountKey)
		// convert from 18 decimals, discard the remainder
		return util.EthereumDecimalsToBaseTokenDecimals(amount, parameters.L1().BaseToken.Decimals)
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

func (s *StateWriter) setBaseTokens(accountKey kv.Key, amount uint64) {
	switch s.v {
	case 0:
		s.state.Set(BaseTokensKey(accountKey), codec.Uint64.Encode(uint64(amount)))
	default:
		fullDecimals := util.BaseTokensDecimalsToEthereumDecimals(amount, parameters.L1().BaseToken.Decimals)
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

func (s *StateWriter) AdjustAccountBaseTokens(account isc.AgentID, adjustment int64, chainID isc.ChainID) {
	switch {
	case adjustment > 0:
		s.CreditToAccount(account, isc.NewAssets(uint64(adjustment), nil), chainID)
	case adjustment < 0:
		s.DebitFromAccount(account, isc.NewAssets(uint64(-adjustment), nil), chainID)
	}
}

func (s *StateReader) GetBaseTokensBalance(agentID isc.AgentID, chainID isc.ChainID) (bts uint64, remainder *big.Int) {
	return s.getBaseTokens(accountKey(agentID, chainID))
}

func (s *StateReader) GetBaseTokensBalanceFullDecimals(agentID isc.AgentID, chainID isc.ChainID) *big.Int {
	return s.getBaseTokensFullDecimals(accountKey(agentID, chainID))
}

func (s *StateReader) GetBaseTokensBalanceDiscardExtraDecimals(agentID isc.AgentID, chainID isc.ChainID) uint64 {
	bts, _ := s.getBaseTokens(accountKey(agentID, chainID))
	return bts
}
