// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"encoding/binary"
	"strconv"
)

type ScImmutableAddress struct {
	objID int32
	keyID Key32
}

func NewScImmutableAddress(objID int32, keyID Key32) ScImmutableAddress {
	return ScImmutableAddress{objID: objID, keyID: keyID}
}

func (o ScImmutableAddress) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_ADDRESS)
}

func (o ScImmutableAddress) String() string {
	return o.Value().String()
}

func (o ScImmutableAddress) Value() ScAddress {
	return NewScAddressFromBytes(GetBytes(o.objID, o.keyID, TYPE_ADDRESS))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableAddressArray struct {
	objID int32
}

func (o ScImmutableAddressArray) GetAddress(index int32) ScImmutableAddress {
	return ScImmutableAddress{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableAddressArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableAgentID struct {
	objID int32
	keyID Key32
}

func NewScImmutableAgentID(objID int32, keyID Key32) ScImmutableAgentID {
	return ScImmutableAgentID{objID: objID, keyID: keyID}
}

func (o ScImmutableAgentID) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_AGENT_ID)
}

func (o ScImmutableAgentID) String() string {
	return o.Value().String()
}

func (o ScImmutableAgentID) Value() ScAgentID {
	return NewScAgentIDFromBytes(GetBytes(o.objID, o.keyID, TYPE_AGENT_ID))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableAgentIDArray struct {
	objID int32
}

func (o ScImmutableAgentIDArray) GetAgentID(index int32) ScImmutableAgentID {
	return ScImmutableAgentID{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableAgentIDArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableBool struct {
	objID int32
	keyID Key32
}

func NewScImmutableBool(objID int32, keyID Key32) ScImmutableBool {
	return ScImmutableBool{objID: objID, keyID: keyID}
}

func (o ScImmutableBool) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_BOOL)
}

func (o ScImmutableBool) String() string {
	if o.Value() {
		return "1"
	}
	return "0"
}

func (o ScImmutableBool) Value() bool {
	bytes := GetBytes(o.objID, o.keyID, TYPE_BOOL)
	return bytes[0] != 0
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableBoolArray struct {
	objID int32
}

func (o ScImmutableBoolArray) GetBool(index int32) ScImmutableBool {
	return ScImmutableBool{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableBoolArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableBytes struct {
	objID int32
	keyID Key32
}

func NewScImmutableBytes(objID int32, keyID Key32) ScImmutableBytes {
	return ScImmutableBytes{objID: objID, keyID: keyID}
}

func (o ScImmutableBytes) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_BYTES)
}

func (o ScImmutableBytes) String() string {
	return base58Encode(o.Value())
}

func (o ScImmutableBytes) Value() []byte {
	return GetBytes(o.objID, o.keyID, TYPE_BYTES)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableBytesArray struct {
	objID int32
}

func (o ScImmutableBytesArray) GetBytes(index int32) ScImmutableBytes {
	return ScImmutableBytes{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableBytesArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableChainID struct {
	objID int32
	keyID Key32
}

func NewScImmutableChainID(objID int32, keyID Key32) ScImmutableChainID {
	return ScImmutableChainID{objID: objID, keyID: keyID}
}

func (o ScImmutableChainID) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_CHAIN_ID)
}

func (o ScImmutableChainID) String() string {
	return o.Value().String()
}

func (o ScImmutableChainID) Value() ScChainID {
	return NewScChainIDFromBytes(GetBytes(o.objID, o.keyID, TYPE_CHAIN_ID))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableChainIDArray struct {
	objID int32
}

func (o ScImmutableChainIDArray) GetChainID(index int32) ScImmutableChainID {
	return ScImmutableChainID{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableChainIDArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableColor struct {
	objID int32
	keyID Key32
}

func NewScImmutableColor(objID int32, keyID Key32) ScImmutableColor {
	return ScImmutableColor{objID: objID, keyID: keyID}
}

func (o ScImmutableColor) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_COLOR)
}

func (o ScImmutableColor) String() string {
	return o.Value().String()
}

func (o ScImmutableColor) Value() ScColor {
	return NewScColorFromBytes(GetBytes(o.objID, o.keyID, TYPE_COLOR))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableColorArray struct {
	objID int32
}

func (o ScImmutableColorArray) GetColor(index int32) ScImmutableColor {
	return ScImmutableColor{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableColorArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableHash struct {
	objID int32
	keyID Key32
}

func NewScImmutableHash(objID int32, keyID Key32) ScImmutableHash {
	return ScImmutableHash{objID: objID, keyID: keyID}
}

func (o ScImmutableHash) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_HASH)
}

func (o ScImmutableHash) String() string {
	return o.Value().String()
}

func (o ScImmutableHash) Value() ScHash {
	return NewScHashFromBytes(GetBytes(o.objID, o.keyID, TYPE_HASH))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableHashArray struct {
	objID int32
}

func (o ScImmutableHashArray) GetHash(index int32) ScImmutableHash {
	return ScImmutableHash{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableHashArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableHname struct {
	objID int32
	keyID Key32
}

func NewScImmutableHname(objID int32, keyID Key32) ScImmutableHname {
	return ScImmutableHname{objID: objID, keyID: keyID}
}

func (o ScImmutableHname) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_HNAME)
}

func (o ScImmutableHname) String() string {
	return o.Value().String()
}

func (o ScImmutableHname) Value() ScHname {
	return NewScHnameFromBytes(GetBytes(o.objID, o.keyID, TYPE_HNAME))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableHnameArray struct {
	objID int32
}

func (o ScImmutableHnameArray) GetHname(index int32) ScImmutableHname {
	return ScImmutableHname{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableHnameArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt8 struct {
	objID int32
	keyID Key32
}

func NewScImmutableInt8(objID int32, keyID Key32) ScImmutableInt8 {
	return ScImmutableInt8{objID: objID, keyID: keyID}
}

func (o ScImmutableInt8) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT8)
}

func (o ScImmutableInt8) String() string {
	return strconv.FormatInt(int64(o.Value()), 10)
}

func (o ScImmutableInt8) Value() int8 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT8)
	return int8(bytes[0])
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt8Array struct {
	objID int32
}

func (o ScImmutableInt8Array) GetInt8(index int32) ScImmutableInt8 {
	return ScImmutableInt8{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableInt8Array) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt16 struct {
	objID int32
	keyID Key32
}

func NewScImmutableInt16(objID int32, keyID Key32) ScImmutableInt16 {
	return ScImmutableInt16{objID: objID, keyID: keyID}
}

func (o ScImmutableInt16) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT16)
}

func (o ScImmutableInt16) String() string {
	return strconv.FormatInt(int64(o.Value()), 10)
}

func (o ScImmutableInt16) Value() int16 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT16)
	return int16(binary.LittleEndian.Uint16(bytes))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt16Array struct {
	objID int32
}

func (o ScImmutableInt16Array) GetInt16(index int32) ScImmutableInt16 {
	return ScImmutableInt16{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableInt16Array) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt32 struct {
	objID int32
	keyID Key32
}

func NewScImmutableInt32(objID int32, keyID Key32) ScImmutableInt32 {
	return ScImmutableInt32{objID: objID, keyID: keyID}
}

func (o ScImmutableInt32) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT32)
}

func (o ScImmutableInt32) String() string {
	return strconv.FormatInt(int64(o.Value()), 10)
}

func (o ScImmutableInt32) Value() int32 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT32)
	return int32(binary.LittleEndian.Uint32(bytes))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt32Array struct {
	objID int32
}

func (o ScImmutableInt32Array) GetInt32(index int32) ScImmutableInt32 {
	return ScImmutableInt32{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableInt32Array) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt64 struct {
	objID int32
	keyID Key32
}

func NewScImmutableInt64(objID int32, keyID Key32) ScImmutableInt64 {
	return ScImmutableInt64{objID: objID, keyID: keyID}
}

func (o ScImmutableInt64) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT64)
}

func (o ScImmutableInt64) String() string {
	return strconv.FormatInt(o.Value(), 10)
}

func (o ScImmutableInt64) Value() int64 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT64)
	return int64(binary.LittleEndian.Uint64(bytes))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt64Array struct {
	objID int32
}

func (o ScImmutableInt64Array) GetInt64(index int32) ScImmutableInt64 {
	return ScImmutableInt64{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableInt64Array) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableMap struct {
	objID int32
}

func (o ScImmutableMap) CallFunc(keyID Key32, params []byte) []byte {
	return CallFunc(o.objID, keyID, params)
}

func (o ScImmutableMap) GetAddress(key MapKey) ScImmutableAddress {
	return ScImmutableAddress{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetAddressArray(key MapKey) ScImmutableAddressArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_ADDRESS|TYPE_ARRAY)
	return ScImmutableAddressArray{objID: arrID}
}

func (o ScImmutableMap) GetAgentID(key MapKey) ScImmutableAgentID {
	return ScImmutableAgentID{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetAgentIDArray(key MapKey) ScImmutableAgentIDArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_AGENT_ID|TYPE_ARRAY)
	return ScImmutableAgentIDArray{objID: arrID}
}

func (o ScImmutableMap) GetBool(key MapKey) ScImmutableBool {
	return ScImmutableBool{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetBoolArray(key MapKey) ScImmutableBoolArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_BOOL|TYPE_ARRAY)
	return ScImmutableBoolArray{objID: arrID}
}

func (o ScImmutableMap) GetBytes(key MapKey) ScImmutableBytes {
	return ScImmutableBytes{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetBytesArray(key MapKey) ScImmutableBytesArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_BYTES|TYPE_ARRAY)
	return ScImmutableBytesArray{objID: arrID}
}

func (o ScImmutableMap) GetChainID(key MapKey) ScImmutableChainID {
	return ScImmutableChainID{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetChainIDArray(key MapKey) ScImmutableChainIDArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_CHAIN_ID|TYPE_ARRAY)
	return ScImmutableChainIDArray{objID: arrID}
}

func (o ScImmutableMap) GetColor(key MapKey) ScImmutableColor {
	return ScImmutableColor{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetColorArray(key MapKey) ScImmutableColorArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_COLOR|TYPE_ARRAY)
	return ScImmutableColorArray{objID: arrID}
}

func (o ScImmutableMap) GetHash(key MapKey) ScImmutableHash {
	return ScImmutableHash{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetHashArray(key MapKey) ScImmutableHashArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_HASH|TYPE_ARRAY)
	return ScImmutableHashArray{objID: arrID}
}

func (o ScImmutableMap) GetHname(key MapKey) ScImmutableHname {
	return ScImmutableHname{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetHnameArray(key MapKey) ScImmutableHnameArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_HNAME|TYPE_ARRAY)
	return ScImmutableHnameArray{objID: arrID}
}

func (o ScImmutableMap) GetInt8(key MapKey) ScImmutableInt8 {
	return ScImmutableInt8{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetInt8Array(key MapKey) ScImmutableInt8Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT8|TYPE_ARRAY)
	return ScImmutableInt8Array{objID: arrID}
}

func (o ScImmutableMap) GetInt16(key MapKey) ScImmutableInt16 {
	return ScImmutableInt16{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetInt16Array(key MapKey) ScImmutableInt16Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT16|TYPE_ARRAY)
	return ScImmutableInt16Array{objID: arrID}
}

func (o ScImmutableMap) GetInt32(key MapKey) ScImmutableInt32 {
	return ScImmutableInt32{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetInt32Array(key MapKey) ScImmutableInt32Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT32|TYPE_ARRAY)
	return ScImmutableInt32Array{objID: arrID}
}

func (o ScImmutableMap) GetInt64(key MapKey) ScImmutableInt64 {
	return ScImmutableInt64{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetInt64Array(key MapKey) ScImmutableInt64Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT64|TYPE_ARRAY)
	return ScImmutableInt64Array{objID: arrID}
}

func (o ScImmutableMap) GetMap(key MapKey) ScImmutableMap {
	mapID := GetObjectID(o.objID, key.KeyID(), TYPE_MAP)
	return ScImmutableMap{objID: mapID}
}

func (o ScImmutableMap) GetMapArray(key MapKey) ScImmutableMapArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_MAP|TYPE_ARRAY)
	return ScImmutableMapArray{objID: arrID}
}

