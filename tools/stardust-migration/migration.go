package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

type SerializableEntity interface {
	Bytes() []byte
}

type DeserializableEntity interface {
	Read(src io.Reader) error
}

func SerializeEntity(entity any) []byte {
	if serializable, ok := entity.(SerializableEntity); ok {
		return serializable.Bytes()
	}

	return codec.Encode(entity)
}

type EntityMigrationFunc[SrcEntity any, DestEntity any] func(srcKey kv.Key, srcVal SrcEntity) (destKey kv.Key, _ DestEntity)

func migrateEntitiesByPrefix[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, prefix string, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcContractState.Iterate(kv.Key(prefix), func(srcKey kv.Key, srcBytes []byte) bool {
		migrateEntityState(srcContractState, destContractState, srcKey, migrationFunc)
		return true
	})
}

func migrateEntitiesByKV[Src any, Dest any](srcEntities map[kv.Key][]byte, destContractState kv.KVStore, migrationFunc EntityMigrationFunc[Src, Dest]) {
	for srcKey, srcBytes := range srcEntities {
		newK, newV := migrateEntityBytes(srcKey, srcBytes, migrationFunc)
		destContractState.Set(newK, newV)
	}
}

func migrateEntitiesMapByName[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, mapName string, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcEntities := collections.NewMapReadOnly(srcContractState, mapName)
	destEntities := collections.NewMap(destContractState, mapName)

	migrateEntitiesMap(srcEntities, destEntities, migrationFunc)
}

func migrateEntitiesMap[Src any, Dest any](srcEntities *collections.ImmutableMap, destEntities *collections.Map, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcEntities.Iterate(func(srcKey, srcBytes []byte) bool {
		newK, newV := migrateEntityBytes(kv.Key(srcKey), srcBytes, migrationFunc)
		destEntities.SetAt([]byte(newK), newV)

		return true
	})
}

func migrateEntityState[Src any, Dest any](srcContractState kv.KVStoreReader, destContractState kv.KVStore, srcKey kv.Key, migrationFunc EntityMigrationFunc[Src, Dest]) {
	srcBytes := srcContractState.Get(srcKey)

	newKey, newVal := migrateEntityBytes(srcKey, srcBytes, migrationFunc)

	destContractState.Set(newKey, newVal)
}

func migrateEntityBytes[Src any, Dest any](srcKey kv.Key, srcBytes []byte, migrationFunc EntityMigrationFunc[Src, Dest]) (newKey kv.Key, newVal []byte) {

	var srcEntity Src

	if readableSrcEntity, isReadable := interface{}(srcEntity).(DeserializableEntity); isReadable {
		r := bytes.NewReader(srcBytes)
		must(readableSrcEntity.Read(r))

		if r.Len() != 0 {
			leftovers := must2(io.ReadAll(r))
			panic(fmt.Sprintf("Leftover bytes after reading entity of type %T: initialValue = %x, leftover = %x, leftoverLen = %v", srcEntity, srcBytes, leftovers, r.Len()))
		}
	} else {
		srcEntity = must2(Decode[Src](srcBytes))
	}

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
