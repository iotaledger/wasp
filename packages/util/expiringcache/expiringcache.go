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
	items map[interface{}]cacheItem
	ttl   time.Duration
	mut   sync.RWMutex
}

func (c *ExpiringCache) cleanup() {
	c.mut.Lock()
	defer c.mut.Unlock()
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if now >= v.expiry {
			delete(c.items, k)
		}
	}
}

func (c *ExpiringCache) Set(k, v interface{}) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.items[k] = cacheItem{
		expiry: time.Now().Add(c.ttl).UnixNano(),
		value:  v,
	}
}

func (c *ExpiringCache) Get(k interface{}) interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()
	item, exists := c.items[k]
	if !exists {
		return nil
	}
	// if needed could check expiration here, and return (nil, false) if item is already expired
	return item.value
}

func New(ttl time.Duration) *ExpiringCache {
	cache := &ExpiringCache{
		items: make(map[interface{}]cacheItem),
	}

	stopCleanup := make(chan bool)
	ticker := time.NewTicker(ttl)
	go func() {
		for {
			<-ticker.C
			cache.cleanup()
			<-stopCleanup
			ticker.Stop()
			return
		}
	}()

	runtime.SetFinalizer(cache, func() {
		close(stopCleanup)
	})

	return cache
}