func (o ScImmutableMap) GetRequestID(key MapKey) ScImmutableRequestID {
	return ScImmutableRequestID{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetRequestIDArray(key MapKey) ScImmutableRequestIDArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_REQUEST_ID|TYPE_ARRAY)
	return ScImmutableRequestIDArray{objID: arrID}
}

func (o ScImmutableMap) GetString(key MapKey) ScImmutableString {
	return ScImmutableString{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetStringArray(key MapKey) ScImmutableStringArray {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_STRING|TYPE_ARRAY)
	return ScImmutableStringArray{objID: arrID}
}

func (o ScImmutableMap) GetUint8(key MapKey) ScImmutableUint8 {
	return ScImmutableUint8{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetUint8Array(key MapKey) ScImmutableUint8Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT8|TYPE_ARRAY)
	return ScImmutableUint8Array{objID: arrID}
}

func (o ScImmutableMap) GetUint16(key MapKey) ScImmutableUint16 {
	return ScImmutableUint16{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetUint16Array(key MapKey) ScImmutableUint16Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT16|TYPE_ARRAY)
	return ScImmutableUint16Array{objID: arrID}
}

func (o ScImmutableMap) GetUint32(key MapKey) ScImmutableUint32 {
	return ScImmutableUint32{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetUint32Array(key MapKey) ScImmutableUint32Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT32|TYPE_ARRAY)
	return ScImmutableUint32Array{objID: arrID}
}

func (o ScImmutableMap) GetUint64(key MapKey) ScImmutableUint64 {
	return ScImmutableUint64{objID: o.objID, keyID: key.KeyID()}
}

func (o ScImmutableMap) GetUint64Array(key MapKey) ScImmutableUint64Array {
	arrID := GetObjectID(o.objID, key.KeyID(), TYPE_INT64|TYPE_ARRAY)
	return ScImmutableUint64Array{objID: arrID}
}

func (o ScImmutableMap) MapID() int32 {
	return o.objID
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableMapArray struct {
	objID int32
}

func (o ScImmutableMapArray) GetMap(index int32) ScImmutableMap {
	mapID := GetObjectID(o.objID, Key32(index), TYPE_MAP)
	return ScImmutableMap{objID: mapID}
}

func (o ScImmutableMapArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableRequestID struct {
	objID int32
	keyID Key32
}

func NewScImmutableRequestID(objID int32, keyID Key32) ScImmutableRequestID {
	return ScImmutableRequestID{objID: objID, keyID: keyID}
}

func (o ScImmutableRequestID) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_REQUEST_ID)
}

func (o ScImmutableRequestID) String() string {
	return o.Value().String()
}

func (o ScImmutableRequestID) Value() ScRequestID {
	return NewScRequestIDFromBytes(GetBytes(o.objID, o.keyID, TYPE_REQUEST_ID))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableRequestIDArray struct {
	objID int32
}

func (o ScImmutableRequestIDArray) GetRequestID(index int32) ScImmutableRequestID {
	return ScImmutableRequestID{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableRequestIDArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableString struct {
	objID int32
	keyID Key32
}

func NewScImmutableString(objID int32, keyID Key32) ScImmutableString {
	return ScImmutableString{objID: objID, keyID: keyID}
}

func (o ScImmutableString) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_STRING)
}

func (o ScImmutableString) String() string {
	return o.Value()
}

func (o ScImmutableString) Value() string {
	bytes := GetBytes(o.objID, o.keyID, TYPE_STRING)
	if bytes == nil {
		return ""
	}
	return string(bytes)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableStringArray struct {
	objID int32
}

func (o ScImmutableStringArray) GetString(index int32) ScImmutableString {
	return ScImmutableString{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableStringArray) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint8 struct {
	objID int32
	keyID Key32
}

func NewScImmutableUint8(objID int32, keyID Key32) ScImmutableUint8 {
	return ScImmutableUint8{objID: objID, keyID: keyID}
}

func (o ScImmutableUint8) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT8)
}

func (o ScImmutableUint8) String() string {
	return strconv.FormatUint(uint64(o.Value()), 10)
}

func (o ScImmutableUint8) Value() uint8 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT8)
	return bytes[0]
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint8Array struct {
	objID int32
}

func (o ScImmutableUint8Array) GetUint8(index int32) ScImmutableUint8 {
	return ScImmutableUint8{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableUint8Array) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint16 struct {
	objID int32
	keyID Key32
}

func NewScImmutableUint16(objID int32, keyID Key32) ScImmutableUint16 {
	return ScImmutableUint16{objID: objID, keyID: keyID}
}

func (o ScImmutableUint16) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT16)
}

