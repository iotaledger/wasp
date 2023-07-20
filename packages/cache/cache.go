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

func init() {
	cache = fastcache.New(512 * 1024 * 1024)
	if cache == nil {
		panic("creating lru cache failed")
	}
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
	v := cache.Get(nil, append(c.partition[:], []byte(key)...))
	return v, v != nil
}

func (c *CachePartition) Add(key []byte, value []byte) bool {
	cache.Set(append(c.partition[:], []byte(key)...), value)
	return false
}
