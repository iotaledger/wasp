// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"sort"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

type ScDict map[string][]byte

var _ wasmtypes.IKvStore = ScDict{}

func NewScDict() ScDict {
	return make(ScDict)
}

func NewScDictFromBytes(buf []byte) ScDict {
	if len(buf) == 0 {
		return make(ScDict)
	}
	dec := wasmtypes.NewWasmDecoder(buf)
	size := wasmtypes.Uint32FromBytes(dec.FixedBytes(wasmtypes.ScUint32Length))
	dict := make(ScDict, size)
	for i := uint32(0); i < size; i++ {
		keyLen := wasmtypes.Uint16FromBytes(dec.FixedBytes(wasmtypes.ScUint16Length))
		key := dec.FixedBytes(uint32(keyLen))
		valLen := wasmtypes.Uint32FromBytes(dec.FixedBytes(wasmtypes.ScUint32Length))
		val := dec.FixedBytes(valLen)
		dict.Set(key, val)
	}
	return dict
}

func (d ScDict) AsProxy() wasmtypes.Proxy {
	return wasmtypes.NewProxy(d)
}

func (d ScDict) Bytes() []byte {
	keys := make([]string, 0, len(d))
	for key := range d {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	enc := wasmtypes.NewWasmEncoder()
	enc.FixedBytes(wasmtypes.BytesFromUint32(uint32(len(keys))), wasmtypes.ScUint32Length)
	for _, k := range keys {
		key := []byte(k)
		val := d[k]
		enc.FixedBytes(wasmtypes.BytesFromUint16(uint16(len(key))), wasmtypes.ScUint16Length)
		enc.FixedBytes(key, uint32(len(key)))
		enc.FixedBytes(wasmtypes.BytesFromUint32(uint32(len(val))), wasmtypes.ScUint32Length)
		enc.FixedBytes(val, uint32(len(val)))
	}
	return enc.Buf()
}

func (d ScDict) Delete(key []byte) {
	delete(d, string(key))
}

func (d ScDict) Exists(key []byte) bool {
	return d[string(key)] != nil
}

func (d ScDict) Get(key []byte) []byte {
	return d[string(key)]
}

func (d ScDict) Set(key, value []byte) {
	d[string(key)] = value
}

func (d ScDict) Immutable() ScImmutableDict {
	return ScImmutableDict{d: d}
}

type ScImmutableDict struct {
	d ScDict
}

func (d ScImmutableDict) Exists(key []byte) bool {
	return d.d.Exists(key)
}

func (d ScImmutableDict) Get(key []byte) []byte {
	return d.d.Get(key)
}
