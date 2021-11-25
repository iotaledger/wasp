// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// immutable proxies to host objects

use std::convert::TryInto;

use crate::context::*;
use crate::hashtypes::*;
use crate::host::*;
use crate::keys::*;

// value proxy for immutable ScAddress in host container
pub struct ScImmutableAddress {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableAddress {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableAddress {
        ScImmutableAddress { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_ADDRESS)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> ScAddress {
        ScAddress::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_ADDRESS))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScAddress
pub struct ScImmutableAddressArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableAddressArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_address(&self, index: i32) -> ScImmutableAddress {
        ScImmutableAddress { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScAgentID in host container
pub struct ScImmutableAgentID {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableAgentID {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableAgentID {
        ScImmutableAgentID { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_AGENT_ID)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> ScAgentID {
        ScAgentID::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_AGENT_ID))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScAgentID
pub struct ScImmutableAgentIDArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableAgentIDArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_agent_id(&self, index: i32) -> ScImmutableAgentID {
        ScImmutableAgentID { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Bool in host container
pub struct ScImmutableBool {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableBool {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableBool {
        ScImmutableBool { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BOOL)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> bool {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_BOOL);
        bytes[0] != 0
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Bool
pub struct ScImmutableBoolArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableBoolArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_bool(&self, index: i32) -> ScImmutableBool {
        ScImmutableBool { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable bytes array in host container
pub struct ScImmutableBytes {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableBytes {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableBytes {
        ScImmutableBytes { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BYTES)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        base58_encode(&self.value())
    }

    // get value from host container
    pub fn value(&self) -> Vec<u8> {
        get_bytes(self.obj_id, self.key_id, TYPE_BYTES)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of byte array
pub struct ScImmutableBytesArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableBytesArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_bytes(&self, index: i32) -> ScImmutableBytes {
        ScImmutableBytes { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScChainID in host container
pub struct ScImmutableChainID {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableChainID {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableChainID {
        ScImmutableChainID { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_CHAIN_ID)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> ScChainID {
        ScChainID::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_CHAIN_ID))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScChainID
pub struct ScImmutableChainIDArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableChainIDArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_chain_id(&self, index: i32) -> ScImmutableChainID {
        ScImmutableChainID { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScColor in host container
pub struct ScImmutableColor {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableColor {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableColor {
        ScImmutableColor { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_COLOR)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> ScColor {
        ScColor::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_COLOR))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScColor
pub struct ScImmutableColorArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableColorArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_color(&self, index: i32) -> ScImmutableColor {
        ScImmutableColor { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScHash in host container
pub struct ScImmutableHash {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableHash {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableHash {
        ScImmutableHash { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_HASH)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> ScHash {
        ScHash::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_HASH))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScHash
pub struct ScImmutableHashArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableHashArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_hash(&self, index: i32) -> ScImmutableHash {
        ScImmutableHash { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScHname in host container
pub struct ScImmutableHname {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableHname {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableHname {
        ScImmutableHname { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_HNAME)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> ScHname {
        ScHname::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_HNAME))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScHname
pub struct ScImmutableHnameArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableHnameArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_hname(&self, index: i32) -> ScImmutableHname {
        ScImmutableHname { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int8 in host container
pub struct ScImmutableInt8 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableInt8 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableInt8 {
        ScImmutableInt8 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT8)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> i8 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT8);
        bytes[0] as i8
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Int8
pub struct ScImmutableInt8Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableInt8Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_int8(&self, index: i32) -> ScImmutableInt8 {
        ScImmutableInt8 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int16 in host container
pub struct ScImmutableInt16 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableInt16 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableInt16 {
        ScImmutableInt16 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT16)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> i16 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT16);
        i16::from_le_bytes(bytes.try_into().expect("invalid i16 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Int16
pub struct ScImmutableInt16Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableInt16Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_int16(&self, index: i32) -> ScImmutableInt16 {
        ScImmutableInt16 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int32 in host container
pub struct ScImmutableInt32 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableInt32 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableInt32 {
        ScImmutableInt32 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT32)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> i32 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT32);
        i32::from_le_bytes(bytes.try_into().expect("invalid i32 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Int32
pub struct ScImmutableInt32Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableInt32Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_int32(&self, index: i32) -> ScImmutableInt32 {
        ScImmutableInt32 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int64 in host container
pub struct ScImmutableInt64 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableInt64 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableInt64 {
        ScImmutableInt64 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT64)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> i64 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT64);
        i64::from_le_bytes(bytes.try_into().expect("invalid i64 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Int64
pub struct ScImmutableInt64Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableInt64Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_int64(&self, index: i32) -> ScImmutableInt64 {
        ScImmutableInt64 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// map proxy for immutable map
pub struct ScImmutableMap {
    pub(crate) obj_id: i32,
}

impl ScImmutableMap {
    pub fn call_func(&self, key_id: Key32, params: &[u8]) -> Vec<u8> {
        call_func(self.obj_id, key_id, params)
    }

    // get value proxy for immutable ScAddress field specified by key
    pub fn get_address<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableAddress {
        ScImmutableAddress { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableAddressArray specified by key
    pub fn get_address_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableAddressArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_ADDRESS | TYPE_ARRAY);
        ScImmutableAddressArray { obj_id: arr_id }
    }

    // get value proxy for immutable ScAgentID field specified by key
    pub fn get_agent_id<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableAgentID {
        ScImmutableAgentID { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableAgentIDArray specified by key
    pub fn get_agent_id_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableAgentIDArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_AGENT_ID | TYPE_ARRAY);
        ScImmutableAgentIDArray { obj_id: arr_id }
    }

    // get value proxy for immutable Bool field specified by key
    pub fn get_bool<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableBool {
        ScImmutableBool { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableBoolArray specified by key
    pub fn get_bool_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableBoolArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_BOOL | TYPE_ARRAY);
        ScImmutableBoolArray { obj_id: arr_id }
    }

    // get value proxy for immutable bytes array field specified by key
    pub fn get_bytes<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableBytes {
        ScImmutableBytes { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableBytesArray specified by key
    pub fn get_bytes_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableBytesArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_BYTES | TYPE_ARRAY);
        ScImmutableBytesArray { obj_id: arr_id }
    }

    // get value proxy for immutable ScChainID field specified by key
    pub fn get_chain_id<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableChainID {
        ScImmutableChainID { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableChainIDArray specified by key
    pub fn get_chain_id_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableChainIDArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_CHAIN_ID | TYPE_ARRAY);
        ScImmutableChainIDArray { obj_id: arr_id }
    }

    // get value proxy for immutable ScColor field specified by key
    pub fn get_color<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableColor {
        ScImmutableColor { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableColorArray specified by key
    pub fn get_color_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableColorArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_COLOR | TYPE_ARRAY);
        ScImmutableColorArray { obj_id: arr_id }
    }

    // get value proxy for immutable ScHash field specified by key
    pub fn get_hash<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableHash {
        ScImmutableHash { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableHashArray specified by key
    pub fn get_hash_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableHashArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_HASH | TYPE_ARRAY);
        ScImmutableHashArray { obj_id: arr_id }
    }

    // get value proxy for immutable ScHname field specified by key
    pub fn get_hname<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableHname {
        ScImmutableHname { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableHnameArray specified by key
    pub fn get_hname_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableHnameArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_HNAME | TYPE_ARRAY);
        ScImmutableHnameArray { obj_id: arr_id }
    }

    // get value proxy for immutable Int8 field specified by key
    pub fn get_int8<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt8 {
        ScImmutableInt8 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableInt8Array specified by key
    pub fn get_int8_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt8Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT8 | TYPE_ARRAY);
        ScImmutableInt8Array { obj_id: arr_id }
    }

    // get value proxy for immutable Int16 field specified by key
    pub fn get_int16<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt16 {
        ScImmutableInt16 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableInt16Array specified by key
    pub fn get_int16_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt16Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT16 | TYPE_ARRAY);
        ScImmutableInt16Array { obj_id: arr_id }
    }

    // get value proxy for immutable Int32 field specified by key
    pub fn get_int32<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt32 {
        ScImmutableInt32 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableInt32Array specified by key
    pub fn get_int32_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt32Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT32 | TYPE_ARRAY);
        ScImmutableInt32Array { obj_id: arr_id }
    }

    // get value proxy for immutable Int64 field specified by key
    pub fn get_int64<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt64 {
        ScImmutableInt64 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableInt64Array specified by key
    pub fn get_int64_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableInt64Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT64 | TYPE_ARRAY);
        ScImmutableInt64Array { obj_id: arr_id }
    }

    // get map proxy for ScImmutableMap specified by key
    pub fn get_map<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableMap {
        let map_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_MAP);
        ScImmutableMap { obj_id: map_id }
    }

    // get array proxy for ScImmutableMapArray specified by key
    pub fn get_map_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableMapArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_MAP | TYPE_ARRAY);
        ScImmutableMapArray { obj_id: arr_id }
    }

    // get value proxy for immutable ScRequestID field specified by key
    pub fn get_request_id<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableRequestID {
        ScImmutableRequestID { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableRequestIDArray specified by key
    pub fn get_request_id_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableRequestIDArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_REQUEST_ID | TYPE_ARRAY);
        ScImmutableRequestIDArray { obj_id: arr_id }
    }

    // get value proxy for immutable UTF-8 text string field specified by key
    pub fn get_string<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableString {
        ScImmutableString { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableStringArray specified by key
    pub fn get_string_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableStringArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_STRING | TYPE_ARRAY);
        ScImmutableStringArray { obj_id: arr_id }
    }

    // get value proxy for immutable Uint8 field specified by key
    pub fn get_uint8<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint8 {
        ScImmutableUint8 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableUint8Array specified by key
    pub fn get_uint8_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint8Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT8 | TYPE_ARRAY);
        ScImmutableUint8Array { obj_id: arr_id }
    }

    // get value proxy for immutable Uint16 field specified by key
    pub fn get_uint16<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint16 {
        ScImmutableUint16 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableUint16Array specified by key
    pub fn get_uint16_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint16Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT16 | TYPE_ARRAY);
        ScImmutableUint16Array { obj_id: arr_id }
    }

    // get value proxy for immutable Uint32 field specified by key
    pub fn get_uint32<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint32 {
        ScImmutableUint32 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableUint32Array specified by key
    pub fn get_uint32_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint32Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT32 | TYPE_ARRAY);
        ScImmutableUint32Array { obj_id: arr_id }
    }

    // get value proxy for immutable Uint64 field specified by key
    pub fn get_uint64<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint64 {
        ScImmutableUint64 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScImmutableUint64Array specified by key
    pub fn get_uint64_array<T: MapKey + ?Sized>(&self, key: &T) -> ScImmutableUint64Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT64 | TYPE_ARRAY);
        ScImmutableUint64Array { obj_id: arr_id }
    }

    pub fn map_id(&self) -> i32 {
        self.obj_id
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of maps
pub struct ScImmutableMapArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableMapArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_map(&self, index: i32) -> ScImmutableMap {
        let map_id = get_object_id(self.obj_id, Key32(index), TYPE_MAP);
        ScImmutableMap { obj_id: map_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// value proxy for immutable ScRequestID in host container
pub struct ScImmutableRequestID {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableRequestID {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableRequestID {
        ScImmutableRequestID { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_REQUEST_ID)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> ScRequestID {
        ScRequestID::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_REQUEST_ID))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScRequestID
pub struct ScImmutableRequestIDArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableRequestIDArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_request_id(&self, index: i32) -> ScImmutableRequestID {
        ScImmutableRequestID { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable UTF-8 text string in host container
pub struct ScImmutableString {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableString {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableString {
        ScImmutableString { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_STRING)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value()
    }

    // get value from host container
    pub fn value(&self) -> String {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_STRING);
        unsafe { String::from_utf8_unchecked(bytes) }
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of UTF-8 text string
pub struct ScImmutableStringArray {
    pub(crate) obj_id: i32,
}

impl ScImmutableStringArray {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_string(&self, index: i32) -> ScImmutableString {
        ScImmutableString { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint8 in host container
pub struct ScImmutableUint8 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableUint8 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableUint8 {
        ScImmutableUint8 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT8)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> u8 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT8);
        bytes[0]
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Uint8
pub struct ScImmutableUint8Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableUint8Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_uint8(&self, index: i32) -> ScImmutableUint8 {
        ScImmutableUint8 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint16 in host container
pub struct ScImmutableUint16 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableUint16 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableUint16 {
        ScImmutableUint16 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT16)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> u16 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT16);
        u16::from_le_bytes(bytes.try_into().expect("invalid u16 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Uint16
pub struct ScImmutableUint16Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableUint16Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_uint16(&self, index: i32) -> ScImmutableUint16 {
        ScImmutableUint16 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint32 in host container
pub struct ScImmutableUint32 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableUint32 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableUint32 {
        ScImmutableUint32 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT32)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> u32 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT32);
        u32::from_le_bytes(bytes.try_into().expect("invalid u32 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Uint32
pub struct ScImmutableUint32Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableUint32Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_uint32(&self, index: i32) -> ScImmutableUint32 {
        ScImmutableUint32 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint64 in host container
pub struct ScImmutableUint64 {
    obj_id: i32,
    key_id: Key32,
}

impl ScImmutableUint64 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScImmutableUint64 {
        ScImmutableUint64 { obj_id, key_id }
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT64)
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // get value from host container
    pub fn value(&self) -> u64 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT64);
        u64::from_le_bytes(bytes.try_into().expect("invalid u64 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of Uint64
pub struct ScImmutableUint64Array {
    pub(crate) obj_id: i32,
}

impl ScImmutableUint64Array {
    // get value proxy for item at index, index can be 0..length()-1
    pub fn get_uint64(&self, index: i32) -> ScImmutableUint64 {
        ScImmutableUint64 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}
