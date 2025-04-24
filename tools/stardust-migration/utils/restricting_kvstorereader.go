package utils

import old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"

func OnlyIterator(r old_kv.KVStoreReader) old_kv.KVStoreReader {
	return &KVStoreReaderOnlyIterator{
		KVStoreReader: r,
	}
}

type KVStoreReaderOnlyIterator struct {
	old_kv.KVStoreReader
}

func (*KVStoreReaderOnlyIterator) Get(key old_kv.Key) []byte {
	panic("Get() is forbidden for this reader")
}

func (*KVStoreReaderOnlyIterator) Has(key old_kv.Key) bool {
	panic("Has() is forbidden for this reader")
}

func OnlyReader(r old_kv.KVStoreReader) old_kv.KVStoreReader {
	return &KVStoreReaderOnlyReader{
		KVStoreReader: r,
	}
}

type KVStoreReaderOnlyReader struct {
	old_kv.KVStoreReader
}

func (*KVStoreReaderOnlyReader) Iterate(prefix old_kv.Key, f func(key old_kv.Key, value []byte) bool) {
	panic("Iterate() is forbidden for this reader")
}
func (*KVStoreReaderOnlyReader) IterateKeys(prefix old_kv.Key, f func(key old_kv.Key) bool) {
	panic("IterateKeys() is forbidden for this reader")
}
func (*KVStoreReaderOnlyReader) IterateSorted(prefix old_kv.Key, f func(key old_kv.Key, value []byte) bool) {
	panic("IterateSorted() is forbidden for this reader")
}
func (*KVStoreReaderOnlyReader) IterateKeysSorted(prefix old_kv.Key, f func(key old_kv.Key) bool) {
	panic("IterateKeysSorted() is forbidden for this reader")
}
