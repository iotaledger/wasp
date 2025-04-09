package utils

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
)

type KVStoreReader[Key ~string] interface {
	Get(key Key) []byte
	Has(key Key) bool
	Iterate(prefix Key, f func(key Key, value []byte) bool)
	IterateKeys(prefix Key, f func(key Key) bool)
	IterateSorted(prefix Key, f func(key Key, value []byte) bool)
	IterateKeysSorted(prefix Key, f func(key Key) bool)
}

type KVStoreWriter[Key ~string] interface {
	Set(key Key, value []byte)
	Del(key Key)
}

type KVStore[Key ~string] interface {
	KVStoreReader[Key]
	KVStoreWriter[Key]
}

var _ old_kv.KVWriter = KVStoreWriter[old_kv.Key](nil)
var _ old_kv.KVReader = KVStoreReader[old_kv.Key](nil)
var _ old_kv.KVIterator = KVStoreReader[old_kv.Key](nil)
var _ kv.KVReader = KVStoreReader[kv.Key](nil)
var _ kv.KVWriter = KVStoreWriter[kv.Key](nil)
var _ kv.KVIterator = KVStoreReader[kv.Key](nil)

func NewRecordingKVStoreReadOnly[Key ~string, Reader KVStoreReader[Key]](r Reader) *RecordingKVStore[Key, Reader, KVStoreWriter[Key]] {
	return &RecordingKVStore[Key, Reader, KVStoreWriter[Key]]{
		KVStoreReader: r,
		R:             r,
	}
}

func NewRecordingKVStore[Key ~string, Store KVStore[Key]](s Store) *RecordingKVStore[Key, Store, Store] {
	return &RecordingKVStore[Key, Store, Store]{
		KVStoreReader: s,
		KVStoreWriter: s,
		R:             s,
		W:             s,
	}
}

type RecordingKVStore[Key ~string, Reader KVStoreReader[Key], Writer KVStoreWriter[Key]] struct {
	KVStoreReader[Key]
	KVStoreWriter[Key]
	R Reader
	W Writer

	lastReadPrefix Key
	lastReadKey    Key
	lastReadValue  []byte
	lastWriteKey   Key
	lastWriteValue []byte
}

var _ old_kv.KVStore = &RecordingKVStore[old_kv.Key, *PrefixKVStore, *PrefixKVStore]{}
var _ old_kv.KVStoreReader = &RecordingKVStore[old_kv.Key, *PrefixKVStore, *PrefixKVStore]{}
var _ kv.KVStore = &RecordingKVStore[kv.Key, *InMemoryKVStore, *InMemoryKVStore]{}
var _ kv.KVStoreReader = &RecordingKVStore[kv.Key, *InMemoryKVStore, *InMemoryKVStore]{}

func (s *RecordingKVStore[Key, _, _]) Get(key Key) []byte {
	s.lastReadPrefix = ""
	s.lastReadKey = key
	s.lastReadValue = s.R.Get(key)

	return s.lastReadValue
}

func (s *RecordingKVStore[Key, _, _]) Set(key Key, value []byte) {
	s.lastWriteKey = key
	s.lastWriteValue = value

	s.W.Set(key, value)
}

func (s *RecordingKVStore[Key, _, _]) Del(key Key) {
	s.lastWriteKey = key
	s.lastWriteValue = nil

	s.W.Del(key)
}

func (s *RecordingKVStore[Key, _, _]) Iterate(prefix Key, f func(key Key, value []byte) bool) {
	s.lastReadPrefix = prefix

	s.R.Iterate(prefix, func(key Key, value []byte) bool {
		s.lastReadKey = key
		s.lastReadValue = value

		return f(key, value)
	})
}

func (s *RecordingKVStore[Key, _, _]) IterateKeys(prefix Key, f func(key Key) bool) {
	s.Iterate(prefix, func(key Key, _ []byte) bool {
		return f(key)
	})
}

func (s *RecordingKVStore[Key, _, _]) IterateSorted(prefix Key, f func(key Key, value []byte) bool) {
	s.lastReadPrefix = prefix

	s.R.IterateSorted(prefix, func(key Key, value []byte) bool {
		s.lastReadKey = key
		s.lastReadValue = value

		return f(key, value)
	})
}

func (s *RecordingKVStore[Key, _, _]) IterateKeysSorted(prefix Key, f func(key Key) bool) {
	s.IterateSorted(prefix, func(key Key, _ []byte) bool {
		return f(key)
	})
}

func (s *RecordingKVStore[Key, _, _]) LastRead() (prefix Key, key Key, value []byte) {
	return s.lastReadPrefix, s.lastReadKey, s.lastReadValue
}

func (s *RecordingKVStore[Key, _, _]) LastWrite() (key Key, value []byte) {
	return s.lastWriteKey, s.lastWriteValue
}

func PrintLastDBOperations[SrcKey ~string, SrcReader KVStoreReader[SrcKey], SrcWriter KVStoreWriter[SrcKey], DestKey ~string, DestReader KVStoreReader[DestKey], DestWriter KVStoreWriter[DestKey]](
	oldState *RecordingKVStore[SrcKey, SrcReader, SrcWriter],
	newState *RecordingKVStore[DestKey, DestReader, DestWriter],
) {
	lastSrcReadPrefix, lastSrcReadKey, lastSrcReadValue := oldState.LastRead()
	if lastSrcReadPrefix == "" {
		cli.Logf("Last src db read (get):\n\tKey:\n\t\t%x\n\t\t%v\n\tValue:\n\t\t%x\n\t\t%v\n",
			lastSrcReadKey, string(lastSrcReadKey), lastSrcReadValue, string(lastSrcReadValue))
	} else {
		cli.Logf("Last src db read (iter):\n\tPrefix:\n\t\t%x\n\t\t%v\n\tKey:\n\t\t%x\n\t\t%v\n\tValue:\n\t\t%x\n\t\t%v\n",
			lastSrcReadPrefix, string(lastSrcReadPrefix), lastSrcReadKey, string(lastSrcReadKey), lastSrcReadValue, string(lastSrcReadValue))
	}

	lastDestReadPrefix, lastDestReadKey, lastDestReadValue := newState.LastRead()
	if lastDestReadPrefix == "" {
		cli.Logf("Last dest db read (get):\n\tKey:\n\t\t%x\n\t\t%v\n\tValue:\n\t\t%x\n\t\t%v\n",
			lastDestReadKey, string(lastDestReadKey), lastDestReadValue, string(lastDestReadValue))
	} else {
		cli.Logf("Last dest db read (iter):\n\tPrefix:\n\t\t%x\n\t\t%v\n\tKey:\n\t\t%x\n\t\t%v\n\tValue:\n\t\t%x\n\t\t%v\n",
			lastDestReadPrefix, string(lastDestReadPrefix), lastDestReadKey, string(lastDestReadKey), lastDestReadValue, string(lastDestReadValue))
	}

	lastDestWriteKey, lastDestWriteValue := newState.LastWrite()
	if lastDestWriteKey != "" {
		cli.Logf("Last dest db write:\n\tKey:\n\t\t%x\n\t\t%v\n\tValue:\n\t\t%x\n\t\t%v\n",
			lastDestWriteKey, string(lastDestWriteKey), lastDestWriteValue, string(lastDestWriteValue))
	}
}
