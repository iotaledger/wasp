package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (s *StateReader) nonceKey(callerAgentID isc.AgentID) kv.Key {
	return keyNonce + s.accountKey(callerAgentID)
}

// AccountNonce returns the "total request count" for an account (it's the AccountNonce that is expected in the next request)
func (s *StateReader) AccountNonce(callerAgentID isc.AgentID) uint64 {
	if callerAgentID.Kind() == isc.AgentIDKindEthereumAddress {
		panic("to get EVM nonce, call EVM contract")
	}
	data := s.state.Get(s.nonceKey(callerAgentID))
	if data == nil {
		return 0
	}
	return lo.Must(codec.Uint64.Decode(data)) + 1
}

func (s *StateWriter) IncrementNonce(callerAgentID isc.AgentID) {
	if callerAgentID.Kind() == isc.AgentIDKindEthereumAddress {
		// don't update EVM nonces
		return
	}
	next := s.AccountNonce(callerAgentID)
	s.state.Set(s.nonceKey(callerAgentID), codec.Uint64.Encode(next))
}
