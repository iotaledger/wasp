// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"sort"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type ScImmutableDict struct {
	dict map[string][]byte
}

func (d *ScImmutableDict) Exists(key []byte) bool {
	return d.dict[string(key)] != nil
}

func (d *ScImmutableDict) Get(key []byte) []byte {
	return d.dict[string(key)]
}

type ScDict struct {
	ScImmutableDict
}

var _ wasmtypes.IKvStore = new(ScDict)

func NewScDict() *ScDict {
	return &ScDict{ScImmutableDict{dict: make(map[string][]byte)}}
}

func ScDictFromBytes(buf []byte) *ScDict {
	if len(buf) == 0 {
		return NewScDict()
	}
	dec := wasmtypes.NewWasmDecoder(buf)
	size := uint32(dec.VluDecode(32))
	dict := NewScDict()
	for i := uint32(0); i < size; i++ {
		keyLen := dec.VluDecode(32)
		key := dec.FixedBytes(uint32(keyLen))
		valLen := dec.VluDecode(32)
		val := dec.FixedBytes(uint32(valLen))
		dict.Set(key, val)
	}
	return dict
}

func (d *ScDict) AsProxy() wasmtypes.Proxy {
	return wasmtypes.NewProxy(d)
}

func (d *ScDict) Bytes() []byte {
	if d == nil {
		return []byte{0}
	}
	keys := make([]string, 0, len(d.dict))
	for key := range d.dict {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	enc := wasmtypes.NewWasmEncoder()
	enc.VluEncode(uint64(len(keys)))
	for _, k := range keys {
		key := []byte(k)
		val := d.dict[k]
		enc.VluEncode(uint64(len(key)))
		enc.FixedBytes(key, uint32(len(key)))
		enc.VluEncode(uint64(len(val)))
		enc.FixedBytes(val, uint32(len(val)))
	}
	return enc.Buf()
}

func (d *ScDict) Delete(key []byte) {
	delete(d.dict, string(key))
}

func (d *ScDict) Immutable() *ScImmutableDict {
	return &d.ScImmutableDict
}

func (d *ScDict) Set(key, value []byte) {
	d.dict[string(key)] = value
}
