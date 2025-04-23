// Package expiringcache implements a cache with automatic expiration of entries.
package expiringcache

import (
	"runtime"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

type cacheItem[V any] struct {
	value  V
	expiry int64
}

type ExpiringCache[K comparable, V any] struct {
	*cache[K, V]
}

type cache[K comparable, V any] struct {
	items *shrinkingmap.ShrinkingMap[K, cacheItem[V]]
	ttl   time.Duration
	mut   sync.RWMutex
}

func (c *cache[K, V]) cleanup() {
	c.mut.Lock()
	defer c.mut.Unlock()
	now := time.Now().UnixNano()
	c.items.ForEach(func(k K, v cacheItem[V]) bool {
		if now >= v.expiry {
			c.items.Delete(k)
		}
		return true
	})
}

func (c *cache[K, V]) Set(k K, v V) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.items.Set(k, cacheItem[V]{
		expiry: time.Now().Add(c.ttl).UnixNano(),
		value:  v,
	})
}

func (c *cache[K, V]) Get(k K) interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()
	item, exists := c.items.Get(k)
	if !exists {
		return nil
	}
	// if needed could check expiration here, and return (nil, false) if item is already expired
	return item.value
}

const defaultCleanupInterval = 60 * time.Second

func New[K comparable, V any](ttl time.Duration, cleanupInterval ...time.Duration) *ExpiringCache[K, V] {
	var cleanint time.Duration
	if len(cleanupInterval) == 1 {
		cleanint = cleanupInterval[0]
	} else {
		cleanint = defaultCleanupInterval
	}

	c := &cache[K, V]{
		items: shrinkingmap.New[K, cacheItem[V]](),
		ttl:   ttl,
	}

	ret := &ExpiringCache[K, V]{c} // prevent the cleanup loop from blocking garbage collection on `c`

	stopCleanup := make(chan bool)
	ticker := time.NewTicker(cleanint)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-stopCleanup:
				ticker.Stop()
				return
			}
		}
	}()

	runtime.SetFinalizer(ret, func(_ *ExpiringCache[K, V]) {
		// when the exported `*ExpiringCache` is GC'd, we stop the cleanup loop and allow `c` to be GC'd
		close(stopCleanup)
	})

	return ret
}