func (o ScImmutableUint16) String() string {
	return strconv.FormatUint(uint64(o.Value()), 10)
}

func (o ScImmutableUint16) Value() uint16 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT16)
	return binary.LittleEndian.Uint16(bytes)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint16Array struct {
	objID int32
}

func (o ScImmutableUint16Array) GetUint16(index int32) ScImmutableUint16 {
	return ScImmutableUint16{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableUint16Array) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint32 struct {
	objID int32
	keyID Key32
}

func NewScImmutableUint32(objID int32, keyID Key32) ScImmutableUint32 {
	return ScImmutableUint32{objID: objID, keyID: keyID}
}

func (o ScImmutableUint32) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT32)
}

func (o ScImmutableUint32) String() string {
	return strconv.FormatUint(uint64(o.Value()), 10)
}

func (o ScImmutableUint32) Value() uint32 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT32)
	return binary.LittleEndian.Uint32(bytes)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint32Array struct {
	objID int32
}

func (o ScImmutableUint32Array) GetUint32(index int32) ScImmutableUint32 {
	return ScImmutableUint32{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableUint32Array) Length() int32 {
	return GetLength(o.objID)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint64 struct {
	objID int32
	keyID Key32
}

func NewScImmutableUint64(objID int32, keyID Key32) ScImmutableUint64 {
	return ScImmutableUint64{objID: objID, keyID: keyID}
}

func (o ScImmutableUint64) Exists() bool {
	return Exists(o.objID, o.keyID, TYPE_INT64)
}

func (o ScImmutableUint64) String() string {
	return strconv.FormatUint(o.Value(), 10)
}

func (o ScImmutableUint64) Value() uint64 {
	bytes := GetBytes(o.objID, o.keyID, TYPE_INT64)
	return binary.LittleEndian.Uint64(bytes)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint64Array struct {
	objID int32
}

func (o ScImmutableUint64Array) GetUint64(index int32) ScImmutableUint64 {
	return ScImmutableUint64{objID: o.objID, keyID: Key32(index)}
}

func (o ScImmutableUint64Array) Length() int32 {
	return GetLength(o.objID)
}
