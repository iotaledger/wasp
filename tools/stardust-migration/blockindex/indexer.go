package blockindex

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"time"

	"github.com/iotaledger/wasp/tools/stardust-migration/cli"
	"github.com/samber/lo"

	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
)

type TrieRootWithIndex struct {
	Index uint32
	Hash  old_trie.Hash
}

func BuildIndex(s old_state.Store) []TrieRootWithIndex {
	// Copied from indexedstore

	latestState, err := s.LatestState()
	if err != nil {
		panic(err)
	}

	blockKeepAmount := old_governance.NewStateAccess(latestState).GetBlockKeepAmount()
	if blockKeepAmount == -1 {
		// pruning is not enabled - we can get any block just by index
		return []TrieRootWithIndex{{Index: latestState.BlockIndex(), Hash: latestState.TrieRoot()}}
	}

	edgeTrieRoots := make([]TrieRootWithIndex, 0, latestState.BlockIndex()/uint32(blockKeepAmount)+1)
	state := latestState
	blockIndex := state.BlockIndex()
	trieRoot := state.TrieRoot()

	for {
		edgeTrieRoots = append(edgeTrieRoots, TrieRootWithIndex{Index: blockIndex, Hash: trieRoot})

		earliestAvailableBlockIndex := uint32(0)
		if uint32(blockKeepAmount) >= blockIndex {
			// reached the beginning of the chain
			break
		}

		earliestAvailableBlockIndex = blockIndex - uint32(blockKeepAmount) + 1

		bi, ok := old_blocklog.NewStateAccess(state).BlockInfo(earliestAvailableBlockIndex + 1)
		if !ok {
			panic(fmt.Errorf("blocklog missing block index %d on active state %d", earliestAvailableBlockIndex, blockIndex))
		}

		trieRoot = bi.PreviousL1Commitment().TrieRoot()
		state = lo.Must(s.StateByTrieRoot(trieRoot))
		blockIndex = state.BlockIndex()
	}

	slices.Reverse(edgeTrieRoots)

	return edgeTrieRoots
}

func LoadOrCreate(db old_state.Store) *BlockIndexer {
	return LoadOrCreateFromFile(db, defaultIndexFilePath(db))
}

func LoadOrCreateFromFile(db old_state.Store, indexFilePath string) *BlockIndexer {
	cli.Logf("Loading index from %v...\n", indexFilePath)

	indexBytes, err := os.ReadFile(indexFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}

		cli.Logf("Index file not found at %v, building index...\n", indexFilePath)
		startTime := time.Now()
		index := BuildIndex(db)
		cli.Logf("Index built: time = %v, edge states = %v, first edge index = %v, last edge index = %v.",
			time.Since(startTime), len(index), index[0].Index, index[len(index)-1].Index)

		cli.Logf("Saving index to %v...\n", indexFilePath)
		indexBytes = lo.Must(json.MarshalIndent(index, "", "  "))
		lo.Must0(os.MkdirAll(path.Dir(indexFilePath), 0o755))
		lo.Must0(os.WriteFile(indexFilePath, indexBytes, 0o655))

		return NewIndexer(db, index)
	}

	var index []TrieRootWithIndex
	lo.Must0(json.Unmarshal(indexBytes, &index))

	cli.Logf("Index loaded: edge states = %v, first edge index = %v, last edge index = %v.",
		len(index), index[0].Index, index[len(index)-1].Index)

	return NewIndexer(db, index)
}

func defaultIndexFilePath(db old_state.Store) string {
	trieRoot := lo.Must(db.LatestTrieRoot())
	return path.Join(os.TempDir(), "stardust-migration-blockindex-"+trieRoot.String()+".json")
}

func NewIndexer(db old_state.Store, edgeTrieRoots []TrieRootWithIndex) *BlockIndexer {
	return &BlockIndexer{
		s:             db,
		edgeTrieRoots: slices.Clone(edgeTrieRoots),
	}
}

type BlockIndexer struct {
	s             old_state.Store
	edgeTrieRoots []TrieRootWithIndex
}

func (bi *BlockIndexer) BlockByIndex(index uint32) old_state.Block {
	if latestIndex := bi.LatestBlockIndex(); index > latestIndex {
		panic(fmt.Errorf("block index %v is out of range [0; %v]", index, latestIndex))
	}

	i, exactMatch := slices.BinarySearchFunc(bi.edgeTrieRoots, index, func(i TrieRootWithIndex, index uint32) int {
		// unsigned... cannot use i.Index - j
		if i.Index < index {
			return -1
		}
		if i.Index > index {
			return 1
		}
		return 0
	})

	if exactMatch {
		block := lo.Must(bi.s.BlockByTrieRoot(bi.edgeTrieRoots[i].Hash))
		if block.StateIndex() != index {
			// Just double checking
			// TODO: remove for perf
			panic(fmt.Errorf("unexpected block index %v, expected %v", block.StateIndex(), index))
		}

		return block
	}
	if i >= len(bi.edgeTrieRoots) {
		panic("unexpected")
	}

	edgeStateRoot := bi.edgeTrieRoots[i].Hash
	state := lo.Must(bi.s.StateByTrieRoot(edgeStateRoot))

	nextBlockInfo, ok := old_blocklog.NewStateAccess(state).BlockInfo(index + 1)
	if !ok {
		panic(fmt.Errorf("state %v does not have info about block %v", state.BlockIndex(), index+1))
	}

	block := lo.Must(bi.s.BlockByTrieRoot(nextBlockInfo.PreviousL1Commitment().TrieRoot()))
	if block.StateIndex() != index {
		// Just double checking
		// TODO: remove for perf
		panic(fmt.Errorf("unexpected block index %v, expected %v", block.StateIndex(), index))
	}

	return block
}

func (bi *BlockIndexer) LatestBlockIndex() uint32 {
	return bi.edgeTrieRoots[len(bi.edgeTrieRoots)-1].Index
}
