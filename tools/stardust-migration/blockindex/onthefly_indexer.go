package blockindex

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"

	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
)

type TrieRootWithIndex struct {
	Index uint32
	Hash  old_trie.Hash
}

func buildOnTheFlyIndex(s old_state.Store) []TrieRootWithIndex {
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

func NewOnTheFlyIndexer(db old_state.Store) *OnTheFlyBlockIndexer {
	indexFilePath := onTheFlyIdexerCacheFilePath(db)

	cli.Logf("Loading index cache from %v...", indexFilePath)

	indexBytes, err := os.ReadFile(indexFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}

		cli.Logf("Index cache not found. Building index...")
		startTime := time.Now()
		index := buildOnTheFlyIndex(db)
		cli.Logf("Index built: time = %v, edge states = %v, first edge index = %v, last edge index = %v.",
			time.Since(startTime), len(index), index[0].Index, index[len(index)-1].Index)

		cli.Logf("Saving index cache to %v...", indexFilePath)
		indexBytes = lo.Must(json.MarshalIndent(index, "", "  "))
		lo.Must0(os.MkdirAll(path.Dir(indexFilePath), 0o755))
		lo.Must0(os.WriteFile(indexFilePath, indexBytes, 0o655))

		return &OnTheFlyBlockIndexer{
			s:             db,
			edgeTrieRoots: index,
		}
	}

	var index []TrieRootWithIndex
	lo.Must0(json.Unmarshal(indexBytes, &index))

	cli.Logf("Index cache loaded: edge states = %v, first edge index = %v, last edge index = %v.",
		len(index), index[0].Index, index[len(index)-1].Index)
	latestBlockIndex := lo.Must(db.LatestBlockIndex())

	if index[len(index)-1].Index != latestBlockIndex {
		cli.Logf("** WARNING: index cache was built for other database: latest block index in db = %v, last index entry = %v",
			latestBlockIndex, index[len(index)-1].Index)
		if err := os.Remove(indexFilePath); err != nil {
			panic(fmt.Errorf("failed to remove index cache file %v: %w", indexFilePath, err))
		}

		cli.Logf("Rebuilding index...")
		return NewOnTheFlyIndexer(db)
	}

	return &OnTheFlyBlockIndexer{
		s:             db,
		edgeTrieRoots: index,
	}
}

func onTheFlyIdexerCacheFilePath(db old_state.Store) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	trieRoot := lo.Must(db.LatestTrieRoot())
	return path.Join(cwd, "stardust-migration-blockindex-"+trieRoot.String()+".json")
}

type OnTheFlyBlockIndexer struct {
	s             old_state.Store
	edgeTrieRoots []TrieRootWithIndex
}

var _ BlockIndex = (*OnTheFlyBlockIndexer)(nil)

func (bi *OnTheFlyBlockIndexer) BlockByIndex(index uint32) (old_state.Block, old_trie.Hash) {
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
		trieRoot := bi.edgeTrieRoots[i].Hash
		block := lo.Must(bi.s.BlockByTrieRoot(trieRoot))
		if block.StateIndex() != index {
			// Just double checking (can be removed for perf)
			panic(fmt.Errorf("unexpected block index %v, expected %v", block.StateIndex(), index))
		}

		return block, trieRoot
	}
	if i >= len(bi.edgeTrieRoots) {
		panic("unexpected")
	}

	edgeTrieRoot := bi.edgeTrieRoots[i].Hash
	state := lo.Must(bi.s.StateByTrieRoot(edgeTrieRoot))

	nextBlockInfo, ok := old_blocklog.NewStateAccess(state).BlockInfo(index + 1)
	if !ok {
		panic(fmt.Errorf("state %v does not have info about block %v", state.BlockIndex(), index+1))
	}

	trieRoot := nextBlockInfo.PreviousL1Commitment().TrieRoot()
	block := lo.Must(bi.s.BlockByTrieRoot(trieRoot))
	if block.StateIndex() != index {
		// Just double checking (can be removed for perf)
		panic(fmt.Errorf("unexpected block index %v, expected %v", block.StateIndex(), index))
	}

	return block, trieRoot
}

func (bi *OnTheFlyBlockIndexer) LatestBlockIndex() uint32 {
	return bi.edgeTrieRoots[len(bi.edgeTrieRoots)-1].Index
}

func (bi *OnTheFlyBlockIndexer) BlocksCount() uint32 {
	return bi.LatestBlockIndex() + 1
}
