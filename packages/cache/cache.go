package cache

import (
	"errors"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
)

const partitionSize = 5

type partition [partitionSize]byte

type Stats struct {
	*fastcache.Stats
	NumHandles uint64
}

type CacheInterface interface {
	Get(key []byte) ([]byte, bool)
	Add(key []byte, value []byte)
}

// cache using the fastcache instance
type CachePartition struct {
	CacheInterface
	partition *partition
}

// cache doing nothing
type CacheNoop struct {
	CacheInterface
}

var (
	handleCounter uint64
	mutex         = &sync.Mutex{}
	cache         *fastcache.Cache
)

func InitCache(size int) error {
	mutex.Lock()
	defer mutex.Unlock()

	if size < 32*1024*1024 || size > 1024*1024*1024 {
		return errors.New("allowed size 32MiB to 1GiB")
	}
	cache = fastcache.New(size)
	if cache == nil {
		return errors.New("creating lru cache failed")
	}
	return nil
}

func GetStats() *Stats {
	// cache disabled
	if cache == nil {
		return nil
	}

	mutex.Lock()
	defer mutex.Unlock()

	stats := &fastcache.Stats{}
	cache.UpdateStats(stats)
	return &Stats{
		Stats:      stats,
		NumHandles: handleCounter,
	}
}

func NewCacheParition() (CacheInterface, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// if cache disabled or we used all handles
	// return a cache (as failsafe) that does nothing
	if cache == nil || handleCounter >= (1<<(partitionSize*8))-1 {
		return &CacheNoop{}, nil
	}

	handleCounter++

	var partitionBytes partition
	for i := 0; i < partitionSize; i++ {
		partitionBytes[i] = byte(handleCounter >> (i * 8))
	}

	return &CachePartition{
		partition: &partitionBytes,
	}, nil
}

func (c *CacheNoop) Get(key []byte) ([]byte, bool) {
	return nil, false
}

func (c *CacheNoop) Add(key []byte, value []byte) {
}

func (c *CachePartition) Get(key []byte) ([]byte, bool) {
	return cache.HasGet(nil, append(c.partition[:], key...))
}

func (c *CachePartition) Add(key []byte, value []byte) {
	cache.Set(append(c.partition[:], key...), value)
}
