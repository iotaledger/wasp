package coreutil

import (
	"sync"

	"go.uber.org/atomic"
)

// ChainStateSync and StateBaseline interfaces implements optimistic (non-blocking) access to the
// global state (database) of the chain

// region ChainStateSync  ////////////////////////////////

type ChainStateSync interface {
	GetSolidIndexBaseline() StateBaseline
	SetSolidIndex(idx uint32) ChainStateSync // for use in state manager
	InvalidateSolidIndex() ChainStateSync    // only for state manager
	Mutex() *sync.RWMutex
}

type ChainStateSyncImpl struct {
	solidIndex  atomic.Uint32
	globalMutex *sync.RWMutex
}

// we assume last state index 2^32 will never be reached :)

func NewChainStateSync() *ChainStateSyncImpl {
	ret := &ChainStateSyncImpl{
		globalMutex: &sync.RWMutex{},
	}
	ret.solidIndex.Store(^uint32(0))
	return ret
}

// SetSolidIndex sets solid index to the global sync and makes it valid
// To validate baselines, method Set should be called for each
func (g *ChainStateSyncImpl) SetSolidIndex(idx uint32) ChainStateSync {
	if idx == ^uint32(0) {
		panic("SetSolidIndex: wrong state index")
	}
	g.solidIndex.Store(idx)
	return g
}

// GetSolidIndexBaseline creates an instance of the state baseline linked to the global sync
func (g *ChainStateSyncImpl) GetSolidIndexBaseline() StateBaseline {
	return newStateIndexBaseline(&g.solidIndex)
}

// InvalidateSolidIndex invalidates state index and, globally, all baselines
//.All vaselines remain invalid until SetSolidIndex is called globally
// and Set for each baseline individually
func (g *ChainStateSyncImpl) InvalidateSolidIndex() ChainStateSync {
	g.solidIndex.Store(^uint32(0))
	return g
}

// Mutex return global mutex which is locked by the state manager during write to DB
// The read lock ar not used atm, it may be removed in the future
func (g *ChainStateSyncImpl) Mutex() *sync.RWMutex {
	return g.globalMutex
}

// endregion  ///////////////////////////////////////////////////

// region StateBaseline //////////////////////////////////////////////

type StateBaseline interface {
	Set()
	IsValid() bool
}

type stateBaseline struct {
	solidStateIndex *atomic.Uint32
	baseline        uint32
}

func newStateIndexBaseline(globalStateIndex *atomic.Uint32) *stateBaseline {
	return &stateBaseline{
		solidStateIndex: globalStateIndex,
		baseline:        globalStateIndex.Load(),
	}
}

func (g *stateBaseline) Set() {
	g.baseline = g.solidStateIndex.Load()
}

func (g *stateBaseline) IsValid() bool {
	return g.baseline != ^uint32(0) && g.baseline == g.solidStateIndex.Load()
}

// endregion /////////////////////////////////////////////////////////////
