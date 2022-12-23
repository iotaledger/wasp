package smGPAUtils

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/state"
)

type emptyBlockWAL struct{}

var _ BlockWAL = &emptyBlockWAL{}

// For tests only
func NewEmptyBlockWAL() BlockWAL                     { return &emptyBlockWAL{} }
func (*emptyBlockWAL) Write(state.Block) error       { return nil }
func (*emptyBlockWAL) Contains(state.BlockHash) bool { return false }
func (*emptyBlockWAL) Read(state.BlockHash) (state.Block, error) {
	return nil, fmt.Errorf("Default WAL contains no elements")
}
