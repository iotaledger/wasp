package main

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	"github.com/samber/lo"
)

// Defines a function signature for a single entry migration.
// NOTE: if SrcEntity and/or DestEntity types are []byte, they are not encoded/decoded and just passed as is.
type EntityMigrationFunc[SrcEntity any, DestEntity any] func(srcKey old_kv.Key, srcVal SrcEntity) (destKey kv.Key, _ DestEntity)

// Iterate by prefix and migrate each entry from old state to new state
func migrateEntitiesByPrefix[Src any, Dest any](srcContractState old_kv.KVStoreReader, destContractState kv.KVStore, oldPrefix string, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	count := uint32(0)

	srcContractState.Iterate(old_kv.Key(oldPrefix), func(srcKey old_kv.Key, srcBytes []byte) bool {
		migrateEntityState(srcContractState, destContractState, srcKey, migrationFunc)
		count++

		return true
	})

	return count
}

// Migrate entities from Go map into destination state
func migrateEntitiesByKV[Src any, Dest any](srcEntities map[old_kv.Key][]byte, destContractState kv.KVStore, migrationFunc EntityMigrationFunc[Src, Dest]) {
	for srcKey, srcBytes := range srcEntities {
		newK, newV := migrateEntityBytes(srcKey, srcBytes, migrationFunc)
		destContractState.Set(newK, newV)
	}
}

// Migrate entities from the named map int another named map
func migrateEntitiesMapByName[Src any, Dest any](srcContractState old_kv.KVStoreReader, destContractState kv.KVStore, oldMapName, newMapName string, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	if oldMapName == "" {
		panic("oldMapName is empty")
	}
	if newMapName == "" {
		panic("newMapName is empty")
	}

	srcEntities := old_collections.NewMapReadOnly(srcContractState, oldMapName)
	destEntities := collections.NewMap(destContractState, newMapName)

	migrateEntitiesMap(srcEntities, destEntities, migrationFunc)

	return srcEntities.Len()
}

// Migrate entities from state map into another state map
func migrateEntitiesMap[Src any, Dest any](srcEntities *old_collections.ImmutableMap, destEntities *collections.Map, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcEntities.Iterate(func(srcKey, srcBytes []byte) bool {
		newK, newV := migrateEntityBytes(old_kv.Key(srcKey), srcBytes, migrationFunc)
		destEntities.SetAt([]byte(newK), newV)

		return true
	})
}

// Migrate entities from the named array into another named array
func migrateEntitiesArrayByName[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, oldArrName, newArrName string, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	if oldArrName == "" {
		panic("oldArrName is empty")
	}
	if newArrName == "" {
		panic("newArrName is empty")
	}

	srcEntities := collections.NewArrayReadOnly(srcContractState, oldArrName)
	destEntities := collections.NewArray(destContractState, newArrName)

	migrateEntitiesArray(srcEntities, destEntities, migrationFunc)

	return srcEntities.Len()
}

// Migrate entities from state array into another state array
func migrateEntitiesArray[Src any, Dest any](srcEntities *collections.ArrayReadOnly, destEntities *collections.Array, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	for i := uint32(0); i < srcEntities.Len(); i++ {
		srcBytes := srcEntities.GetAt(i)
		_, newV := migrateEntityBytes("", srcBytes, migrationFunc)
		destEntities.SetAt(i, newV)
	}

	return srcEntities.Len()
}

// Migrate entity from old state to new state
func migrateEntityState[Src any, Dest any](srcContractState old_kv.KVStoreReader, destContractState kv.KVStore, srcKey old_kv.Key, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcBytes := srcContractState.Get(srcKey)

	newKey, newVal := migrateEntityBytes(srcKey, srcBytes, migrationFunc)

	destContractState.Set(newKey, newVal)
}

// Migrate entity from old bytes to new bytes
func migrateEntityBytes[Src any, Dest any](srcKey old_kv.Key, srcBytes []byte, migrationFunc EntityMigrationFunc[Src, Dest]) (newKey kv.Key, newVal []byte) {
	srcEntity := lo.Must(Deserialize[Src](srcBytes))
	destKey, destEntity := migrationFunc(srcKey, srcEntity)
	return destKey, Serialize(destEntity)
}

// Old bytes are copied into new state
func migrateAsIs(newKey kv.Key) EntityMigrationFunc[[]byte, []byte] {
	if newKey == "" {
		panic("newKey cannot be empty")
	}

	return func(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
		return newKey, srcVal
	}
}

// Old bytes are just decoded and re-encoded again as new bytes
func migrateEncoding[ValueType any](newKey kv.Key) EntityMigrationFunc[ValueType, ValueType] {
	if newKey == "" {
		panic("newKey cannot be empty")
	}

	return func(srcKey old_kv.Key, srcVal ValueType) (destKey kv.Key, destVal ValueType) {
		if _, ok := interface{}(srcVal).([]byte); ok {
			panic("srcVal cannot be []byte - use migrateAsIs instead")
		}

		return newKey, srcVal
	}
}

func IterateByPrefix[Src any](srcContractState old_kv.KVStoreReader, prefix string, f func(k old_kv.Key, v Src)) uint32 {
	var count uint32

	srcContractState.Iterate(old_kv.Key(prefix), func(k old_kv.Key, b []byte) bool {
		count++
		v := lo.Must(Deserialize[Src](b))
		f(k, v)
		return true
	})

	return count
}

// Iterate by prefix and return all matchign entries
func ListByPrefix(store kv.KVStoreReader, prefix string) map[kv.Key][]byte {
	entries := map[kv.Key][]byte{}

	store.Iterate(kv.Key(prefix), func(key kv.Key, val []byte) bool {
		entries[key] = val
		return true
	})

	return entries
}

// Split map key into map name and element key
func SplitMapKey(storeKey old_kv.Key) (mapName, elemKey old_kv.Key) {
	const elemSep = "."
	pos := strings.Index(string(storeKey), elemSep)

	isMap := pos >= 0 && pos < len(storeKey)-1
	if !isMap {
		return storeKey, ""
	}

	return storeKey[:pos], storeKey[pos+1:]
}

// Wraps a function and adds to it printing of number of times it was called
func p[Src any, Dest any](f EntityMigrationFunc[Src, Dest]) EntityMigrationFunc[Src, Dest] {
	callCount := 0

	return func(oldKey old_kv.Key, srcVal Src) (kv.Key, Dest) {
		callCount++
		if callCount%100 == 0 {
			fmt.Printf("\rProcessed: %v         ", callCount)
		}

		return f(oldKey, srcVal)
	}
}
