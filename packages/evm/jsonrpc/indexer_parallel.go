package jsonrpc

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

type Checkpoint struct {
	StartBlock uint32
	EndBlock   uint32
	TrieRoot   trie.Hash
}

func (c *Index) loadOrCreateCheckpoints(store indexedstore.IndexedStore, blockIndexToCache, checkpointInterval uint32) ([]Checkpoint, error) {
	checkpointFilePath := filepath.Join("/tmp/checkpoints.bin")

	// Calculate checkpoint blocks
	var checkpointBlocks []uint32

	// Start with the highest block and go down in intervals
	blockIdx := blockIndexToCache
	for {
		if blockIdx != 0 {
			checkpointBlocks = append(checkpointBlocks, blockIdx)
		}

		// Move to next checkpoint
		if blockIdx <= checkpointInterval {
			break
		}

		blockIdx -= checkpointInterval
		if blockIdx < 0 {
			blockIdx = 0
		}
	}

	fmt.Printf("Need %d checkpoints at %d block intervals\n", len(checkpointBlocks), checkpointInterval)

	// Try to load checkpoints from file
	checkpoints := make([]Checkpoint, 0)

	// Check if file exists
	if _, err := os.Stat(checkpointFilePath); err == nil {
		// File exists, try to load it
		data, err := os.ReadFile(checkpointFilePath)
		if err == nil {
			// Try to unmarshal
			checkpoints = bcs.MustUnmarshal[[]Checkpoint](data)
			return checkpoints, nil
		} else {
			fmt.Printf("Warning: could not read checkpoint file: %v\n", err)
		}
	}

	for i := 0; i < len(checkpointBlocks); i++ {
		startBlock := checkpointBlocks[i]
		endBlock := uint32(0)

		if i+1 == len(checkpointBlocks) {
			if startBlock > 10000 {
				endBlock = startBlock - 9999
			} else {

				endBlock = 0
			}
		} else {
			endBlock = checkpointBlocks[i+1] + 1
		}

		fmt.Printf("Creating checkpoint for block %d\n", startBlock)

		blockinfo, err := store.BlockByIndex(startBlock)
		if err != nil {
			return nil, fmt.Errorf("block %d not found: %w", startBlock, err)
		}

		checkpoints = append(checkpoints, Checkpoint{
			StartBlock: startBlock,
			EndBlock:   uint32(endBlock),
			TrieRoot:   blockinfo.TrieRoot(),
		})
	}

	data := bcs.MustMarshal[[]Checkpoint](&checkpoints)
	if err := os.WriteFile(checkpointFilePath, data, 0644); err != nil {
		fmt.Printf("Warning: failed to save checkpoints to file: %v\n", err)
	} else {
		fmt.Printf("Saved %d checkpoints to file\n", len(checkpoints))
	}

	return checkpoints, nil
}

type blockData struct {
	blockIndex        uint32
	blockTrieRoot     trie.Hash
	blockHash         common.Hash
	transactionHashes []common.Hash
}

func (c *Index) IndexBlockParallel(store indexedstore.IndexedStore, trieRoot trie.Hash, numWorkers int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("Starting parallel block indexing")

	state, err := c.stateByTrieRoot(trieRoot)
	if err != nil {
		return fmt.Errorf("stateByTrieRoot: %w", err)
	}

	blockIndexToCache := state.BlockIndex()

	fmt.Printf("Indexing %d blocks \n", blockIndexToCache)

	const checkpointInterval = 10000

	checkpoints, err := c.loadOrCreateCheckpoints(store, blockIndexToCache, checkpointInterval)
	if err != nil {
		return fmt.Errorf("loadOrCreateCheckpoints: %w", err)
	}

	fmt.Printf("Using %d checkpoints for processing\n", len(checkpoints))

	resultChan := make(chan []blockData, len(checkpoints))
	errChan := make(chan error, len(checkpoints))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, numWorkers)

	for _, checkpoint := range checkpoints {
		semaphore <- struct{}{}
		wg.Add(1)

		go func(cp Checkpoint) {
			defer wg.Done()
			defer func() { <-semaphore }()

			fmt.Printf("Processing checkpoint from block %d to %d\n", cp.StartBlock, cp.EndBlock)

			batchResults, err := c.processCheckpoint(cp)
			if err != nil {
				fmt.Printf("Error processing checkpoint %d: %v\n", cp.StartBlock, err)
				errChan <- fmt.Errorf("processCheckpoint: %w", err)
			}

			resultChan <- batchResults

		}(checkpoint)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Collect all results
	var allResults []blockData
	for batchResults := range resultChan {
		allResults = append(allResults, batchResults...)
	}

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Sort results by block index (descending)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].blockIndex > allResults[j].blockIndex
	})

	err = c.writeResultsToIndex(allResults, blockIndexToCache)
	if err != nil {
		return fmt.Errorf("writeResultsToIndex: %w", err)
	}

	c.setLastBlockIndexed(blockIndexToCache)
	c.store.Flush()
	fmt.Printf("Successfully indexed blocks from %d to 0\n", blockIndexToCache)
	return nil
}

