package bcs

import (
	"reflect"
	"sync/atomic"

	"github.com/samber/lo"
)

// Upon serialization the types are checked for the presense of customizations. This make take significant time.
// To improve that, the type information is stored in cache.
// To avoid global mutex locks when reading/writing cache, atomic swapping is used instead.
// It works like that:
// * When coder is created, it get pointer to the current version of cache.
// * Current version of cache is only read, never written, so multiple encoders can read it at the same time.
// * When coder is done with coding, it creates new cache with updated information.
// * Then it atomically swaps the pointer to the new cache with the pointer to the current cache.
// * Multiple coders may update cache, thus overwritting modifications of each other. But it is not a problem, because
//   the info is only extended by them, so eventually cache will have information about all types.

func newSharedTypeInfoCache() *sharedTypeInfoCache {
	var c sharedTypeInfoCache
	c.entries.Store(&map[reflect.Type]typeInfo{})

	return &c
}

type sharedTypeInfoCache struct {
	entries atomic.Pointer[map[reflect.Type]typeInfo]
}

func (c *sharedTypeInfoCache) Get() localTypeInfoCache {
	return newLocalTypeInfoCache(c)
}

func newLocalTypeInfoCache(shared *sharedTypeInfoCache) localTypeInfoCache {
	return localTypeInfoCache{
		sharedCache:      shared,
		prevCacheEntries: *shared.entries.Load(),
		newCacheEntries:  make(map[reflect.Type]typeInfo),
	}
}

type localTypeInfoCache struct {
	sharedCache      *sharedTypeInfoCache
	prevCacheEntries map[reflect.Type]typeInfo
	newCacheEntries  map[reflect.Type]typeInfo
}

func (c *localTypeInfoCache) Get(t reflect.Type) (typeInfo, bool) {
	if cached, isCached := c.prevCacheEntries[t]; isCached {
		return cached, true
	}
	if cached, isCached := c.newCacheEntries[t]; isCached {
		return cached, true
	}

	return typeInfo{}, false
}

func (c *localTypeInfoCache) Add(t reflect.Type, ti typeInfo) {
	c.newCacheEntries[t] = ti
}

func (c *localTypeInfoCache) Save() {
	if len(c.newCacheEntries) == 0 {
		return
	}

	// Refreshing shared version of cache in case it was extended by somebody else while we were working on our local copy.
	// This is not mandatory, but may be useful in cases e.g. when two coders are used in parallel on two independant sets of types.
	// In that case without this line they would overwrite each others cache entries on every save.
	// Still, even with this line there is a teeny-tiny chance of that happening, but on a long run its not a problem.
	c.prevCacheEntries = *c.sharedCache.entries.Load()

	for k, v := range c.prevCacheEntries {
		c.newCacheEntries[k] = v
	}

	c.sharedCache.entries.Store(lo.ToPtr(c.newCacheEntries)) // NOTE: This is SUPER imporatant to use lo.ToPtr instead of &. Otherwise the pointer to field will be taken, which will cause data races.

	// This local cache may be reused (e.g. multiple calls to Encode for one Encoder). But we cannot continue
	// writing to c.newCacheEntries, because it is now shared with other coders, others may read from it.
	// But we can safely continue using it for reading.
	c.prevCacheEntries = c.newCacheEntries
	c.newCacheEntries = make(map[reflect.Type]typeInfo)
}
