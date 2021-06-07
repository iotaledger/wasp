package coreutil

import (
	"sync"

	"go.uber.org/atomic"
)

// GlobalSync and StateBaseline interfaces implements optimistic (non-blocking) access to the
// global state (database) of the chain

// region GlobalSync  ////////////////////////////////

type GlobalSync interface {
	GetSolidIndexBaseline() StateBaseline
	SetSolidIndex(idx uint32) GlobalSync // for use in state manager
	InvalidateSolidIndex() GlobalSync    // only for state manager
	Mutex() *sync.RWMutex
}

type globalSync struct {
	// the lower 32 bits stores index.
	// If the higher 32 bits are not 0, the stateBaseline is invalid until we set 0
	// into higher 32 bits. The one uint64 variable is used for atomicity
	solidIndex atomic.Uint64
	// may be used for exclusive global access (not used atm)
	globalMutex *sync.RWMutex
}

func NewGlobalSync() *globalSync {
	ret := &globalSync{
		globalMutex: &sync.RWMutex{},
	}
	ret.solidIndex.Store(^uint64(0))
	return ret
}

// SetSolidIndex sets solid index to the global sync and makes it valid
// To validate baselines, method Set should be called for each
func (g *globalSync) SetSolidIndex(idx uint32) GlobalSync {
	g.solidIndex.Store(uint64(idx))
	return g
}

// GetSolidIndexBaseline creates an instance of the state baseline linked to the global sync
func (g *globalSync) GetSolidIndexBaseline() StateBaseline {
	return newStateIndexBaseline(&g.solidIndex)
}

// InvalidateSolidIndex invalidates state index and, globally, all baselines
//.All vaselines remain invalid until SetSolidIndex is called globalle
// and Set for each baseline individually
func (g *globalSync) InvalidateSolidIndex() GlobalSync {
	g.solidIndex.Store(^uint64(0))
	return g
}

// Mutex return global mutex which is locked by the state manager during write to DB
// The read lock ar not used atm, it may be removed in the future
func (g *globalSync) Mutex() *sync.RWMutex {
	return g.globalMutex
}

// endregion  ///////////////////////////////////////////////////

// region StateBaseline //////////////////////////////////////////////

type StateBaseline interface {
	Set()
	IsValid() bool
}

type stateBaseline struct {
	globalStateIndex *atomic.Uint64
	baseline         uint32
}

func newStateIndexBaseline(globalStateIndex *atomic.Uint64) *stateBaseline {
	return &stateBaseline{
		globalStateIndex: globalStateIndex,
		baseline:         uint32(globalStateIndex.Load()),
	}
}

func (g *stateBaseline) Set() {
	g.baseline = uint32(g.globalStateIndex.Load())
}

func (g *stateBaseline) IsValid() bool {
	f := g.globalStateIndex.Load()
	if f>>32 != 0 {
		return false
	}
	return g.baseline == uint32(g.globalStateIndex.Load())
}

// endregion /////////////////////////////////////////////////////////////
