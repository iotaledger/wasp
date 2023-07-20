package trie

import "github.com/iotaledger/wasp/packages/cache"

type cachedKVReader struct {
	r     KVReader
	cache *cache.CachePartition
}

func makeCachedKVReader(r KVReader) KVReader {
	cache, err := cache.NewCacheParition()
	if err != nil {
		panic(err)
	}
	return &cachedKVReader{r: r, cache: cache}
}

func (c *cachedKVReader) Get(key []byte) []byte {
	if v, ok := c.cache.Get(string(key)); ok {
		return v
	}
	v := c.r.Get(key)
	c.cache.Add(string(key), v)
	return v
}

func (c *cachedKVReader) Has(key []byte) bool {
	if v, ok := c.cache.Get(string(key)); ok {
		return v != nil
	}
	v := c.r.Get(key)
	c.cache.Add(string(key), v)
	return v != nil
}
