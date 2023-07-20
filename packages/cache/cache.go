package cache

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
)

type partition [4]byte

type Cache struct {
	handleCounter uint32
	mutex         *sync.Mutex
	cache         *fastcache.Cache
}

var sharedCache *Cache

func init() {
	var err error
	sharedCache, err = newCache(512 * 1024 * 1024)
	if err != nil {
		panic("creating lru cache failed")
	}
}

func newCache(size int) (*Cache, error) {
	cache := fastcache.New(size)
	if cache == nil {
		return nil, errors.New("creating lru cache failed")
	}
	return &Cache{
		handleCounter: 0,
		mutex:         &sync.Mutex{},
		cache:         cache,
	}, nil
}

func handleToPartition(handle uint32) *partition {
	var partition partition
	binary.LittleEndian.PutUint32(partition[:], handle)
	return &partition
}

func NewCacheParition() (*CachePartition, error) {
	sharedCache.mutex.Lock()
	defer sharedCache.mutex.Unlock()

	if sharedCache.handleCounter == math.MaxUint32 {
		return nil, errors.New("too many cache partitions")
	}
	sharedCache.handleCounter++

	return &CachePartition{
		cache:     sharedCache,
		partition: handleToPartition(sharedCache.handleCounter),
	}, nil
}

func (c *Cache) Get(partition *partition, key string) ([]byte, bool) {
	v := c.cache.Get(nil, append(partition[:], []byte(key)...))
	return v, v != nil
}

func (c *Cache) Add(partition *partition, key string, value []byte) bool {
	c.cache.Set(append(partition[:], []byte(key)...), value)
	return false
}

type CachePartition struct {
	cache     *Cache
	partition *partition
}

func (c *CachePartition) Get(key string) ([]byte, bool) {
	return c.cache.Get(c.partition, key)
}

func (c *CachePartition) Add(key string, value []byte) bool {
	return c.cache.Add(c.partition, key, value)
}