func (c *Index) processCheckpoint(cp Checkpoint) ([]blockData, error) {
	checkpointState, err := c.stateByTrieRoot(cp.TrieRoot)
	if err != nil {
		return nil, fmt.Errorf("stateByTrieRoot for checkpoint %d: %w", cp.StartBlock, err)
	}

	db := blockchainDB(checkpointState)

	var results []blockData
	blockIdx := cp.StartBlock

	for {
		if blockIdx%1000 == 0 {
			fmt.Printf("Processing block %d\n", blockIdx)
		}

		var trieRoot trie.Hash
		if blockIdx == cp.StartBlock {
			trieRoot = checkpointState.TrieRoot()
		} else {
			iscBlock, ok := blocklog.NewStateReaderFromChainState(checkpointState).GetBlockInfo(blockIdx + 1)
			if !ok {
				return nil, fmt.Errorf("failed to get isc block %d (+1). Checkpoint: start: %d / end: %d", blockIdx, cp.StartBlock, cp.EndBlock)
			}
			trieRoot = iscBlock.PreviousL1Commitment().TrieRoot()
		}

		evmBlock := db.GetBlockByNumber(uint64(blockIdx))
		if evmBlock == nil {
			return nil, fmt.Errorf("block %d not found in checkpoint range", blockIdx)
		}

		txHashes := make([]common.Hash, 0, len(evmBlock.Transactions()))
		for _, tx := range evmBlock.Transactions() {
			txHashes = append(txHashes, tx.Hash())
		}

		results = append(results, blockData{
			blockIndex:        blockIdx,
			blockTrieRoot:     trieRoot,
			blockHash:         evmBlock.Hash(),
			transactionHashes: txHashes,
		})

		if blockIdx == cp.EndBlock {
			break
		}

		blockIdx--
	}

	return results, nil
}

// writeResultsToIndex writes all collected data to the index
func (c *Index) writeResultsToIndex(results []blockData, blockIndexToCache uint32) error {
	fmt.Println("Writing collected data to index...")

	// Track which blocks have been processed
	processed := make(map[uint32]bool)

	// Write all results to the index
	for _, data := range results {
		fmt.Printf("Writing block: %d\n	TrieRoot: %s\n	BlockHash: %s\n", data.blockIndex, data.blockTrieRoot.String(), data.blockHash.String())

		c.setBlockTrieRootByIndex(data.blockIndex, data.blockTrieRoot)
		c.setBlockIndexByHash(data.blockHash, data.blockIndex)

		for _, txHash := range data.transactionHashes {
			c.setBlockIndexByTxHash(txHash, data.blockIndex)
			fmt.Printf("		TX Hash for Block: %d: %s\n", data.blockIndex, txHash.String())
		}

		processed[data.blockIndex] = true
	}

	// Verify all blocks were processed
	blockIdx := blockIndexToCache
	for {
		if !processed[blockIdx] {
			return fmt.Errorf("block %d was not processed", blockIdx)
		}

		if blockIdx == 0 {
			break
		}

	}

	return nil
}
