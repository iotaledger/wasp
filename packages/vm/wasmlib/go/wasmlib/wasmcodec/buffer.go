// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmcodec

import (
	"encoding/binary"
)

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
