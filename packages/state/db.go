// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/chaindb"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/trie"
)

var (
	ErrTrieRootNotFound      = errors.New("trie root not found")
	ErrUnknownLatestTrieRoot = errors.New("latest trie root is unknown")
)

func keyBlockByTrieRoot(root trie.Hash) []byte {
	return append([]byte{chaindb.PrefixBlockByTrieRoot}, root.Bytes()...)
}

func keyLatestTrieRoot() []byte {
	return []byte{chaindb.PrefixLatestTrieRoot}
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

func trieStore(db kvstore.KVStore) trie.KVStore {
	return trie.NewHiveKVStoreAdapter(db, []byte{chaindb.PrefixTrie})
}

func (db *storeDB) trieUpdatable(root trie.Hash) (*trie.TrieUpdatable, error) {
	return trie.NewTrieUpdatable(trieStore(db), root)
}

func (db *storeDB) initTrie() trie.Hash {
	return trie.MustInitRoot(trieStore(db))
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

func (db *storeDB) takeSnapshot(root trie.Hash, snapshot kvstore.KVStore) error {
	if !db.hasBlock(root) {
		return fmt.Errorf("cannot take snapshot: trie root not found: %s", root)
	}
	blockKey := keyBlockByTrieRoot(root)
	err := snapshot.Set(blockKey, db.mustGet(blockKey))
	if err != nil {
		return err
	}

	trie, err := db.trieReader(root)
	if err != nil {
		return err
	}
	trie.CopyToStore(trieStore(snapshot))
	return nil
}

func (db *storeDB) restoreSnapshot(root trie.Hash, snapshot kvstore.KVStore) error {
	blockKey := keyBlockByTrieRoot(root)
	blockBytes, err := snapshot.Get(blockKey)
	if err != nil {
		return err
	}
	db.mustSet(blockKey, blockBytes)

	trieSnapshot, err := trieReader(trieStore(snapshot), root)
	if err != nil {
		return err
	}
	trieSnapshot.CopyToStore(trieStore(db))

	db.setLatestTrieRoot(root)
	return nil
}
