package coreutil

import "go.uber.org/atomic"

type StateIndexBaseline struct {
	globalStateIndex *atomic.Uint32
	baseline         uint32
}

func NewStateIndexBaseline(global *atomic.Uint32) *StateIndexBaseline {
	return &StateIndexBaseline{
		globalStateIndex: global,
		baseline:         global.Load(),
	}
}

func (g *StateIndexBaseline) SetBaseline() {
	g.baseline = g.globalStateIndex.Load()
}

func (g *StateIndexBaseline) IsValid() bool {
	return g.baseline == g.globalStateIndex.Load()
}
