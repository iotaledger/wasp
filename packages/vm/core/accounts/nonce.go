package accounts

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func nonceKey(callerAgentID isc.AgentID) kv.Key {
	return keyMaxAssumedNonce + accountKey(callerAgentID)
}

// Nonce is maintained for each caller with the purpose of replay protection of off-ledger requests
func Nonce(state kv.KVStoreReader, callerAgentID isc.AgentID) *uint64 {
	data := state.Get(nonceKey(callerAgentID))
	if data == nil {
		return nil
	}
	ret := codec.MustDecodeUint64(data, 0)
	return &ret
}

func IncrementNonce(state kv.KVStore, callerAgentID isc.AgentID) {
	accountNonce := Nonce(state, callerAgentID)
	var next uint64
	if accountNonce == nil {
		next = 0
	} else {
		next = *accountNonce + 1
		if *accountNonce > next {
			panic("nonce overflow") // prevent overflow
		}
	}
	state.Set(nonceKey(callerAgentID), codec.EncodeUint64(next))
}

func CheckNonce(state kv.KVStoreReader, agentID isc.AgentID, nonce uint64) error {
	accountNonce := Nonce(state, agentID)
	if accountNonce == nil {
		if nonce == 0 {
			return nil // new account, first request
		} else {
			return fmt.Errorf("Invalid nonce, expected %d, got %d", 0, nonce)
		}
	}
	expected := *accountNonce + 1
	if expected < *accountNonce {
		// this will never occur, it would require something like 5,845,420,460 Tx a second (by the same address) for 100 years
		panic("maximum uint64 nonce reached, this account cannot send more requests")
	}
	if nonce != expected {
		return fmt.Errorf("Invalid nonce, expected %d, got %d", expected, nonce)
	}
	return nil
}
