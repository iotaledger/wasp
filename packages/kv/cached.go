package kv

import lru "github.com/hashicorp/golang-lru"

type cachedKVStoreReader struct {
	KVStoreReader
	cache *lru.Cache
}

// NewCachedKVStoreReader wraps a KVStoreReader with an in-memory cache.
// IMPORTANT: there is no logic for cache invalidation, so make sure that the
// underlying KVStoreReader is never mutated.
func NewCachedKVStoreReader(r KVStoreReader, cacheSize int) KVStoreReader {
	cache, err := lru.New(cacheSize)
	if err != nil {
		panic(err)
	}
	return &cachedKVStoreReader{
		KVStoreReader: r,
		cache:         cache,
	}
}

func (c *cachedKVStoreReader) Get(key Key) ([]byte, error) {
	if v, ok := c.cache.Get(key); ok {
		return v.([]byte), nil
	}
	v, err := c.KVStoreReader.Get(key)
	if err != nil {
		return nil, err
	}
	c.cache.Add(key, v)
	return v, nil
}

func (c *cachedKVStoreReader) Has(key Key) (bool, error) {
	v, err := c.Get(key)
	if err != nil {
		return false, err
	}
	return v != nil, nil
}

func (c *cachedKVStoreReader) MustGet(key Key) []byte {
	return MustGet(c, key)
}

func (c *cachedKVStoreReader) MustHas(key Key) bool {
	return MustHas(c, key)
}
