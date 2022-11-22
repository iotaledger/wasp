// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"bytes"
	"errors"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/trie.go/hive_adaptor"
	"github.com/iotaledger/trie.go/immutable"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

const (
	prefixBlockByTrieRoot = iota
	prefixTrie
	prefixLatestTrieRoot

	// KeyChainID is the key used to store the chain ID in the state.
	// It should not collide with any hname prefix (which are 32 bits long).
	KeyChainID = kv.Key(rune(0))
)

var (
	ErrTrieRootNotFound      = errors.New("trie root not found")
	ErrUnknownLatestTrieRoot = errors.New("latest trie root is unknown")
)

// trieIdentity is the value assigned to the empty key in trie.go. It can be any non-empty value.
var trieIdentity = []byte{0}

func keyBlockByTrieRoot(root common.VCommitment) []byte {
	return append([]byte{prefixBlockByTrieRoot}, root.Bytes()...)
}

func keyLatestTrieRoot() []byte {
	return []byte{prefixLatestTrieRoot}
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

func (db *storeDB) latestTrieRoot() (common.VCommitment, error) {
	if !db.hasLatestTrieRoot() {
		return nil, ErrUnknownLatestTrieRoot
	}
	b := db.mustGet(keyLatestTrieRoot())
	ret := commitmentModel.NewVectorCommitment()
	err := ret.Read(bytes.NewReader(b))
	mustNoErr(err)
	return ret, nil
}

func (db *storeDB) hasLatestTrieRoot() bool {
	return db.mustHas(keyLatestTrieRoot())
}

func (db *storeDB) setLatestTrieRoot(root common.VCommitment) {
	db.mustSet(keyLatestTrieRoot(), root.Bytes())
}

func (db *storeDB) trieStore() common.KVStore {
	return hive_adaptor.NewHiveKVStoreAdaptor(db, []byte{prefixTrie})
}

func (db *storeDB) trieUpdatable(root common.VCommitment) (*immutable.TrieUpdatable, error) {
	return immutable.NewTrieUpdatable(commitmentModel, db.trieStore(), root)
}

func (db *storeDB) initTrie() common.VCommitment {
	return immutable.MustInitRoot(db.trieStore(), commitmentModel, trieIdentity)
}

func (db *storeDB) trieReader(root common.VCommitment) (*immutable.TrieReader, error) {
	return immutable.NewTrieReader(commitmentModel, db.trieStore(), root)
}

func (db *storeDB) hasBlock(root common.VCommitment) bool {
	return db.mustHas(keyBlockByTrieRoot(root))
}

func (db *storeDB) addBlock(block Block) {
	prev := block.PreviousL1Commitment()
	if prev != nil && !db.mustHas(keyBlockByTrieRoot(prev.GetTrieRoot())) {
		panic("cannot add block to store: previous block not found")
	}
	db.saveBlock(block)
}

func (db *storeDB) saveBlock(block Block) {
	db.mustSet(keyBlockByTrieRoot(block.TrieRoot()), block.Bytes())
}

func (db *storeDB) readBlock(root common.VCommitment) (*block, error) {
	key := keyBlockByTrieRoot(root)
	if !db.mustHas(key) {
		return nil, ErrTrieRootNotFound
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
