package smGPAUtils

import (
	"time"

	"github.com/iotaledger/wasp/packages/state"
)

type BlockCache interface {
	AddBlock(state.Block) error
	GetBlock(*state.L1Commitment) state.Block
	CleanOlderThan(time.Time)
	/*StoreBlock(state.Block) error
	StoreStateDraft(state.StateDraft) state.Block
	IsBlockStored(*state.L1Commitment) bool*/
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
