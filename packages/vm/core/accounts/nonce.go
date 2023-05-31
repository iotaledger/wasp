package accounts

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func nonceKey(callerAgentID isc.AgentID) kv.Key {
	return keyNonce + accountKey(callerAgentID)
}

// Nonce returns the "total request count" for an account (its the nonce that is expected in the next request)
func Nonce(state kv.KVStoreReader, callerAgentID isc.AgentID) uint64 {
	data := state.Get(nonceKey(callerAgentID))
	if data == nil {
		return 0
	}
	return codec.MustDecodeUint64(data) + 1
}

func IncrementNonce(state kv.KVStore, callerAgentID isc.AgentID) {
	next := Nonce(state, callerAgentID)
	state.Set(nonceKey(callerAgentID), codec.EncodeUint64(next))
}

func CheckNonce(state kv.KVStoreReader, agentID isc.AgentID, nonce uint64) error {
	expected := Nonce(state, agentID)
	if nonce != expected {
		return fmt.Errorf("Invalid nonce, expected %d, got %d", expected, nonce)
	}
	return nil
}
