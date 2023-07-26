package cache

import (
	"errors"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
)

const (
	partitionSize = 5
)

type partition [partitionSize]byte

type Stats struct {
	*fastcache.Stats

	// Number of handles
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
	// size for the cache
	// default value of 0 means disabled
	cacheSize int

	initOnce sync.Once

	// the handle counter
	handleCounter uint64

	// mutex
	mutex = &sync.Mutex{}

	// the fastcache
	cache *fastcache.Cache
)

// set Cache size
// called from cache component to set parameter
// before first use
func SetCacheSize(size int) error {
	mutex.Lock()
	defer mutex.Unlock()

	if size < 32*1024*1024 || size > 1024*1024*1024 {
		return errors.New("allowed size 32MiB to 1GiB")
	}
	cacheSize = size
	return nil
}

// get fastcache statistics
func GetStats() *Stats {
	mutex.Lock()
	defer mutex.Unlock()

	// cache disabled
	if cache == nil {
		return nil
	}

	stats := &fastcache.Stats{}
	cache.UpdateStats(stats)
	return &Stats{
		Stats:      stats,
		NumHandles: handleCounter,
	}
}

// create a new cache partition
// initializes the cache if not already happened
func NewCacheParition() (CacheInterface, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// initialize the cache first time it is used
	if cacheSize != 0 {
		initOnce.Do(func() {
			cache = fastcache.New(cacheSize)
		})
	}

	// if cache disabled or we used all handles
	// return a cache (as failsafe) that does nothing
	if cache == nil || handleCounter >= (1<<(partitionSize*8))-1 {
		return &CacheNoop{}, nil
	}

	// increment the handle counter
	handleCounter++

	// store counter into byte array by selecting 8bit-wise via
	// shift and cast and store each byte at the 'i'th position
	// of the partitionBytes array
	var partitionBytes partition
	for i := 0; i < partitionSize; i++ {
		partitionBytes[i] = byte(handleCounter >> (i * 8))
	}

	// and return it as CachePartition struct
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
