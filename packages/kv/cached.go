package kv

import (
	"github.com/iotaledger/wasp/v2/packages/cache"
)

type cachedKVStoreReader struct {
	KVStoreReader
	cache cache.CacheInterface
}

// NewCachedKVStoreReader wraps a KVStoreReader with an in-memory cache.
// IMPORTANT: there is no logic for cache invalidation, so make sure that the
// underlying KVStoreReader is never mutated.
func NewCachedKVStoreReader(r KVStoreReader) KVStoreReader {
	cache, err := cache.NewCacheParition()
	if err != nil {
		panic(err)
	}
	return &cachedKVStoreReader{KVStoreReader: r, cache: cache}
}

func (c *cachedKVStoreReader) Get(key Key) []byte {
	if v, ok := c.cache.Get([]byte(key)); ok {
		return v
	}
	v := c.KVStoreReader.Get(key)
	c.cache.Add([]byte(key), v)
	return v
}

func (c *cachedKVStoreReader) Has(key Key) bool {
	return c.Get(key) != nil
}
