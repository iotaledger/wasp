package utils

import (
	"errors"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/wasp/packages/state"
)

// May be used in tests or (very unlikely) in production as a memory only WAL.
type mockedBlockWAL struct {
	walContents *shrinkingmap.ShrinkingMap[state.BlockHash, state.Block]
}

var (
	_ BlockWAL     = &mockedBlockWAL{}
	_ TestBlockWAL = &mockedBlockWAL{}
)

func NewMockedTestBlockWAL() TestBlockWAL {
	return &mockedBlockWAL{walContents: shrinkingmap.New[state.BlockHash, state.Block]()}
}

func (mbwT *mockedBlockWAL) Write(block state.Block) error {
	mbwT.walContents.Set(block.Hash(), block)
	return nil
}

func (mbwT *mockedBlockWAL) Contains(blockHash state.BlockHash) bool {
	return mbwT.walContents.Has(blockHash)
}

func (mbwT *mockedBlockWAL) Read(blockHash state.BlockHash) (state.Block, error) {
	block, exists := mbwT.walContents.Get(blockHash)
	if !exists {
		return nil, errors.New("not found")
	}
	return block, nil
}

func (mbwT *mockedBlockWAL) Delete(blockHash state.BlockHash) bool {
	contains := mbwT.Contains(blockHash)
	if contains {
		mbwT.walContents.Delete(blockHash)
	}
	return contains
}

func (mbwT *mockedBlockWAL) ReadAllByStateIndex(cb func(stateIndex uint32, block state.Block) bool) error {
	return nil // Not needed in this mock.
}
