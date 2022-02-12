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

func NewScDictFromBytes(buf []byte) *ScDict {
	if len(buf) == 0 {
		return NewScDict()
	}
	dec := wasmtypes.NewWasmDecoder(buf)
	size := wasmtypes.Uint32FromBytes(dec.FixedBytes(wasmtypes.ScUint32Length))
	dict := NewScDict()
	for i := uint32(0); i < size; i++ {
		keyBuf := dec.FixedBytes(wasmtypes.ScUint16Length)
		keyLen := wasmtypes.Uint16FromBytes(keyBuf)
		key := dec.FixedBytes(uint32(keyLen))
		valBuf := dec.FixedBytes(wasmtypes.ScUint32Length)
		valLen := wasmtypes.Uint32FromBytes(valBuf)
		val := dec.FixedBytes(valLen)
		dict.Set(key, val)
	}
	return dict
}

func (d *ScDict) AsProxy() wasmtypes.Proxy {
	return wasmtypes.NewProxy(d)
}

func (d *ScDict) Bytes() []byte {
	if d == nil {
		return []byte{0, 0, 0, 0}
	}
	keys := make([]string, 0, len(d.dict))
	for key := range d.dict {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	enc := wasmtypes.NewWasmEncoder()
	enc.FixedBytes(wasmtypes.Uint32ToBytes(uint32(len(keys))), wasmtypes.ScUint32Length)
	for _, k := range keys {
		key := []byte(k)
		val := d.dict[k]
		enc.FixedBytes(wasmtypes.Uint16ToBytes(uint16(len(key))), wasmtypes.ScUint16Length)
		enc.FixedBytes(key, uint32(len(key)))
		enc.FixedBytes(wasmtypes.Uint32ToBytes(uint32(len(val))), wasmtypes.ScUint32Length)
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
