package smGPAUtils

import (
	"errors"

	"github.com/iotaledger/wasp/packages/state"
)

type mockedBlockWAL struct {
	walContents map[state.BlockHash]state.Block
}

var _ BlockWAL = &mockedBlockWAL{}

// For tests only
func NewMockedBlockWAL() BlockWAL {
	return &mockedBlockWAL{walContents: make(map[state.BlockHash]state.Block)}
}

func (mbwT *mockedBlockWAL) Write(block state.Block) error {
	mbwT.walContents[block.Hash()] = block
	return nil
}

func (mbwT *mockedBlockWAL) Contains(blockHash state.BlockHash) bool {
	_, ok := mbwT.walContents[blockHash]
	return ok
}

func (mbwT *mockedBlockWAL) Read(blockHash state.BlockHash) (state.Block, error) {
	block, ok := mbwT.walContents[blockHash]
	if ok {
		return block, nil
	}
	return nil, errors.New("not found")
}
