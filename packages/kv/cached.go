package kv

import lru "github.com/hashicorp/golang-lru/v2"

type cachedKVStoreReader struct {
	KVStoreReader
	cache *lru.Cache[Key, []byte]
}

// NewCachedKVStoreReader wraps a KVStoreReader with an in-memory cache.
// IMPORTANT: there is no logic for cache invalidation, so make sure that the
// underlying KVStoreReader is never mutated.
func NewCachedKVStoreReader(r KVStoreReader, cacheSize int) KVStoreReader {
	cache, err := lru.New[Key, []byte](cacheSize)
	if err != nil {
		panic(err)
	}
	return &cachedKVStoreReader{
		KVStoreReader: r,
		cache:         cache,
	}
}

func (c *cachedKVStoreReader) Get(key Key) []byte {
	if v, ok := c.cache.Get(key); ok {
		return v
	}
	v := c.KVStoreReader.Get(key)
	c.cache.Add(key, v)
	return v
}

func (c *cachedKVStoreReader) Has(key Key) bool {
	return c.Get(key) != nil
}
