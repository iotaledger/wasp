// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/iotaledger/wasp/packages/kv"
)

// KVAdapter implements ethdb.KeyValueStore with a kv.KVStore as backend
type KVAdapter struct {
	kv kv.KVStore
}

var _ ethdb.KeyValueStore = &KVAdapter{}

func NewKVAdapter(store kv.KVStore) *KVAdapter {
	return &KVAdapter{store}
}

// Has retrieves if a key is present in the key-value data store.
func (k *KVAdapter) Has(key []byte) (bool, error) {
	return k.kv.Has(kv.Key(key))
}

// Get retrieves the given key if it's present in the key-value data store.
func (k *KVAdapter) Get(key []byte) ([]byte, error) {
	return k.kv.Get(kv.Key(key))
}

// Put inserts the given value into the key-value data store.
func (k *KVAdapter) Put(key, value []byte) error {
	if value == nil {
		value = []byte{}
	}
	k.kv.Set(kv.Key(key), value)
	return nil
}

// Delete removes the key from the key-value data store.
func (k *KVAdapter) Delete(key []byte) error {
	k.kv.Del(kv.Key(key))
	return nil
}

// NewBatch creates a write-only database that buffers changes to its host db
// until a final write is called.
func (k *KVAdapter) NewBatch() ethdb.Batch {
	return &batch{k: k}
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
//
// Note: This method assumes that the prefix is NOT part of the start, so there's
// no need for the caller to prepend the prefix to the start
func (k *KVAdapter) NewIterator(prefix, start []byte) ethdb.Iterator {
	it := &iterator{
		next:  make(chan struct{}),
		items: make(chan keyvalue, 1),
		errCh: make(chan error, 1),
	}
	inited := false
	go func() {
		_, ok := <-it.next
		if !ok {
			return
		}
		it.errCh <- k.kv.IterateSorted(kv.Key(prefix), func(key kv.Key, value []byte) bool {
			if start != nil && !inited && key < kv.Key(prefix)+kv.Key(start) {
				return true
			}
			inited = true
			it.items <- keyvalue{key: []byte(key), value: value}
			_, ok := <-it.next
			return ok
		})
		close(it.items)
	}()
	return it
}

// Stat returns a particular internal stat of the database.
func (k *KVAdapter) Stat(property string) (string, error) {
	panic("not implemented") // TODO: Implement
}

// Compact flattens the underlying data store for the given key range. In essence,
// deleted and overwritten versions are discarded, and the data is rearranged to
// reduce the cost of operations needed to access them.
//
// A nil start is treated as a key before all keys in the data store; a nil limit
// is treated as a key after all keys in the data store. If both is nil then it
// will compact entire data store.
func (k *KVAdapter) Compact(start, limit []byte) error {
	return nil
}

func (k *KVAdapter) Close() error {
	return nil
}

// keyvalue is a key-value tuple tagged with a deletion field to allow creating
// memory-database write batches.
type keyvalue struct {
	key    []byte
	value  []byte
	delete bool
}

// batch is a write-only memory batch that commits changes to its host
// database when Write is called. A batch cannot be used concurrently.
type batch struct {
	k      *KVAdapter
	writes []keyvalue
	size   int
}

// Put inserts the given value into the batch for later committing.
func (b *batch) Put(key, value []byte) error {
	b.writes = append(b.writes, keyvalue{common.CopyBytes(key), common.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

// Delete inserts the a key removal into the batch for later committing.
func (b *batch) Delete(key []byte) error {
	b.writes = append(b.writes, keyvalue{common.CopyBytes(key), nil, true})
	b.size += len(key)
	return nil
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *batch) ValueSize() int {
	return b.size
}

// Write flushes any accumulated data to the memory database.
func (b *batch) Write() error {
	for _, keyvalue := range b.writes {
		var err error
		if keyvalue.delete {
			err = b.k.Delete(keyvalue.key)
		} else {
			err = b.k.Put(keyvalue.key, keyvalue.value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Reset resets the batch for reuse.
func (b *batch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}

// Replay replays the batch contents.
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	for _, keyvalue := range b.writes {
		if keyvalue.delete {
			if err := w.Delete(keyvalue.key); err != nil {
				return err
			}
			continue
		}
		if err := w.Put(keyvalue.key, keyvalue.value); err != nil {
			return err
		}
	}
	return nil
}

// iterator implements ethdb.Iterator
type iterator struct {
	next  chan struct{}
	items chan keyvalue
	errCh chan error

	current keyvalue
	err     error
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted.
func (it *iterator) Next() (ok bool) {
	it.next <- struct{}{}
	select {
	case it.err = <-it.errCh:
		return false
	case it.current, ok = <-it.items:
	}
	return
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error. A memory iterator cannot encounter errors.
func (it *iterator) Error() error {
	return it.err
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (it *iterator) Key() []byte {
	return it.current.key
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (it *iterator) Value() []byte {
	return it.current.value
}

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (it *iterator) Release() {
	close(it.next)
}
