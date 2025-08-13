package trie

import "github.com/samber/lo"

//----------------------------------------------------------------------------
// generic abstraction interfaces of key/value storage

// KVReader is a key/value reader
type KVReader interface {
	// Get retrieves value by key. Returned nil means absence of the key
	Get(key []byte) []byte

	// MultiGet retrieves multiple values by keys.
	MultiGet([][]byte) [][]byte

	// Has checks presence of the key in the key/value store
	Has(key []byte) bool // for performance

	Iterate(prefix []byte, f func(k, v []byte) bool)

	IterateKeys(prefix []byte, f func(k []byte) bool)
}

// KVWriter is a key/value writer
type KVWriter interface {
	// Set writes new or updates existing key with the value.
	// value == nil means deletion of the key from the store
	Set(key, value []byte)
	Del(key []byte)
}

// KVStore is a compound interface
type KVStore interface {
	KVReader
	KVWriter
}

type kvStorePartition struct {
	prefix byte
	s      KVStore
}

func (p *kvStorePartition) Get(key []byte) []byte {
	return p.s.Get(concat([]byte{p.prefix}, key))
}

func (p *kvStorePartition) MultiGet(keys [][]byte) [][]byte {
	return p.s.MultiGet(lo.Map(keys, func(k []byte, _ int) []byte {
		return concat([]byte{p.prefix}, k)
	}))
}

func (p *kvStorePartition) Has(key []byte) bool {
	return p.s.Has(concat([]byte{p.prefix}, key))
}

func (p *kvStorePartition) Del(key []byte) {
	p.s.Del(concat([]byte{p.prefix}, key))
}

func (p *kvStorePartition) Set(key []byte, value []byte) {
	p.s.Set(concat([]byte{p.prefix}, key), value)
}

func (p *kvStorePartition) Iterate(prefix []byte, f func([]byte, []byte) bool) {
	p.s.Iterate(append([]byte{p.prefix}, prefix...), func(k, v []byte) bool {
		return f(k[1:], v)
	})
}

func (p *kvStorePartition) IterateKeys(prefix []byte, f func([]byte) bool) {
	p.s.IterateKeys(append([]byte{p.prefix}, prefix...), func(k []byte) bool {
		return f(k[1:])
	})
}

func makeKVStorePartition(s KVStore, prefix byte) *kvStorePartition {
	return &kvStorePartition{
		prefix: prefix,
		s:      s,
	}
}

type readerPartition struct {
	prefix byte
	r      KVReader
}

func (p *readerPartition) Get(key []byte) []byte {
	return p.r.Get(concat([]byte{p.prefix}, key))
}

func (p *readerPartition) MultiGet(keys [][]byte) [][]byte {
	return p.r.MultiGet(lo.Map(keys, func(k []byte, _ int) []byte {
		return concat([]byte{p.prefix}, k)
	}))
}

func (p *readerPartition) Has(key []byte) bool {
	return p.r.Has(concat([]byte{p.prefix}, key))
}

func (p *readerPartition) Iterate(prefix []byte, f func([]byte, []byte) bool) {
	p.r.Iterate(append([]byte{p.prefix}, prefix...), func(k, v []byte) bool {
		return f(k[1:], v)
	})
}

func (p *readerPartition) IterateKeys(prefix []byte, f func([]byte) bool) {
	p.r.IterateKeys(append([]byte{p.prefix}, prefix...), func(k []byte) bool {
		return f(k[1:])
	})
}

func makeReaderPartition(r KVReader, prefix byte) KVReader {
	return &readerPartition{
		prefix: prefix,
		r:      r,
	}
}

type writerPartition struct {
	prefix byte
	w      KVWriter
}

func (w *writerPartition) Set(key, value []byte) {
	w.w.Set(concat([]byte{w.prefix}, key), value)
}

func (w *writerPartition) Del(key []byte) {
	w.w.Del(concat([]byte{w.prefix}, key))
}

func makeWriterPartition(w KVWriter, prefix byte) KVWriter {
	return &writerPartition{
		prefix: prefix,
		w:      w,
	}
}
