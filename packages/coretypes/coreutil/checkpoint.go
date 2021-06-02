package coreutil

import "go.uber.org/atomic"

// endregion ////////////////////////////////////////////

// region GlobalReadCheckpoint  ///////////////////////////////////////////
type GlobalReadCheckpoint struct {
	stateIndex atomic.Uint32
	baseline   uint32
}

func NewGlobalReadCheckpoint() *GlobalReadCheckpoint {
	ret := &GlobalReadCheckpoint{}
	ret.stateIndex.Store(^uint32(0))
	return ret
}

func (g *GlobalReadCheckpoint) Start() {
	g.baseline = g.stateIndex.Load()
}

func (g *GlobalReadCheckpoint) IsValid() bool {
	return g.baseline == g.stateIndex.Load()
}

func (g *GlobalReadCheckpoint) SetGlobalStateIndex(stateIndex uint32) {
	g.stateIndex.Store(stateIndex)
}

// endregion //////////////////////////////////////////////
