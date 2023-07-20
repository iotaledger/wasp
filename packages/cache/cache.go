package cache

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
)

type partition [4]byte

type CachePartition struct {
	partition *partition
}

var (
	handleCounter uint32
	mutex         = &sync.Mutex{}
	cache         *fastcache.Cache
)

func InitCache(size int) error {
	cache = fastcache.New(size)
	if cache == nil {
		return errors.New("creating lru cache failed")
	}
	return nil
}

func GetStats() *fastcache.Stats {
	// cache disabled
	if cache == nil {
		return nil
	}

	stats := &fastcache.Stats{}
	cache.UpdateStats(stats)
	return stats
}

func NewCacheParition() (*CachePartition, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if handleCounter == math.MaxUint32 {
		return nil, errors.New("too many cache partitions")
	}
	handleCounter++

	var partition partition
	binary.LittleEndian.PutUint32(partition[:], handleCounter)

	return &CachePartition{
		partition: &partition,
	}, nil
}

func (c *CachePartition) Get(key []byte) ([]byte, bool) {
	// cache disabled
	if cache == nil {
		return nil, false
	}
	v, ok := cache.HasGet(nil, append(c.partition[:], []byte(key)...))
	return v, ok
}

func (c *CachePartition) Add(key []byte, value []byte) {
	// cache disabled
	if cache == nil {
		return
	}
	cache.Set(append(c.partition[:], []byte(key)...), value)
}
