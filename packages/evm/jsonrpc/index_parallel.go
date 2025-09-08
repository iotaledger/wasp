package jsonrpc

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
)

/**
 * This parallel indexer implementation is able to re-index a whole db in parallel
 * It is not meant to be run by our Indexer itself. Rather by external tooling, like the wasp-cli.
 */

type Checkpoint struct {
	StartBlock uint32
	EndBlock   uint32
	TrieRoot   trie.Hash
}

// Our block retention is 10k, so this constant is needed.
// 9999 due to implementation details. But it catches all 10k.
const checkpointInterval = 9999

func (c *Index) loadOrCreateCheckpoints(log log.Logger, store indexedstore.IndexedStore, latestBlockIndex, checkpointInterval uint32) ([]Checkpoint, error) {
	// Building a list of checkpoints first, so the actual build of the index db runs faster.
	checkpointFilePath := "/tmp/checkpoints.bin"

	var checkpointBlocks []uint32

	blockIdx := latestBlockIndex
	for {
		if blockIdx != 0 {
			checkpointBlocks = append(checkpointBlocks, blockIdx)
		}

		// Move to next checkpoint
		if blockIdx <= checkpointInterval {
			break
		}

		blockIdx -= checkpointInterval
	}

	log.LogInfof("Need %d checkpoints at %d block intervals\n", len(checkpointBlocks), checkpointInterval)

	checkpoints := make([]Checkpoint, 0)

	if _, err := os.Stat(checkpointFilePath); err == nil {
		data, err := os.ReadFile(checkpointFilePath)
		if err == nil {
			checkpoints = bcs.MustUnmarshal[[]Checkpoint](data)

			if checkpoints[0].StartBlock != latestBlockIndex {
				panic(fmt.Sprintf("Invalid checkpoint db (Unexpected amount of hashes). Remove '%s' and try again.", checkpointFilePath))
			}

			return checkpoints, nil
		} else {
			log.LogInfof("Warning: could not read checkpoint file: %v\n", err)
		}
	}

	for i := 0; i < len(checkpointBlocks); i++ {
		startBlock := checkpointBlocks[i]
		var endBlock uint32

		if i+1 == len(checkpointBlocks) {
			if startBlock > 10000 {
				endBlock = startBlock - checkpointInterval
			} else {
				endBlock = 0
			}
		} else {
			endBlock = checkpointBlocks[i+1] + 1
		}

		log.LogInfof("Creating checkpoint for block %d\n", startBlock)

		blockInfo, err := store.BlockByIndex(startBlock)
		if err != nil {
			return nil, fmt.Errorf("block %d not found: %w", startBlock, err)
		}

		checkpoints = append(checkpoints, Checkpoint{
			StartBlock: startBlock,
			EndBlock:   endBlock,
			TrieRoot:   blockInfo.TrieRoot(),
		})
	}

	data := bcs.MustMarshal[[]Checkpoint](&checkpoints)
	if err := os.WriteFile(checkpointFilePath, data, 0o600); err != nil {
		log.LogInfof("Warning: failed to save checkpoints to file: %v\n", err)
	} else {
		log.LogInfof("Saved %d checkpoints to file\n", len(checkpoints))
	}

	return checkpoints, nil
}

type blockData struct {
	blockIndex        uint32
	blockTrieRoot     trie.Hash
	blockHash         common.Hash
	transactionHashes []common.Hash
}

