package kv

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// Since map cannot have []byte as key, to avoid unnecessary conversions
// between string and []byte, we use string as key data type, but it does
// not necessarily have to be a valid UTF-8 string.
type Key string

func (k Key) Hex() string {
	return fmt.Sprintf("kv.Key('%X')", k)
}

const EmptyPrefix = Key("")

func (k Key) HasPrefix(prefix Key) bool {
	if len(prefix) > len(k) {
		return false
	}
	return k[:len(prefix)] == prefix
}

// KVStore represents a key-value store
// where both keys and values are arbitrary byte slices.
type KVStore interface {
	KVWriter
	KVStoreReader
}

type KVReader interface {
	// Get returns the value, or nil if not found
	Get(key Key) ([]byte, error)
	Has(key Key) (bool, error)
}

type KVWriter interface {
	Set(key Key, value []byte)
	Del(key Key)

	// TODO add DelPrefix(prefix []byte)
}

type KVIterator interface {
	Iterate(prefix Key, f func(key Key, value []byte) bool) error
	IterateKeys(prefix Key, f func(key Key) bool) error
	IterateSorted(prefix Key, f func(key Key, value []byte) bool) error
	IterateKeysSorted(prefix Key, f func(key Key) bool) error
}

type KVMustReader interface {
	// MustGet returns the value, or nil if not found
	MustGet(key Key) []byte
	MustHas(key Key) bool
}

type KVMustIterator interface {
	MustIterate(prefix Key, f func(key Key, value []byte) bool)
	MustIterateKeys(prefix Key, f func(key Key) bool)
	MustIterateSorted(prefix Key, f func(key Key, value []byte) bool)
	MustIterateKeysSorted(prefix Key, f func(key Key) bool)
}

type KVStoreReader interface {
	KVReader
	KVIterator
	KVMustReader
	KVMustIterator
}

func MustGet(kvs KVStoreReader, key Key) []byte {
	v, err := kvs.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}

func MustHas(kvs KVStoreReader, key Key) bool {
	v, err := kvs.Has(key)
	if err != nil {
		panic(err)
	}
	return v
}

func MustIterate(kvs KVStoreReader, prefix Key, f func(key Key, value []byte) bool) {
	err := kvs.Iterate(prefix, f)
	if err != nil {
		panic(err)
	}
}

func MustIterateKeys(kvs KVStoreReader, prefix Key, f func(key Key) bool) {
	err := kvs.IterateKeys(prefix, f)
	if err != nil {
		panic(err)
	}
}

func MustIterateSorted(kvs KVStoreReader, prefix Key, f func(key Key, value []byte) bool) {
	err := kvs.IterateSorted(prefix, f)
	if err != nil {
		panic(err)
	}
}

func MustIterateKeysSorted(kvs KVStoreReader, prefix Key, f func(key Key) bool) {
	err := kvs.IterateKeysSorted(prefix, f)
	if err != nil {
		panic(err)
	}
}

func Concat(fragments ...interface{}) []byte {
	var buf bytes.Buffer
	for _, v := range fragments {
		switch v := v.(type) {
		case string:
			buf.WriteString(v)
		case []byte:
			buf.Write(v)
		case byte:
			buf.WriteByte(v)
		case uint16:
			buf.Write(util.Uint16To2Bytes(v))
		case uint32:
			buf.Write(util.Uint32To4Bytes(v))
		case uint64:
			buf.Write(util.Uint64To8Bytes(v))
		case interface{ Bytes() []byte }:
			buf.Write(v.Bytes())
		default:
			panic(xerrors.Errorf("Concat: unknown key fragment type %T", v))
		}
	}
	return buf.Bytes()
}

const nilprefix = ""

func ByteSize(s KVStoreReader) int {
	accLen := 0
	err := s.Iterate(nilprefix, func(k Key, v []byte) bool {
		accLen += len([]byte(k)) + len(v)
		return true
	})
	if err != nil {
		return 0
	}
	return accLen
}
