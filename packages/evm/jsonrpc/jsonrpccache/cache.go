package jsonrpccache

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

type Cache struct {
	lastBlockIndexCached *uint32
	blockchainDB         func(chainState state.State) *emulator.BlockchainDB
	blockTrieHashByIndex map[uint32]*trie.Hash
	blockByHash          map[common.Hash]*types.Block
	blockByNum           map[uint64]*types.Block
	txBlockByHash        map[common.Hash]*types.Block
	txsByBlockHash       map[common.Hash]types.Transactions
	mutex                sync.RWMutex
}

func New(blockchainDB func(chainState state.State) *emulator.BlockchainDB) *Cache {
	return &Cache{
		lastBlockIndexCached: nil,
		blockchainDB:         blockchainDB,
		blockTrieHashByIndex: make(map[uint32]*trie.Hash),
		blockByHash:          make(map[common.Hash]*types.Block),
		blockByNum:           make(map[uint64]*types.Block),
		txBlockByHash:        make(map[common.Hash]*types.Block),
		txsByBlockHash:       make(map[common.Hash]types.Transactions),
		mutex:                sync.RWMutex{},
	}
}

func (c *Cache) CacheBlock(trieRoot trie.Hash, stateByTrieRoot func(trieRoot trie.Hash) (state.State, error)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	state, err := stateByTrieRoot(trieRoot)
	if err != nil {
		panic(err)
	}
	blockKeepAmount := governance.NewStateAccess(state).GetBlockKeepAmount()
	if blockKeepAmount == -1 {
		return // pruning disabled, never cache anything
	}
	// cache the block that will be pruned next (this way reorgs are okay, as long as it never reorgs more than `blockKeepAmount`, which would be catastrophic)
	if state.BlockIndex() < uint32(blockKeepAmount-1) {
		return
	}
	blockIndexToCache := state.BlockIndex() - uint32(blockKeepAmount-1)
	cacheUntil := blockIndexToCache
	if c.lastBlockIndexCached != nil {
		cacheUntil = *c.lastBlockIndexCached
	}

	// we need to look at the next block to get the trie commitment of the block we want to cache
	nextBlockInfo, found := blocklog.NewStateAccess(state).BlockInfo(blockIndexToCache + 1)
	if !found {
		panic(fmt.Errorf("block %d not found on active state %d", blockIndexToCache, state.BlockIndex()))
	}

	// start in the active state of the block to cache
	activeStateToCache, err := stateByTrieRoot(nextBlockInfo.PreviousL1Commitment().TrieRoot())
	if err != nil {
		panic(err)
	}

	for i := blockIndexToCache; i >= cacheUntil; i-- {
		// walk back and save all blocks between [lastBlockIndexCached...blockIndexToCache]

		blockinfo, found := blocklog.NewStateAccess(activeStateToCache).BlockInfo(i)
		if !found {
			panic(fmt.Errorf("block %d not found on active state %d", i, state.BlockIndex()))
		}

		db := c.blockchainDB(activeStateToCache)
		blockTrieHash := activeStateToCache.TrieRoot()
		c.blockTrieHashByIndex[i] = &blockTrieHash

		evmBlock := db.GetCurrentBlock()
		c.blockByHash[evmBlock.Hash()] = evmBlock

		c.blockByNum[uint64(i)] = evmBlock
		blockTransactions := evmBlock.Transactions()
		c.txsByBlockHash[evmBlock.Hash()] = blockTransactions
		for _, tx := range blockTransactions {
			c.txBlockByHash[tx.Hash()] = evmBlock
		}
		// walk backwards until all blocks are cached
		if i == 0 {
			// nothing more to cache, don't try to walk back further
			break
		}
		activeStateToCache, err = stateByTrieRoot(blockinfo.PreviousL1Commitment().TrieRoot())
		if err != nil {
			panic(err)
		}
	}
	c.lastBlockIndexCached = &blockIndexToCache
}

func (c *Cache) BlockByNumber(n *big.Int) *types.Block {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if n == nil {
		return nil
	}
	return c.blockByNum[n.Uint64()]
}

func (c *Cache) BlockByHash(hash common.Hash) *types.Block {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.blockByHash[hash]
}

func (c *Cache) BlockTrieHashByIndex(n uint32) *trie.Hash {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.blockTrieHashByIndex[n]
}

func (c *Cache) TxByHash(hash common.Hash) (*types.Transaction, *types.Block, uint64) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	block := c.txBlockByHash[hash]
	if block == nil {
		return nil, nil, 0
	}
	for i, tx := range c.txsByBlockHash[block.Hash()] {
		if tx.Hash() == hash {
			return tx, block, uint64(i)
		}
	}
	return nil, nil, 0
}

func (c *Cache) TxByBlockHashAndIndex(hash common.Hash, index uint64) (tx *types.Transaction, blockNumber uint64) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	txs, exists := c.txsByBlockHash[hash]
	if !exists || index >= uint64(len(txs)) {
		return nil, 0
	}
	block := c.blockByHash[hash]
	return txs[index], block.NumberU64()
}

func (c *Cache) TxByBlockNumberAndIndex(blockNumber *big.Int, index uint64) (tx *types.Transaction, blockHash common.Hash) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if blockNumber == nil {
		return nil, common.Hash{}
	}
	block := c.blockByNum[blockNumber.Uint64()]
	if block == nil {
		return nil, common.Hash{}
	}
	txs, exists := c.txsByBlockHash[block.Hash()]
	if !exists || index >= uint64(len(txs)) {
		return nil, common.Hash{}
	}
	return tx, block.Hash()
}
