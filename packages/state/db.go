// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"errors"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/v2/packages/chaindb"
	"github.com/iotaledger/wasp/v2/packages/kv/buffered"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

var (
	ErrTrieRootNotFound      = errors.New("trie root not found")
	ErrUnknownLatestTrieRoot = errors.New("latest trie root is unknown")
	ErrNoBlocksPruned        = errors.New("no blocks were pruned from the store yet")
)

func keyBlockByTrieRootNoTrieRoot() []byte {
	return []byte{chaindb.PrefixBlockByTrieRoot}
}

func keyBlockByTrieRoot(root trie.Hash) []byte {
	return append(keyBlockByTrieRootNoTrieRoot(), root.Bytes()...)
}

func keyLatestTrieRoot() []byte {
	return []byte{chaindb.PrefixLatestTrieRoot}
}

func keyLargestPrunedBlockIndex() []byte {
	return []byte{chaindb.PrefixLargestPrunedBlockIndex}
}

func mustNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

// storeDB encapsulates all data that is stored on the store's DB.
// The DB includes:
// - The trie storage, under the prefixTrie partition. This includes one trie root for each state index.
// - One block per trie root, under the prefixBlockByTrieRoot partition.
// - The trie root that is considered 'latest' in the chain, under prefixLatestTrieRoot
type storeDB struct {
	kvstore.KVStore
}

func (db *storeDB) latestTrieRoot() (trie.Hash, error) {
	if !db.hasLatestTrieRoot() {
		return trie.Hash{}, ErrUnknownLatestTrieRoot
	}
	b := db.mustGet(keyLatestTrieRoot())
	ret, err := trie.HashFromBytes(b)
	mustNoErr(err)
	return ret, nil
}

func (db *storeDB) hasLatestTrieRoot() bool {
	return db.mustHas(keyLatestTrieRoot())
}

func (db *storeDB) setLatestTrieRoot(root trie.Hash) {
	db.mustSet(keyLatestTrieRoot(), root.Bytes())
}

func (db *storeDB) largestPrunedBlockIndex() (uint32, error) {
	if !db.hasLargestPrunedBlockIndex() {
		return 0, ErrNoBlocksPruned
	}
	b := db.mustGet(keyLargestPrunedBlockIndex())
	ret := codec.MustDecode[uint32](b)
	return ret, nil
}

func (db *storeDB) hasLargestPrunedBlockIndex() bool {
	return db.mustHas(keyLargestPrunedBlockIndex())
}

func (db *storeDB) setLargestPrunedBlockIndex(blockIndex uint32) {
	db.mustSet(keyLargestPrunedBlockIndex(), codec.Encode[uint32](blockIndex))
}

func (db *storeDB) isEmpty() bool {
	empty := true
	err := db.Iterate(keyBlockByTrieRootNoTrieRoot(), func(kvstore.Key, kvstore.Value) bool {
		empty = false
		return false
	})
	mustNoErr(err)
	return empty
}

func trieStore(db kvstore.KVStore) trie.KVStore {
	return trie.NewHiveKVStoreAdapter(db, []byte{chaindb.PrefixTrie})
}

func (db *storeDB) trieUpdatable(root trie.Hash) (*trie.TrieUpdatable, error) {
	return trie.NewTrieUpdatable(trieStore(db), root)
}

func (db *storeDB) initTrie(refcountsEnabled bool) (trie.Hash, error) {
	return trie.InitRoot(trieStore(db), refcountsEnabled)
}

func (db *storeDB) trieReader(root trie.Hash) (*trie.TrieReader, error) {
	return trieReader(trieStore(db), root)
}

func trieReader(trieStore trie.KVStore, root trie.Hash) (*trie.TrieReader, error) {
	return trie.NewTrieReader(trieStore, root)
}

func (db *storeDB) hasBlock(root trie.Hash) bool {
	return db.mustHas(keyBlockByTrieRoot(root))
}

func (db *storeDB) saveBlock(block Block) {
	db.mustSet(keyBlockByTrieRoot(block.TrieRoot()), block.Bytes())
}

func (db *storeDB) pruneBlock(trieRoot trie.Hash) {
	db.mustDel(keyBlockByTrieRoot(trieRoot))
}

func (db *storeDB) readBlock(root trie.Hash) (Block, error) {
	key := keyBlockByTrieRoot(root)
	if !db.mustHas(key) {
		return nil, fmt.Errorf("%w %s", ErrTrieRootNotFound, root)
	}
	return BlockFromBytes(db.mustGet(key))
}

func (db *storeDB) commitToDB(muts *buffered.Mutations) {
	batch, err := db.Batched()
	mustNoErr(err)
	for k, v := range muts.Sets {
		err = batch.Set([]byte(k), v)
		mustNoErr(err)
	}
	for k := range muts.Dels {
		err = batch.Delete([]byte(k))
		mustNoErr(err)
	}
	err = batch.Commit()
	mustNoErr(err)
}

func (db *storeDB) mustSet(key []byte, value []byte) {
	err := db.Set(key, value)
	mustNoErr(err)
}

func (db *storeDB) mustDel(key []byte) {
	err := db.Delete(key)
	mustNoErr(err)
}

func (db *storeDB) mustHas(key []byte) bool {
	has, err := db.Has(key)
	mustNoErr(err)
	return has
}

func (db *storeDB) mustGet(key []byte) []byte {
	v, err := db.Get(key)
	mustNoErr(err)
	return v
}

func (db *storeDB) buffered() (*bufferedKVStore, *storeDB) {
	buf := newBufferedKVStore(db)
	return buf, &storeDB{buf}
}

// increment when changing the snapshot format
const snapshotVersion = 0

func (db *storeDB) takeSnapshot(root trie.Hash, w io.Writer) error {
	block, err := db.readBlock(root)
	if err != nil {
		return err
	}
	ww := rwutil.NewWriter(w)
	ww.WriteUint8(snapshotVersion)
	ww.WriteBytes(block.Bytes())
	if ww.Err != nil {
		return ww.Err
	}
	trie, err := db.trieReader(block.TrieRoot())
	if err != nil {
		return err
	}
	return trie.TakeSnapshot(w)
}

func (db *storeDB) restoreSnapshot(root trie.Hash, r io.Reader, refcountsEnabled bool) error {
	rr := rwutil.NewReader(r)
	v := rr.ReadUint8()
	if v != snapshotVersion {
		return errors.New("snapshot version mismatch")
	}
	blockBytes := rr.ReadBytes()
	if rr.Err != nil {
		return rr.Err
	}
	block, err := BlockFromBytes(blockBytes)
	if err != nil {
		return err
	}
	if block.TrieRoot() != root {
		return errors.New("trie root mismatch")
	}
	db.saveBlock(block)

	err = trie.RestoreSnapshot(r, trieStore(db), refcountsEnabled)
	if err != nil {
		return err
	}
	return nil
}
