package smGPAUtils

import (
	"github.com/iotaledger/wasp/packages/state"
)

type BlockKey state.BlockHash

func NewBlockKey(commitment *state.L1Commitment) BlockKey {
	return BlockKey(commitment.GetBlockHash())
}

func (bkT BlockKey) Equals(other BlockKey) bool {
	return bkT == other
}

func (bkT BlockKey) String() string {
	return string(bkT[:])
}
