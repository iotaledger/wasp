// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"sort"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
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
	size, buf := wasmcodec.ExtractUint32(buf)
	dict := make(ScDict, size)
	var k uint16
	var v uint32
	var key []byte
	var val []byte
	for i := uint32(0); i < size; i++ {
		k, buf = wasmcodec.ExtractUint16(buf)
		key, buf = wasmcodec.ExtractBytes(buf, int(k))
		v, buf = wasmcodec.ExtractUint32(buf)
		val, buf = wasmcodec.ExtractBytes(buf, int(v))
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
	buf := wasmcodec.AppendUint32(nil, uint32(len(keys)))
	for _, k := range keys {
		v := d[k]
		buf = wasmcodec.AppendUint16(buf, uint16(len(k)))
		buf = wasmcodec.AppendBytes(buf, []byte(k))
		buf = wasmcodec.AppendUint32(buf, uint32(len(v)))
		buf = wasmcodec.AppendBytes(buf, v)
	}
	return buf
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
