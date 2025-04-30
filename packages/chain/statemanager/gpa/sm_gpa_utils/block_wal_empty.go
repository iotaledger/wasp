package sm_gpa_utils

import (
	"errors"

	"github.com/iotaledger/wasp/packages/state"
)

// May be used in tests or in production as a noop WAL.
type emptyBlockWAL struct{}

var (
	_ BlockWAL     = &emptyBlockWAL{}
	_ TestBlockWAL = &emptyBlockWAL{}
)

func NewEmptyBlockWAL() BlockWAL                     { return NewEmptyTestBlockWAL() }
func NewEmptyTestBlockWAL() TestBlockWAL             { return &emptyBlockWAL{} }
func (*emptyBlockWAL) Write(state.Block) error       { return nil }
func (*emptyBlockWAL) Contains(state.BlockHash) bool { return false }
func (*emptyBlockWAL) Read(state.BlockHash) (state.Block, error) {
	return nil, errors.New("default WAL contains no elements")
}
func (*emptyBlockWAL) Delete(state.BlockHash) bool { return false }

func (*emptyBlockWAL) ReadAllByStateIndex(cb func(stateIndex uint32, block state.Block) bool) error {
	return nil // Not needed in this mock.
}
