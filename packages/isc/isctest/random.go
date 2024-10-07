package isctest

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
)

// RandomChainID creates a random chain ID. Used for testing only
func RandomChainID(seed ...[]byte) isc.ChainID {
	var h hashing.HashValue
	if len(seed) > 0 {
		h = hashing.HashData(seed[0])
	} else {
		h = hashing.PseudoRandomHash(nil)
	}
	chainID, err := isc.ChainIDFromBytes(h[:isc.ChainIDLength])
	if err != nil {
		panic(err)
	}
	return chainID
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() isc.AgentID {
	return isc.NewContractAgentID(RandomChainID(), isc.Hn("testName"))
}
