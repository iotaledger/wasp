package smGPAUtils

import (
	"github.com/iotaledger/wasp/packages/state"
)

// Type for block identifier to be used when putting blocks in maps.
type BlockKey state.BlockHash

func NewBlockKey(commitment *state.L1Commitment) BlockKey {
	return BlockKey(commitment.BlockHash())
}

func (bkT BlockKey) Equals(other BlockKey) bool {
	return bkT == other
}

func (bkT BlockKey) AsBlockHash() state.BlockHash {
	return state.BlockHash(bkT)
}

func (bkT BlockKey) String() string {
	return state.BlockHash(bkT).String()
}
