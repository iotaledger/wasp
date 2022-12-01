package smGPAUtils

import (
	"time"

	"github.com/iotaledger/wasp/packages/state"
)

type BlockCache interface {
	AddBlock(state.Block) error
	GetBlock(uint32, state.BlockHash) state.Block // TODO: first parameter is temporary. Remove it after DB is refactored.
	CleanOlderThan(time.Time)
}

type BlockWAL interface {
	Write(state.Block) error
	Contains(state.BlockHash) bool
	Read(state.BlockHash) (state.Block, error)
}

type TimeProvider interface {
	SetNow(time.Time)
	GetNow() time.Time
}
