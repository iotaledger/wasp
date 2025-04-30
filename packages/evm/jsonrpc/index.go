package jsonrpc

import (
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// Index allows efficient retrieval of EVM blocks and transactions given the block number.
// The indexed information is stored in a database (either persisted or in-memory).
type Index struct {
	store           kvstore.KVStore
	stateByTrieRoot func(trieRoot trie.Hash) (state.State, error)

	mu sync.RWMutex
}

func NewIndex(
	stateByTrieRoot func(trieRoot trie.Hash) (state.State, error),
	indexDbEngine hivedb.Engine,
	indexDbPath string,
) *Index {
	db, err := database.NewDatabase(indexDbEngine, indexDbPath, true, false, database.CacheSizeDefault)
	if err != nil {
		panic(err)
	}
	return &Index{
		store:           db.KVStore(),
		stateByTrieRoot: stateByTrieRoot,
		mu:              sync.RWMutex{},
	}
}

// IndexBlock is called only in an archive node, whenever a block is published.
//
// It walks back following the previous trie root until the latest cached
// block, associating the EVM block and transaction hashes with the
// corresponding block index.
func (c *Index) IndexBlock(trieRoot trie.Hash) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("Start index block")

	state, err := c.stateByTrieRoot(trieRoot)
	if err != nil {
		return fmt.Errorf("stateByTrieRoot: %w", err)
	}
	blockKeepAmount := governance.NewStateReaderFromChainState(state).GetBlockKeepAmount()
	if blockKeepAmount == -1 {
		return nil // pruning disabled, never cache anything
	}
	// cache the block that will be pruned next (this way reorgs are okay, as long as it never reorgs more than `blockKeepAmount`, which would be catastrophic)
	if state.BlockIndex() < uint32(blockKeepAmount-1) {
		return nil
	}
	blockIndexToCache := state.BlockIndex() - uint32(blockKeepAmount-1)
	cacheUntil := uint32(0)
	lastBlockIndexed := c.lastBlockIndexed()
	if lastBlockIndexed != nil {
		cacheUntil = *lastBlockIndexed
	}

	// we need to look at the next block to get the trie commitment of the block we want to cache
	nextBlockInfo, found := blocklog.NewStateReaderFromChainState(state).GetBlockInfo(blockIndexToCache + 1)
	if !found {
		return fmt.Errorf("block %d not found on active state %d", blockIndexToCache, state.BlockIndex())
	}

	// start in the active state of the block to cache
	activeStateToCache, err := c.stateByTrieRoot(nextBlockInfo.PreviousL1Commitment().TrieRoot())
	if err != nil {
		return fmt.Errorf("stateByTrieRoot: %w", err)
	}

	for i := blockIndexToCache; i >= cacheUntil; i-- {
		// walk back and save all blocks between [lastBlockIndexCached...blockIndexToCache]

		if i%1000 == 0 {
			fmt.Printf("indexing block %d, cacheUntil: %d \n", i, cacheUntil)
		}

		blockinfo, found := blocklog.NewStateReaderFromChainState(activeStateToCache).GetBlockInfo(i)
		if !found {
			return fmt.Errorf("block %d not found on active state %d", i, state.BlockIndex())
		}

		db := blockchainDB(activeStateToCache)
		blockTrieRoot := activeStateToCache.TrieRoot()
		c.setBlockTrieRootByIndex(i, blockTrieRoot)

		evmBlock := db.GetCurrentBlock()
		c.setBlockIndexByHash(evmBlock.Hash(), i)

		blockTransactions := evmBlock.Transactions()
		for _, tx := range blockTransactions {
			c.setBlockIndexByTxHash(tx.Hash(), i)
		}
		// walk backwards until all blocks are cached
		if i == 0 {
			// nothing more to cache, don't try to walk back further
			break
		}
		activeStateToCache, err = c.stateByTrieRoot(blockinfo.PreviousL1Commitment().TrieRoot())
		if err != nil {
			return fmt.Errorf("stateByTrieRoot: %w", err)
		}
	}
	c.setLastBlockIndexed(blockIndexToCache)
	c.store.Flush()
	return nil
}

func (c *Index) BlockByNumber(n *big.Int) *types.Block {
	if n == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.evmDBFromBlockIndex(uint32(n.Uint64()))
	if db == nil {
		return nil
	}
	return db.GetBlockByNumber(n.Uint64())
}

func (c *Index) BlockByHash(hash common.Hash) *types.Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blockIndex := c.blockIndexByHash(hash)
	if blockIndex == nil {
		return nil
	}
	return c.evmDBFromBlockIndex(*blockIndex).GetBlockByHash(hash)
}

func (c *Index) BlockTrieRootByIndex(n uint32) *trie.Hash {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.blockTrieRootByIndex(n)
}

func (c *Index) TxByHash(hash common.Hash) (tx *types.Transaction, blockHash common.Hash, blockNumber, txIndex uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blockIndex := c.blockIndexByTxHash(hash)
	if blockIndex == nil {
		return nil, common.Hash{}, 0, 0
	}
	tx, blockHash, blockNumber, txIndex, err := c.evmDBFromBlockIndex(*blockIndex).GetTransactionByHash(hash)
	if err != nil {
		panic(err)
	}
	return tx, blockHash, blockNumber, txIndex
}

