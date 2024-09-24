package bcs

import (
	"reflect"
	"sync/atomic"
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

func newGlobalTypeInfoCache() *globalTypeInfoCache {
	var c globalTypeInfoCache
	c.cache.Store(&map[reflect.Type]typeInfo{})

	return &c
}

type globalTypeInfoCache struct {
	cache atomic.Pointer[map[reflect.Type]typeInfo]
}

func (c *globalTypeInfoCache) Get() localTypeInfoCache {
	return localTypeInfoCache{
		global:                c,
		existingTypeInfoCache: *c.cache.Load(),
		newTypeInfoCache:      make(map[reflect.Type]typeInfo),
	}
}

type localTypeInfoCache struct {
	global                *globalTypeInfoCache
	existingTypeInfoCache map[reflect.Type]typeInfo
	newTypeInfoCache      map[reflect.Type]typeInfo
}

func (c *localTypeInfoCache) Get(t reflect.Type) (typeInfo, bool) {
	if cached, isCached := c.existingTypeInfoCache[t]; isCached {
		return cached, true
	}
	if cached, isCached := c.newTypeInfoCache[t]; isCached {
		return cached, true
	}

	return typeInfo{}, false
}

func (c *localTypeInfoCache) Add(t reflect.Type, ti typeInfo) {
	c.newTypeInfoCache[t] = ti
}

func (c *localTypeInfoCache) Save() {
	if len(c.newTypeInfoCache) == 0 {
		return
	}

	for k, v := range c.existingTypeInfoCache {
		c.newTypeInfoCache[k] = v
	}

	c.global.cache.Store(&c.newTypeInfoCache)
}
