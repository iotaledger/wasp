package smGPAUtils

import (
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type BlockKey state.BlockHash

var _ util.Equatable = BlockKey{}

func NewBlockKey(commitment *state.L1Commitment) BlockKey {
	return BlockKey(commitment.BlockHash())
}

func (bkT BlockKey) Equals(e2 util.Equatable) bool {
	bk2, ok := e2.(BlockKey)
	if !ok {
		return false
	}
	return bkT == bk2
}

func (bkT BlockKey) String() string {
	return state.BlockHash(bkT).String()
}
