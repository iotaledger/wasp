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
// NOTE: if SrcRecord and/or DestRecord types are []byte, they are not encoded/decoded and just passed as is.
type RecordMigrationFunc[SrcKey, SrcRecord, DestKey, DestRecord any] func(srcKey SrcKey, srcVal SrcRecord) (destKey DestKey, _ DestRecord)

// Iterate by prefix and migrate each entry from old state to new state
func migrateEntitiesByPrefix[SrcK, SrcV, DestK, DestV any](srcContractState old_kv.KVStoreReader, destContractState kv.KVStore, oldPrefix string, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) uint32 {
	count := uint32(0)

	srcContractState.Iterate(old_kv.Key(oldPrefix), func(srcKey old_kv.Key, srcBytes []byte) bool {
		migrateRecord(srcContractState, destContractState, srcKey, migrationFunc)
		count++

		return true
	})

	return count
}

// Migrate records from Go map into destination state
func migrateEntitiesByKV[SrcK, SrcV, DestK, DestV any](srcEntities map[old_kv.Key][]byte, destContractState kv.KVStore, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) {
	for srcKey, srcBytes := range srcEntities {
		newK, newV := migrateRecordBytes(srcKey, srcBytes, migrationFunc)
		destContractState.Set(newK, newV)
	}
}

// Migrate records from the named map int another named map
func migrateEntitiesMapByName[SrcK, SrcV, DestK, DestV any](srcContractState old_kv.KVStoreReader, destContractState kv.KVStore, oldMapName, newMapName string, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) uint32 {
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

// Migrate records from state map into another state map
func migrateEntitiesMap[SrcK, SrcV, DestK, DestV any](srcEntities *old_collections.ImmutableMap, destEntities *collections.Map, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) {
	srcEntities.Iterate(func(srcKey, srcBytes []byte) bool {
		newK, newV := migrateRecordBytes(old_kv.Key(srcKey), srcBytes, migrationFunc)
		destEntities.SetAt([]byte(newK), newV)

		return true
	})
}

// Migrate records from the named array into another named array
func migrateEntitiesArrayByName[SrcK, SrcV, DestK, DestV any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, oldArrName, newArrName string, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) uint32 {
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

// Migrate records from state array into another state array
func migrateEntitiesArray[SrcK, SrcV, DestK, DestV any](srcEntities *collections.ArrayReadOnly, destEntities *collections.Array, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) uint32 {
	for i := uint32(0); i < srcEntities.Len(); i++ {
		srcBytes := srcEntities.GetAt(i)
		_, newV := migrateRecordBytes("", srcBytes, migrationFunc)
		destEntities.SetAt(i, newV)
	}

	return srcEntities.Len()
}

// Migrate record from old state to new state
func migrateRecord[SrcK, SrcV, DestK, DestV any](srcContractState old_kv.KVStoreReader, destContractState kv.KVStore, srcKey old_kv.Key, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) {
	srcBytes := srcContractState.Get(srcKey)

	newKey, newVal := migrateRecordBytes(srcKey, srcBytes, migrationFunc)

	destContractState.Set(newKey, newVal)
}

// Migrate record from old bytes to new bytes
func migrateRecordBytes[SrcK, SrcV, DestK, DestV any](srcKeyBytes old_kv.Key, srcValueBytes []byte, migrationFunc RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) (newKeyBytes kv.Key, newValBytes []byte) {
	srcKey := lo.Must(Deserialize[SrcK]([]byte(srcKeyBytes)))
	srcValue := lo.Must(Deserialize[SrcV](srcValueBytes))

	destKey, destRecord := migrationFunc(srcKey, srcValue)

	return kv.Key(Serialize(destKey)), Serialize(destRecord)
}

// Old bytes are copied into new state
func copyBytes(newKey kv.Key) RecordMigrationFunc[old_kv.Key, []byte, kv.Key, []byte] {
	if newKey == "" {
		panic("newKey cannot be empty")
	}

	return func(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
		return newKey, srcVal
	}
}

// Old bytes are just decoded and re-encoded again as new bytes
func asIs[Value any](newKey kv.Key) RecordMigrationFunc[old_kv.Key, Value, kv.Key, Value] {
	if newKey == "" {
		panic("newKey cannot be empty")
	}

	return func(srcKey old_kv.Key, srcVal Value) (destKey kv.Key, destVal Value) {
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
func p[SrcK, SrcV, DestK, DestV any](f RecordMigrationFunc[SrcK, SrcV, DestK, DestV]) RecordMigrationFunc[SrcK, SrcV, DestK, DestV] {
	callCount := 0

	return func(oldKey SrcK, srcVal SrcV) (DestK, DestV) {
		callCount++
		if callCount%100 == 0 {
			fmt.Printf("\rProcessed: %v         ", callCount)
		}

		return f(oldKey, srcVal)
	}
}
