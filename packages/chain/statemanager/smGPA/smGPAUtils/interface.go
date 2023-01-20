package smGPAUtils

import (
	"time"

	"github.com/iotaledger/wasp/packages/state"
)

type BlockCache interface {
	AddBlock(state.Block)
	GetBlock(*state.L1Commitment) state.Block
	CleanOlderThan(time.Time)
}

type BlockWAL interface {
	Write(state.Block) error
	Contains(state.BlockHash) bool
	Read(state.BlockHash) (state.Block, error)
	Contents() []state.BlockHash
}

type TimeProvider interface {
	SetNow(time.Time)
	GetNow() time.Time
	After(time.Duration) <-chan time.Time
}
