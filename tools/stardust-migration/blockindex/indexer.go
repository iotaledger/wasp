package blockindex

import (
	"fmt"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	"github.com/samber/lo"
)

type BlockIndex interface {
	BlockByIndex(index uint32) (old_state.Block, old_trie.Hash)
	BlocksCount() uint32
}

func New(store old_state.Store) BlockIndex {
	const indexFilePath = "index.bin"
	cli.Logf("Trying to read index from %v", indexFilePath)

	fileIndexer, fileIndexFound := NewFileIndexer(indexFilePath, store)
	if fileIndexFound {
		return fileIndexer
	}

	cli.Logf("Index file NOT found at %v, using on-the-fly indexing", indexFilePath)

	// Index file is not available - using on-the-fly indexer
	indexer := NewOnTheFlyIndexer(store)
	printIndexerStats(indexer, store)

	return indexer
}

func printIndexerStats(indexer *OnTheFlyBlockIndexer, s old_state.Store) {
	latestBlockIndex := lo.Must(s.LatestBlockIndex())
	utils.MeasureTimeAndPrint("Time for retrieving block 0", func() { indexer.BlockByIndex(0) })
	utils.MeasureTimeAndPrint("Time for retrieving block 100", func() { indexer.BlockByIndex(100) })
	utils.MeasureTimeAndPrint("Time for retrieving block 10000", func() { indexer.BlockByIndex(10000) })
	utils.MeasureTimeAndPrint("Time for retrieving block 1000000", func() { indexer.BlockByIndex(1000000) })
	utils.MeasureTimeAndPrint(fmt.Sprintf("Time for retrieving block %v", latestBlockIndex-1000), func() { indexer.BlockByIndex(latestBlockIndex - 1000) })
	utils.MeasureTimeAndPrint(fmt.Sprintf("Time for retrieving block %v", latestBlockIndex), func() { indexer.BlockByIndex(latestBlockIndex) })
}
