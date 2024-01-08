package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

func baseTokensKey(accountKey kv.Key) kv.Key {
	return prefixBaseTokens + accountKey
}

func getBaseTokensFullDecimals(state kv.KVStoreReader, accountKey kv.Key) *big.Int {
	return codec.MustDecodeBigIntAbs(state.Get(baseTokensKey(accountKey)), big.NewInt(0))
}

func setBaseTokensFullDecimals(state kv.KVStore, accountKey kv.Key, n *big.Int) {
	state.Set(baseTokensKey(accountKey), codec.EncodeBigIntAbs(n))
}

func getBaseTokens(state kv.KVStoreReader, accountKey kv.Key) uint64 {
	amount := getBaseTokensFullDecimals(state, accountKey)
	// convert from 18 decimals, discard the remainder
	convertedAmount, _ := util.EthereumDecimalsToBaseTokenDecimals(amount, parameters.L1().BaseToken.Decimals)
	return convertedAmount
}

func setBaseTokens(state kv.KVStore, accountKey kv.Key, n uint64) {
	// convert to 18 decimals
	amount := util.MustBaseTokensDecimalsToEthereumDecimalsExact(n, parameters.L1().BaseToken.Decimals)
	state.Set(baseTokensKey(accountKey), codec.EncodeBigIntAbs(amount))
}

func AdjustAccountBaseTokens(state kv.KVStore, account isc.AgentID, adjustment int64, chainID isc.ChainID) {
	switch {
	case adjustment > 0:
		CreditToAccount(state, account, isc.NewAssets(uint64(adjustment), nil), chainID)
	case adjustment < 0:
		DebitFromAccount(state, account, isc.NewAssets(uint64(-adjustment), nil), chainID)
	}
}

func GetBaseTokensBalance(state kv.KVStoreReader, agentID isc.AgentID, chainID isc.ChainID) uint64 {
	return getBaseTokens(state, accountKey(agentID, chainID))
}
