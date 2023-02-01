package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func baseTokensKey(accountKey kv.Key) kv.Key {
	return prefixBaseTokens + accountKey
}

func getBaseTokens(state kv.KVStoreReader, accountKey kv.Key) uint64 {
	return codec.MustDecodeUint64(state.MustGet(baseTokensKey(accountKey)), 0)
}

func setBaseTokens(state kv.KVStore, accountKey kv.Key, n uint64) {
	state.Set(baseTokensKey(accountKey), codec.EncodeUint64(n))
}

func AdjustAccountBaseTokens(state kv.KVStore, account isc.AgentID, adjustment int64) {
	switch {
	case adjustment > 0:
		CreditToAccount(state, account, isc.NewAssets(uint64(adjustment), nil))
	case adjustment < 0:
		DebitFromAccount(state, account, isc.NewAssets(uint64(-adjustment), nil))
	}
}

func GetBaseTokensBalance(state kv.KVStoreReader, agentID isc.AgentID) uint64 {
	return getBaseTokens(state, accountKey(agentID))
}
