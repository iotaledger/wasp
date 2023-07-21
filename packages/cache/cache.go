package cache

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
)

type partition [4]byte

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
	handleCounter uint32
	mutex         = &sync.Mutex{}
	cache         *fastcache.Cache
)

func InitCache(size int) error {
	if size < 32*1024*1024 || size > 1024*1024*1024 {
		return errors.New("allowed size 32MiB to 1GiB")
	}
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

func NewCacheParition() (CacheInterface, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// if cache disabled return a /cache/ that does nothing
	if cache == nil {
		return &CacheNoop{}, nil
	}

	if handleCounter == math.MaxUint32 {
		return nil, errors.New("too many cache partitions")
	}
	handleCounter++

	var partitionBytes partition
	binary.LittleEndian.PutUint32(partitionBytes[:], handleCounter)

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
