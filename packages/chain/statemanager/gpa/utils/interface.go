package utils

import (
	"time"

	"github.com/iotaledger/wasp/packages/state"
)

type BlockCache interface {
	AddBlock(state.Block)
	GetBlock(*state.L1Commitment) state.Block
	CleanOlderThan(time.Time)
	Size() int
}

type BlockWAL interface {
	Write(state.Block) error
	Contains(state.BlockHash) bool
	Read(state.BlockHash) (state.Block, error)
	ReadAllByStateIndex(cb func(stateIndex uint32, block state.Block) bool) error
}
