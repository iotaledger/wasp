package accounts

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

func nonceKey(callerAgentID isc.AgentID) kv.Key {
	return keyNonce + accountKey(callerAgentID)
}

// AccountNonce returns the "total request count" for an account (it's the AccountNonce that is expected in the next request)
func (s *StateReader) AccountNonce(callerAgentID isc.AgentID) uint64 {
	if callerAgentID.Kind() == isc.AgentIDKindEthereumAddress {
		panic("to get EVM nonce, call EVM contract")
	}
	data := s.state.Get(nonceKey(callerAgentID))
	if data == nil {
		return 0
	}
	return codec.MustDecode[uint64](data) + 1
}

func (s *StateWriter) IncrementNonce(callerAgentID isc.AgentID) {
	if callerAgentID.Kind() == isc.AgentIDKindEthereumAddress {
		// don't update EVM nonces
		return
	}
	next := s.AccountNonce(callerAgentID)
	s.state.Set(nonceKey(callerAgentID), codec.Encode(next))
}
