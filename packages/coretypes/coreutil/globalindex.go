package coreutil

import "go.uber.org/atomic"

// region GlobalSolidIndex  ////////////////////////////////

type globalSolidIndex struct {
	v atomic.Uint64
}

func NewGlobalSolidIndex() *globalSolidIndex {
	ret := &globalSolidIndex{}
	ret.v.Store(^uint64(0))
	return ret
}

func (g *globalSolidIndex) Set(idx uint32) {
	g.v.Store(uint64(idx))
}

func (g *globalSolidIndex) GetBaseline() *SolidStateBaseline {
	return NewStateIndexBaseline(&g.v)
}

func (g *globalSolidIndex) Invalidate() {
	g.v.Store(^uint64(0))
}

// endregion  ///////////////////////////////////////////////////

// region SolidStateBaseline //////////////////////////////////////////////

type SolidStateBaseline struct {
	globalStateIndex *atomic.Uint64
	baseline         uint32
}

func NewStateIndexBaseline(global *atomic.Uint64) *SolidStateBaseline {
	return &SolidStateBaseline{
		globalStateIndex: global,
		baseline:         uint32(global.Load()),
	}
}

func (g *SolidStateBaseline) SetBaseline() {
	g.baseline = uint32(g.globalStateIndex.Load())
}

func (g *SolidStateBaseline) IsValid() bool {
	f := g.globalStateIndex.Load()
	if f>>32 != 0 {
		return false
	}
	return g.baseline == uint32(g.globalStateIndex.Load())
}

// enregion /////////////////////////////////////////////////////////////
