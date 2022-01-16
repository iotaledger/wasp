// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"encoding/binary"
	"sort"
)

type ScAssets map[ScColor]uint64

func NewScAssetsFromBytes(buf []byte) ScAssets {
	dict := make(ScAssets)
	size := NewUint32FromBytes(buf[:4])
	buf = buf[4:]
	for i := uint32(0); i < size; i++ {
		k := buf[:32]
		buf = buf[32:]
		v := NewUint64FromBytes(buf[:8])
		buf = buf[8:]
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
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(keys)))
	for _, k := range keys {
		v := a[k]
		val := make([]byte, 8)
		binary.LittleEndian.PutUint64(val, v)
		buf = append(buf, k.Bytes()...)
		buf = append(buf, val...)
	}
	return buf
}

type ScDict map[string][]byte

func NewScDictFromBytes(buf []byte) ScDict {
	dict := make(ScDict)
	size := NewUint32FromBytes(buf[:4])
	buf = buf[4:]
	for i := uint32(0); i < size; i++ {
		k := NewUint16FromBytes(buf[:2])
		buf = buf[2:]
		key := buf[:k]
		buf = buf[k:]
		v := NewUint32FromBytes(buf[:4])
		buf = buf[4:]
		val := buf[:v]
		buf = buf[v:]
		dict[string(key)] = val
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
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(keys)))
	for _, k := range keys {
		v := d[k]
		keyLen := make([]byte, 2)
		binary.LittleEndian.PutUint16(keyLen, uint16(len(k)))
		valLen := make([]byte, 4)
		binary.LittleEndian.PutUint32(valLen, uint32(len(v)))
		buf = append(buf, keyLen...)
		buf = append(buf, []byte(k)...)
		buf = append(buf, valLen...)
		buf = append(buf, v...)
	}
	return buf
}

func NewBoolFromBytes(bytes []byte) bool {
	return bytes[0] != 0
}

func NewInt16FromBytes(bytes []byte) int16 {
	return int16(binary.LittleEndian.Uint16(bytes))
}

func NewInt32FromBytes(bytes []byte) int32 {
	return int32(binary.LittleEndian.Uint32(bytes))
}

func NewInt64FromBytes(bytes []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bytes))
}

func NewUint16FromBytes(bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(bytes)
}

func NewUint32FromBytes(bytes []byte) uint32 {
	return binary.LittleEndian.Uint32(bytes)
}

func NewUint64FromBytes(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}
