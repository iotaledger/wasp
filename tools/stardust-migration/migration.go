package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/nnikolash/wasp-types-exported/packages/kv"
	"github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	"github.com/nnikolash/wasp-types-exported/packages/kv/collections"
)

type SerializableEntity interface {
	Bytes() []byte
}

type DeserializableEntity interface {
	Read(src io.Reader) error
}

func SerializeEntity(entity any) []byte {
	if serializable, isSerializible := entity.(SerializableEntity); isSerializible {
		return serializable.Bytes()
	} else {
		var ep any = &entity
		if serializable, isSerializible := ep.(SerializableEntity); isSerializible {
			return serializable.Bytes()
		}
	}

	return codec.Encode(entity)
}

func DeserializeEntity[Dest any](b []byte) (Dest, error) {
	var v Dest
	vp := &v

	f := func(de DeserializableEntity) (Dest, error) {
		r := bytes.NewReader(b)
		must(de.Read(r))

		if r.Len() != 0 {
			leftovers := must2(io.ReadAll(r))
			panic(fmt.Sprintf("Leftover bytes after reading entity of type %T: initialValue = %x, leftover = %x, leftoverLen = %v", v, b, leftovers, r.Len()))
		}

		return v, nil
	}

	if deserializable, isDeserializible := interface{}(v).(DeserializableEntity); isDeserializible {
		return f(deserializable)
	}

	if deserializable, isDeserializible := interface{}(vp).(DeserializableEntity); isDeserializible {
		return f(deserializable)
	}

	return Decode[Dest](b)
}

type EntityMigrationFunc[SrcEntity any, DestEntity any] func(srcKey kv.Key, srcVal SrcEntity) (destKey kv.Key, _ DestEntity)

func migrateEntitiesByPrefix[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, prefix string, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	count := uint32(0)

	srcContractState.Iterate(kv.Key(prefix), func(srcKey kv.Key, srcBytes []byte) bool {
		migrateEntityState(srcContractState, destContractState, srcKey, migrationFunc)
		count++

		return true
	})

	return count
}

func migrateEntitiesByKV[Src any, Dest any](srcEntities map[kv.Key][]byte, destContractState kv.KVStore, migrationFunc EntityMigrationFunc[Src, Dest]) {
	for srcKey, srcBytes := range srcEntities {
		newK, newV := migrateEntityBytes(srcKey, srcBytes, migrationFunc)
		destContractState.Set(newK, newV)
	}
}

func migrateEntitiesMapByName[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, oldMapName, newMapName string, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	if newMapName == "" {
		newMapName = oldMapName
	}

	srcEntities := collections.NewMapReadOnly(srcContractState, oldMapName)
	destEntities := collections.NewMap(destContractState, newMapName)

	migrateEntitiesMap(srcEntities, destEntities, migrationFunc)

	return srcEntities.Len()
}

func migrateEntitiesMap[Src any, Dest any](srcEntities *collections.ImmutableMap, destEntities *collections.Map, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcEntities.Iterate(func(srcKey, srcBytes []byte) bool {
		newK, newV := migrateEntityBytes(kv.Key(srcKey), srcBytes, migrationFunc)
		destEntities.SetAt([]byte(newK), newV)

		return true
	})
}

func migrateEntitiesArrayByName[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, oldArrName, newArrName string, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	if newArrName == "" {
		newArrName = oldArrName
	}

	srcEntities := collections.NewArrayReadOnly(srcContractState, oldArrName)
	destEntities := collections.NewArray(destContractState, newArrName)

	migrateEntitiesArray(srcEntities, destEntities, migrationFunc)

	return srcEntities.Len()
}

func migrateEntitiesArray[Src any, Dest any](srcEntities *collections.ArrayReadOnly, destEntities *collections.Array, migrationFunc EntityMigrationFunc[Src, Dest]) uint32 {
	for i := uint32(0); i < srcEntities.Len(); i++ {
		srcBytes := srcEntities.GetAt(i)
		_, newV := migrateEntityBytes("", srcBytes, migrationFunc)
		destEntities.SetAt(i, newV)
	}

	return srcEntities.Len()
}

func migrateEntityState[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, srcKey kv.Key, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcBytes := srcContractState.Get(srcKey)

	newKey, newVal := migrateEntityBytes(srcKey, srcBytes, migrationFunc)

	destContractState.Set(newKey, newVal)
}

func migrateEntityBytes[Src any, Dest any](srcKey kv.Key, srcBytes []byte, migrationFunc EntityMigrationFunc[Src, Dest]) (newKey kv.Key, newVal []byte) {
	srcEntity := must2(DeserializeEntity[Src](srcBytes))

	destKey, destEntity := migrationFunc(srcKey, srcEntity)

	return destKey, SerializeEntity(destEntity)
}

func list(store kv.KVStoreReader, prefix string) map[kv.Key][]byte {
	entries := map[kv.Key][]byte{}

	store.Iterate(kv.Key(prefix), func(key kv.Key, val []byte) bool {
		entries[key] = val
		return true
	})

	return entries
}

func SplitMapKey(storeKey kv.Key) (mapName, elemKey kv.Key) {
	const elemSep = "."
	pos := strings.Index(string(storeKey), elemSep)

	isMap := pos >= 0 && pos < len(storeKey)-1
	if !isMap {
		return storeKey, ""
	}

	return storeKey[:pos], storeKey[pos+1:]
}
