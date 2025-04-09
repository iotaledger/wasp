package utils

import (
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"

	"github.com/iotaledger/wasp/packages/kv"
)

type NoopKVStoreReader[KeyType any] struct {
}

var _ old_kv.KVStoreReader = &NoopKVStoreReader[old_kv.Key]{}
var _ kv.KVStoreReader = &NoopKVStoreReader[kv.Key]{}

func (NoopKVStoreReader[KeyType]) Get(key KeyType) []byte {
	return nil
}
func (NoopKVStoreReader[KeyType]) Has(key KeyType) bool {
	return false
}
func (NoopKVStoreReader[KeyType]) Iterate(prefix KeyType, f func(key KeyType, value []byte) bool) {
}
func (NoopKVStoreReader[KeyType]) IterateKeys(prefix KeyType, f func(key KeyType) bool) {
}
func (NoopKVStoreReader[KeyType]) IterateSorted(prefix KeyType, f func(key KeyType, value []byte) bool) {
}
func (NoopKVStoreReader[KeyType]) IterateKeysSorted(prefix KeyType, f func(key KeyType) bool) {
}
