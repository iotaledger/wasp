package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func nonceKey(callerAgentID isc.AgentID) kv.Key {
	return keyMaxAssumedNonce + accountKey(callerAgentID)
}

// GetMaxAssumedNonce is maintained for each caller with the purpose of replay protection of off-ledger requests
func GetMaxAssumedNonce(state kv.KVStoreReader, callerAgentID isc.AgentID) uint64 {
	return codec.MustDecodeUint64(state.Get(nonceKey(callerAgentID)), 0)
}

func SaveMaxAssumedNonce(state kv.KVStore, callerAgentID isc.AgentID, nonce uint64) {
	next := GetMaxAssumedNonce(state, callerAgentID) + 1
	if nonce > next {
		next = nonce
	}
	state.Set(nonceKey(callerAgentID), codec.EncodeUint64(next))
}