func (c *Index) GetReceiptByTxHash(hash common.Hash) *types.Receipt {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blockIndex := c.blockIndexByTxHash(hash)
	if blockIndex == nil {
		return nil
	}
	return c.evmDBFromBlockIndex(*blockIndex).GetReceiptByTxHash(hash)
}

func (c *Index) TxByBlockHashAndIndex(blockHash common.Hash, txIndex uint64) (tx *types.Transaction, blockNumber uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blockIndex := c.blockIndexByHash(blockHash)
	if blockIndex == nil {
		return nil, 0
	}
	block := c.evmDBFromBlockIndex(*blockIndex).GetBlockByHash(blockHash)
	if block == nil {
		return nil, 0
	}
	txs := block.Transactions()
	if txIndex >= uint64(len(txs)) {
		return nil, 0
	}
	return txs[txIndex], block.NumberU64()
}

func (c *Index) TxByBlockNumberAndIndex(blockNumber *big.Int, txIndex uint64) (tx *types.Transaction, blockHash common.Hash) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if blockNumber == nil {
		return nil, common.Hash{}
	}
	db := c.evmDBFromBlockIndex(uint32(blockNumber.Uint64()))
	if db == nil {
		return nil, common.Hash{}
	}
	block := db.GetBlockByHash(blockHash)
	if block == nil {
		return nil, common.Hash{}
	}
	txs := block.Transactions()
	if txIndex >= uint64(len(txs)) {
		return nil, common.Hash{}
	}
	return txs[txIndex], block.Hash()
}

func (c *Index) TxsByBlockNumber(blockNumber *big.Int) types.Transactions {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if blockNumber == nil {
		return nil
	}
	db := c.evmDBFromBlockIndex(uint32(blockNumber.Uint64()))
	if db == nil {
		return nil
	}
	block := db.GetBlockByNumber(blockNumber.Uint64())
	if block == nil {
		return nil
	}
	return block.Transactions()
}

// internals

const (
	prefixLastBlockIndexed     = iota // ISC block index (uint32)
	prefixBlockTrieRootByIndex        // ISC block index (uint32) => ISC trie root
	prefixBlockIndexByTxHash          // EVM tx hash => ISC block index (uint32)
	prefixBlockIndexByHash            // EVM block hash => ISC block index (uint32)
)

func keyLastBlockIndexed() kvstore.Key {
	return []byte{prefixLastBlockIndexed}
}

func keyBlockTrieRootByIndex(i uint32) kvstore.Key {
	key := []byte{prefixBlockTrieRootByIndex}
	key = append(key, codec.Encode(i)...)
	return key
}

func keyBlockIndexByTxHash(hash common.Hash) kvstore.Key {
	key := []byte{prefixBlockIndexByTxHash}
	key = append(key, hash[:]...)
	return key
}

func keyBlockIndexByHash(hash common.Hash) kvstore.Key {
	key := []byte{prefixBlockIndexByHash}
	key = append(key, hash[:]...)
	return key
}

func (c *Index) get(key kvstore.Key) []byte {
	ret, err := c.store.Get(key)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil
		}
		panic(err)
	}
	return ret
}

func (c *Index) set(key kvstore.Key, value []byte) {
	err := c.store.Set(key, value)
	if err != nil {
		panic(err)
	}
}

func (c *Index) setLastBlockIndexed(n uint32) {
	c.set(keyLastBlockIndexed(), codec.Encode(n))
}

func (c *Index) lastBlockIndexed() *uint32 {
	bytes := c.get(keyLastBlockIndexed())
	if bytes == nil {
		return nil
	}
	ret := codec.MustDecode[uint32](bytes)
	return &ret
}

func (c *Index) setBlockTrieRootByIndex(i uint32, hash trie.Hash) {
	c.set(keyBlockTrieRootByIndex(i), hash.Bytes())
}

func (c *Index) blockTrieRootByIndex(i uint32) *trie.Hash {
	bytes := c.get(keyBlockTrieRootByIndex(i))
	if bytes == nil {
		return nil
	}
	hash, err := trie.HashFromBytes(bytes)
	if err != nil {
		panic(err)
	}
	return &hash
}

func (c *Index) setBlockIndexByTxHash(txHash common.Hash, blockIndex uint32) {
	c.set(keyBlockIndexByTxHash(txHash), codec.Encode(blockIndex))
}

func (c *Index) blockIndexByTxHash(txHash common.Hash) *uint32 {
	bytes := c.get(keyBlockIndexByTxHash(txHash))
	if bytes == nil {
		return nil
	}
	ret := codec.MustDecode[uint32](bytes)
	return &ret
}

func (c *Index) setBlockIndexByHash(hash common.Hash, blockIndex uint32) {
	c.set(keyBlockIndexByHash(hash), codec.Encode(blockIndex))
}

func (c *Index) blockIndexByHash(hash common.Hash) *uint32 {
	bytes := c.get(keyBlockIndexByHash(hash))
	if bytes == nil {
		return nil
	}
	ret := codec.MustDecode[uint32](bytes)
	return &ret
}

func (c *Index) evmDBFromBlockIndex(n uint32) *emulator.BlockchainDB {
	trieRoot := c.blockTrieRootByIndex(n)
	if trieRoot == nil {
		return nil
	}
	state, err := c.stateByTrieRoot(*trieRoot)
	if err != nil {
		panic(err)
	}
	return blockchainDB(state)
}
