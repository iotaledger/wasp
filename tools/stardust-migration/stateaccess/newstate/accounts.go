package newstate

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// This file contains pieces of business logic, which are missing in the codebase.

func SetAccountNonce(state kv.KVStore, agentID isc.AgentID, chainID isc.ChainID, nonce uint64) {
	if agentID.Kind() == isc.AgentIDKindEthereumAddress {
		// don't update EVM nonces
		return
	}
	state.Set(accounts.NonceKey(agentID, chainID), codec.Encode(nonce))
}
