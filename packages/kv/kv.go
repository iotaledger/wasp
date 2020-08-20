package kv

// Since map cannot have []byte as key, to avoid unnecessary conversions
// between string and []byte, we use string as key data type, but it does
// not necessarily have to be a valid UTF-8 string.
type Key string

const EmptyPrefix = Key("")

func (k Key) HasPrefix(prefix Key) bool {
	if len(prefix) > len(k) {
		return false
	}
	return k[:len(prefix)] == prefix
}

// KVStore represents a key-value store where both keys and values are
// arbitrary byte slices.
type KVStore interface {
	Set(key Key, value []byte)
	Del(key Key)
	// Get returns the value, or nil if not found
	Get(key Key) ([]byte, error)
	Has(key Key) (bool, error)
	Iterate(prefix Key, f func(key Key, value []byte) bool) error
	IterateKeys(prefix Key, f func(key Key) bool) error

	// TODO add DelPrefix(prefix []byte)
	// deletes all keys with the prefix. Currently we don't have a possibility to iterate over keys
	// and maybe we do not need one in the sandbox. However we need a possibility to efficiently clear arrays,
	// dictionaries and timestamped logs
}
