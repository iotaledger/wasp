package kv

// Since map cannot have []byte as key, to avoid unnecessary conversions
// between string and []byte, we use string as key data type, but it does
// not necessarily have to be a valid UTF-8 string.
type Key string

// KVStore represents a key-value store where both keys and values are
// arbitrary byte slices.
type KVStore interface {
	Set(key Key, value []byte)
	Del(key Key)
	// Get returns the value, or nil if not found
	Get(key Key) ([]byte, error)
}