// IndexAllBlocksInParallel is meant to be used by external tooling, not by the EVM Indexer itself
// It relies on private methods, so it's stored here. Don't ever call this function inside the Indexer itself.
func (c *Index) IndexAllBlocksInParallel(log log.Logger, store func() indexedstore.IndexedStore, trieRoot trie.Hash, workers uint8) error { //nolint: funlen
	c.mu.Lock()
	defer c.mu.Unlock()

	log.LogInfo("Starting parallel block indexing")

	state, err := store().StateByTrieRoot(trieRoot)
	if err != nil {
		return fmt.Errorf("stateByTrieRoot: %w", err)
	}

	latestBlockIndex := state.BlockIndex()

	log.LogInfof("Indexing %d blocks \n", latestBlockIndex)
	log.LogInfo("Loading/Creating checkpoints..")

	checkpoints, err := c.loadOrCreateCheckpoints(log, store(), latestBlockIndex, checkpointInterval)
	if err != nil {
		return fmt.Errorf("loadOrCreateCheckpoints: %w", err)
	}

	log.LogInfo("Collected checkpoints.")
	log.LogInfof("Using %d checkpoints for processing\n", len(checkpoints))
	log.LogInfo("Starting workers..")

	resultChan := make(chan []blockData, len(checkpoints))
	errChan := make(chan error, len(checkpoints))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workers)

	for _, checkpoint := range checkpoints {
		semaphore <- struct{}{}
		wg.Add(1)

		go func(cp Checkpoint) {
			defer wg.Done()
			defer func() { <-semaphore }()

			log.LogInfof("Processing checkpoint from block %d to %d\n", cp.StartBlock, cp.EndBlock)

			batchResults, err2 := c.processCheckpoint(log, store, cp)
			if err2 != nil {
				log.LogInfof("Error processing checkpoint %d: %v\n", cp.StartBlock, err2)
				errChan <- fmt.Errorf("processCheckpoint: %w", err2)
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

	var allResults []blockData
	for batchResults := range resultChan {
		allResults = append(allResults, batchResults...)
	}

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].blockIndex > allResults[j].blockIndex
	})

	err = c.writeResultsToIndex(log, allResults, latestBlockIndex)
	if err != nil {
		return fmt.Errorf("writeResultsToIndex: %w", err)
	}

	c.setLastBlockIndexed(latestBlockIndex)
	err = c.store.Flush()
	if err != nil {
		return fmt.Errorf("store.Flush: %w", err)
	}

	log.LogInfof("Successfully indexed blocks from %d to 0\n", latestBlockIndex)
	return nil
}

func (c *Index) processCheckpoint(log log.Logger, store func() indexedstore.IndexedStore, cp Checkpoint) ([]blockData, error) {
	readStore := store()
	checkpointState, err := readStore.StateByTrieRoot(cp.TrieRoot)
	if err != nil {
		return nil, fmt.Errorf("stateByTrieRoot for checkpoint %d: %w", cp.StartBlock, err)
	}

	db := blockchainDB(checkpointState)

	var results []blockData
	blockIdx := cp.StartBlock

	for {
		if blockIdx%1000 == 0 {
			log.LogInfof("Processing block %d\n", blockIdx)
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

		blockHash := evmBlock.Hash().Hex()

		txHashes := make([]common.Hash, 0, len(evmBlock.Transactions()))
		for _, tx := range evmBlock.Transactions() {
			txHashes = append(txHashes, tx.Hash())
		}

		results = append(results, blockData{
			blockIndex:        blockIdx,
			blockTrieRoot:     trieRoot,
			blockHash:         common.HexToHash(blockHash),
			transactionHashes: txHashes,
		})

		if blockIdx == cp.EndBlock {
			break
		}

		blockIdx--
	}

	return results, nil
}

func (c *Index) writeResultsToIndex(log log.Logger, results []blockData, blockIndexToCache uint32) error {
	log.LogInfo("Writing collected data to index...")

	processed := make(map[uint32]bool)

	for _, data := range results {
		c.setBlockTrieRootByIndex(data.blockIndex, data.blockTrieRoot)
		c.setBlockIndexByHash(data.blockHash, data.blockIndex)

		for _, txHash := range data.transactionHashes {
			c.setBlockIndexByTxHash(txHash, data.blockIndex)
		}

		processed[data.blockIndex] = true

		if data.blockIndex%10000 == 0 {
			c.setLastBlockIndexed(blockIndexToCache)
			c.store.Flush()
		}
	}

	log.LogInfo("Finished writing collected data to index")
	log.LogInfo("Validating included blocks .. ")

	blockIdx := blockIndexToCache
	for {
		if !processed[blockIdx] {
			return fmt.Errorf("block %d was not processed", blockIdx)
		}

		if blockIdx == 0 {
			break
		}

		blockIdx--
	}

	return nil
}
