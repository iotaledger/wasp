package expiringcache

import (
	"runtime"
	"sync"
	"time"
)

type cacheItem struct {
	value  interface{}
	expiry int64
}

type ExpiringCache struct {
	*cache
}

type cache struct {
	items map[interface{}]cacheItem
	ttl   time.Duration
	mut   sync.RWMutex
}

func (c *cache) cleanup() {
	c.mut.Lock()
	defer c.mut.Unlock()
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if now >= v.expiry {
			delete(c.items, k)
		}
	}
}

func (c *cache) Set(k, v interface{}) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.items[k] = cacheItem{
		expiry: time.Now().Add(c.ttl).UnixNano(),
		value:  v,
	}
}

func (c *cache) Get(k interface{}) interface{} {
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

func New(ttl time.Duration, cleanupInterval ...time.Duration) *ExpiringCache {
	var cleanint time.Duration
	if len(cleanupInterval) == 1 {
		cleanint = cleanupInterval[0]
	} else {
		cleanint = defaultCleanupInterval
	}

	c := &cache{
		items: make(map[interface{}]cacheItem),
	}

	ret := &ExpiringCache{c} // prevent the cleanup loop from blocking garbage collection on `c`

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

	runtime.SetFinalizer(ret, func(_ *ExpiringCache) {
		// when the exported `*ExpiringCache` is GC'd, we stop the cleanup loop and allow `c` to be GC'd
		close(stopCleanup)
	})

	return ret
}
