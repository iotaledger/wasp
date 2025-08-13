package trie

import (
	"github.com/iotaledger/wasp/v2/packages/cache"
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
	missing := make([][]byte, 0, len(keys))
	for _, key := range keys {
		_, ok := c.cache.Get(key)
		if !ok {
			missing = append(missing, key)
		}
	}
	var missingValues [][]byte
	if len(missing) > 0 {
		missingValues = c.r.MultiGet(missing)
	}
	values := make([][]byte, len(keys))
	for i, key := range keys {
		v, ok := c.cache.Get(key)
		if ok {
			values[i] = v
		} else {
			values[i] = missingValues[0]
			missingValues = missingValues[1:]
			c.cache.Add(key, values[i])
		}
	}
	return values
}

func (c *cachedKVReader) Has(key []byte) bool {
	return c.Get(key) != nil
}
