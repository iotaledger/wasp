package trie

// generic abstraction interfaces of the underlying key/value storage

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
