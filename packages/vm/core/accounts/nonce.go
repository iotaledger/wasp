package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func nonceKey(callerAgentID isc.AgentID, chainID isc.ChainID) kv.Key {
	return keyNonce + accountKey(callerAgentID, chainID)
}

// Nonce returns the "total request count" for an account (it's the AccountNonce that is expected in the next request)
func AccountNonce(state kv.KVStoreReader, callerAgentID isc.AgentID, chainID isc.ChainID) uint64 {
	if callerAgentID.Kind() == isc.AgentIDKindEthereumAddress {
		panic("to get EVM nonce, call EVM contract")
	}
	data := state.Get(nonceKey(callerAgentID, chainID))
	if data == nil {
		return 0
	}
	return codec.MustDecodeUint64(data) + 1
}

func IncrementNonce(state kv.KVStore, callerAgentID isc.AgentID, chainID isc.ChainID) {
	if callerAgentID.Kind() == isc.AgentIDKindEthereumAddress {
		// don't update EVM nonces
		return
	}
	next := AccountNonce(state, callerAgentID, chainID)
	state.Set(nonceKey(callerAgentID, chainID), codec.EncodeUint64(next))
}
