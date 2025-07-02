package trie

import (
	"github.com/iotaledger/wasp/packages/cache"
)

type cachedKVReader struct {
	r     KVReader
	cache cache.CacheInterface
}

func makeCachedKVReader(r KVReader) KVReader {
	cache, err := cache.NewCacheParition()
	if err != nil {
		panic(err)
	}
	return &cachedKVReader{r: r, cache: cache}
}

func (c *cachedKVReader) Get(key []byte) []byte {
	if v, ok := c.cache.Get(key); ok {
		return v
	}
	v := c.r.Get(key)
	c.cache.Add(key, v)
	return v
}

func (c *cachedKVReader) MultiGet(keys [][]byte) [][]byte {
	// TODO: use cache
	v := c.r.MultiGet(keys)
	return v
}

func (c *cachedKVReader) Has(key []byte) bool {
	return c.Get(key) != nil
}
