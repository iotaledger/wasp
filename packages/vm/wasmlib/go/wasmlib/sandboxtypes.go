// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"encoding/binary"
	"sort"
)

type Map interface {
	Delete(key []byte)
	Exists(key []byte) bool
	Get(key []byte) []byte
	Set(key, value []byte)
}

type ScAssets map[ScColor]uint64

func NewScAssetsFromBytes(buf []byte) ScAssets {
	dict := make(ScAssets)
	size, buf := ExtractUint32(buf)
	var k []byte
	var v uint64
	for i := uint32(0); i < size; i++ {
		k, buf = ExtractBytes(buf, 32)
		v, buf = ExtractUint64(buf)
		dict[NewScColorFromBytes(k)] = v
	}
	return dict
}

func (a ScAssets) Bytes() []byte {
	keys := make([]ScColor, 0, len(a))
	for key := range a {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i].id[:]) < string(keys[j].id[:])
	})
	buf := AppendUint32(nil, uint32(len(keys)))
	for _, k := range keys {
		v := a[k]
		buf = AppendBytes(buf, k.Bytes())
		buf = AppendUint64(buf, v)
	}
	return buf
}

type ScDict map[string][]byte

var _ Map = new(ScDict)

func NewScDict() ScDict {
	return make(ScDict)
}

func NewScDictFromBytes(buf []byte) ScDict {
	size, buf := ExtractUint32(buf)
	dict := make(ScDict, size)
	var k uint16
	var v uint32
	var key []byte
	var val []byte
	for i := uint32(0); i < size; i++ {
		k, buf = ExtractUint16(buf)
		key, buf = ExtractBytes(buf, int(k))
		v, buf = ExtractUint32(buf)
		val, buf = ExtractBytes(buf, int(v))
		dict.Set(key, val)
	}
	return dict
}

func (d ScDict) Bytes() []byte {
	keys := make([]string, 0, len(d))
	for key := range d {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	buf := AppendUint32(nil, uint32(len(keys)))
	for _, k := range keys {
		v := d[k]
		buf = AppendUint16(buf, uint16(len(k)))
		buf = AppendBytes(buf, []byte(k))
		buf = AppendUint32(buf, uint32(len(v)))
		buf = AppendBytes(buf, v)
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

/////////////////////////////////////////////////

type ScState struct{}

var _ Map = new(ScState)

func (d ScState) Delete(key []byte) {
	StateDelete(key)
}

func (d ScState) Exists(key []byte) bool {
	return StateExists(key)
}

func (d ScState) Get(key []byte) []byte {
	return StateGet(key)
}

func (d ScState) Set(key, value []byte) {
	StateSet(key, value)
}

/////////////////////////////////////////////////

func ExtractBool(buf []byte) (bool, []byte) {
	return buf[0] != 0, buf[1:]
}

func ExtractBytes(buf []byte, n int) ([]byte, []byte) {
	return buf[:n], buf[n:]
}

func ExtractInt8(buf []byte) (int8, []byte) {
	return int8(buf[0]), buf[1:]
}

func ExtractInt16(buf []byte) (int16, []byte) {
	return int16(binary.LittleEndian.Uint16(buf)), buf[2:]
}

func ExtractInt32(buf []byte) (int32, []byte) {
	return int32(binary.LittleEndian.Uint32(buf)), buf[4:]
}

func ExtractInt64(buf []byte) (int64, []byte) {
	return int64(binary.LittleEndian.Uint64(buf)), buf[8:]
}

func ExtractUint8(buf []byte) (uint8, []byte) {
	return buf[0], buf[1:]
}

func ExtractUint16(buf []byte) (uint16, []byte) {
	return binary.LittleEndian.Uint16(buf), buf[2:]
}

func ExtractUint32(buf []byte) (uint32, []byte) {
	return binary.LittleEndian.Uint32(buf), buf[4:]
}

func ExtractUint64(buf []byte) (uint64, []byte) {
	return binary.LittleEndian.Uint64(buf), buf[8:]
}

////////////////////////////////////

func AppendBool(buf []byte, value bool) []byte {
	if value {
		return AppendUint8(buf, 1)
	}
	return AppendUint8(buf, 0)
}

func AppendBytes(buf, value []byte) []byte {
	return append(buf, value...)
}

func AppendInt8(buf []byte, value int8) []byte {
	return AppendUint8(buf, uint8(value))
}

func AppendInt16(buf []byte, value int16) []byte {
	return AppendUint16(buf, uint16(value))
}

func AppendInt32(buf []byte, value int32) []byte {
	return AppendUint32(buf, uint32(value))
}

func AppendInt64(buf []byte, value int64) []byte {
	return AppendUint64(buf, uint64(value))
}

func AppendUint8(buf []byte, value uint8) []byte {
	return append(buf, value)
}

func AppendUint16(buf []byte, value uint16) []byte {
	tmp := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmp, value)
	return append(buf, tmp...)
}

func AppendUint32(buf []byte, value uint32) []byte {
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, value)
	return append(buf, tmp...)
}

func AppendUint64(buf []byte, value uint64) []byte {
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, value)
	return append(buf, tmp...)
}
