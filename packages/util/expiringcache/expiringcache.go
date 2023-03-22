package expiringcache

import (
	"runtime"
	"sync"
	"time"
)

type cacheItem[V any] struct {
	value  V
	expiry int64
}

type ExpiringCache[K comparable, V any] struct {
	*cache[K, V]
}

type cache[K comparable, V any] struct {
	items map[K]cacheItem[V]
	ttl   time.Duration
	mut   sync.RWMutex
}

func (c *cache[K, V]) cleanup() {
	c.mut.Lock()
	defer c.mut.Unlock()
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if now >= v.expiry {
			delete(c.items, k)
		}
	}
}

func (c *cache[K, V]) Set(k K, v V) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.items[k] = cacheItem[V]{
		expiry: time.Now().Add(c.ttl).UnixNano(),
		value:  v,
	}
}

func (c *cache[K, V]) Get(k K) interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()
	item, exists := c.items[k]
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
		items: make(map[K]cacheItem[V]),
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
