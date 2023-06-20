package kv

import (
	"github.com/VictoriaMetrics/fastcache"
)

type cachedKVStoreReader struct {
	KVStoreReader
	cache *fastcache.Cache
}

// NewCachedKVStoreReader wraps a KVStoreReader with an in-memory cache.
// IMPORTANT: there is no logic for cache invalidation, so make sure that the
// underlying KVStoreReader is never mutated.
func NewCachedKVStoreReader(r KVStoreReader, cacheSize int) KVStoreReader {
	cache := fastcache.New(cacheSize)
	return &cachedKVStoreReader{
		KVStoreReader: r,
		cache:         cache,
	}
}

func (c *cachedKVStoreReader) Get(key Key) []byte {
	if v := c.cache.Get(nil, []byte(key)); v != nil {
		return v
	}
	v := c.KVStoreReader.Get(key)
	c.cache.Set([]byte(key), v)
	return v
}

func (c *cachedKVStoreReader) Has(key Key) bool {
	return c.Get(key) != nil
}
