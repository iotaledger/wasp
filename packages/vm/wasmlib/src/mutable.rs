// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// mutable proxies to host objects

use std::convert::TryInto;

use crate::context::*;
use crate::hashtypes::*;
use crate::host::*;
use crate::immutable::*;
use crate::keys::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScAddress in host container
pub struct ScMutableAddress {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableAddress {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableAddress {
        ScMutableAddress { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_ADDRESS);
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_ADDRESS)
    }

    // set value in host container
    pub fn set_value(&self, val: &ScAddress) {
        set_bytes(self.obj_id, self.key_id, TYPE_ADDRESS, val.to_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> ScAddress {
        ScAddress::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_ADDRESS))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScAddress
pub struct ScMutableAddressArray {
    pub(crate) obj_id: i32,
}

impl ScMutableAddressArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_address(&self, index: i32) -> ScMutableAddress {
        ScMutableAddress { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableAddressArray {
        ScImmutableAddressArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScAgentID in host container
pub struct ScMutableAgentID {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableAgentID {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableAgentID {
        ScMutableAgentID { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_AGENT_ID)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_AGENT_ID)
    }

    // set value in host container
    pub fn set_value(&self, val: &ScAgentID) {
        set_bytes(self.obj_id, self.key_id, TYPE_AGENT_ID, val.to_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> ScAgentID {
        ScAgentID::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_AGENT_ID))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScAgentID
pub struct ScMutableAgentIDArray {
    pub(crate) obj_id: i32,
}

impl ScMutableAgentIDArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_agent_id(&self, index: i32) -> ScMutableAgentID {
        ScMutableAgentID { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableAgentIDArray {
        ScImmutableAgentIDArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Bool in host container
pub struct ScMutableBool {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableBool {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableBool {
        ScMutableBool { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_BOOL)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BOOL)
    }

    // set value in host container
    pub fn set_value(&self, val: bool) {
        let bytes = [val as u8];
        set_bytes(self.obj_id, self.key_id, TYPE_BOOL, &bytes);
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> bool {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_BOOL);
        bytes[0] != 0
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Bool
pub struct ScMutableBoolArray {
    pub(crate) obj_id: i32,
}

impl ScMutableBoolArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_bool(&self, index: i32) -> ScMutableBool {
        ScMutableBool { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableBoolArray {
        ScImmutableBoolArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable bytes array in host container
pub struct ScMutableBytes {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableBytes {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableBytes {
        ScMutableBytes { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_BYTES)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BYTES)
    }

    // set value in host container
    pub fn set_value(&self, val: &[u8]) {
        set_bytes(self.obj_id, self.key_id, TYPE_BYTES, val);
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        base58_encode(&self.value())
    }

    // retrieve value from host container
    pub fn value(&self) -> Vec<u8> {
        get_bytes(self.obj_id, self.key_id, TYPE_BYTES)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of byte array
pub struct ScMutableBytesArray {
    pub(crate) obj_id: i32,
}

impl ScMutableBytesArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_bytes(&self, index: i32) -> ScMutableBytes {
        ScMutableBytes { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableBytesArray {
        ScImmutableBytesArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScChainID in host container
pub struct ScMutableChainID {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableChainID {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableChainID {
        ScMutableChainID { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_CHAIN_ID)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_CHAIN_ID)
    }

    // set value in host container
    pub fn set_value(&self, val: &ScChainID) {
        set_bytes(self.obj_id, self.key_id, TYPE_CHAIN_ID, val.to_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> ScChainID {
        ScChainID::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_CHAIN_ID))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScChainID
pub struct ScMutableChainIDArray {
    pub(crate) obj_id: i32,
}

impl ScMutableChainIDArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_chain_id(&self, index: i32) -> ScMutableChainID {
        ScMutableChainID { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableChainIDArray {
        ScImmutableChainIDArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScColor in host container
pub struct ScMutableColor {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableColor {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableColor {
        ScMutableColor { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_COLOR)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_COLOR)
    }

    // set value in host container
    pub fn set_value(&self, val: &ScColor) {
        set_bytes(self.obj_id, self.key_id, TYPE_COLOR, val.to_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> ScColor {
        ScColor::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_COLOR))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScColor
pub struct ScMutableColorArray {
    pub(crate) obj_id: i32,
}

impl ScMutableColorArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_color(&self, index: i32) -> ScMutableColor {
        ScMutableColor { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableColorArray {
        ScImmutableColorArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHash in host container
pub struct ScMutableHash {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableHash {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableHash {
        ScMutableHash { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_HASH)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_HASH)
    }

    // set value in host container
    pub fn set_value(&self, val: &ScHash) {
        set_bytes(self.obj_id, self.key_id, TYPE_HASH, val.to_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> ScHash {
        ScHash::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_HASH))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScHash
pub struct ScMutableHashArray {
    pub(crate) obj_id: i32,
}

impl ScMutableHashArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_hash(&self, index: i32) -> ScMutableHash {
        ScMutableHash { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableHashArray {
        ScImmutableHashArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHname in host container
pub struct ScMutableHname {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableHname {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableHname {
        ScMutableHname { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_HNAME)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_HNAME)
    }

    // set value in host container
    pub fn set_value(&self, val: ScHname) {
        set_bytes(self.obj_id, self.key_id, TYPE_HNAME, &val.to_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> ScHname {
        ScHname::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_HNAME))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScHname
pub struct ScMutableHnameArray {
    pub(crate) obj_id: i32,
}

impl ScMutableHnameArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_hname(&self, index: i32) -> ScMutableHname {
        ScMutableHname { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableHnameArray {
        ScImmutableHnameArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int8 in host container
pub struct ScMutableInt8 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableInt8 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableInt8 {
        ScMutableInt8 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT8)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT8)
    }

    // set value in host container
    pub fn set_value(&self, val: i8) {
        let bytes = [val as u8];
        set_bytes(self.obj_id, self.key_id, TYPE_INT8, &bytes);
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> i8 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT8);
        bytes[0] as i8
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int8
pub struct ScMutableInt8Array {
    pub(crate) obj_id: i32,
}

impl ScMutableInt8Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_int8(&self, index: i32) -> ScMutableInt8 {
        ScMutableInt8 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableInt8Array {
        ScImmutableInt8Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int16 in host container
pub struct ScMutableInt16 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableInt16 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableInt16 {
        ScMutableInt16 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT16)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT16)
    }

    // set value in host container
    pub fn set_value(&self, val: i16) {
        set_bytes(self.obj_id, self.key_id, TYPE_INT16, &val.to_le_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> i16 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT16);
        i16::from_le_bytes(bytes.try_into().expect("invalid i16 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int16
pub struct ScMutableInt16Array {
    pub(crate) obj_id: i32,
}

impl ScMutableInt16Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_int16(&self, index: i32) -> ScMutableInt16 {
        ScMutableInt16 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableInt16Array {
        ScImmutableInt16Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int32 in host container
pub struct ScMutableInt32 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableInt32 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableInt32 {
        ScMutableInt32 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT32)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT32)
    }

    // set value in host container
    pub fn set_value(&self, val: i32) {
        set_bytes(self.obj_id, self.key_id, TYPE_INT32, &val.to_le_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> i32 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT32);
        i32::from_le_bytes(bytes.try_into().expect("invalid i32 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int32
pub struct ScMutableInt32Array {
    pub(crate) obj_id: i32,
}

impl ScMutableInt32Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_int32(&self, index: i32) -> ScMutableInt32 {
        ScMutableInt32 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableInt32Array {
        ScImmutableInt32Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int64 in host container
pub struct ScMutableInt64 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableInt64 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableInt64 {
        ScMutableInt64 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT64)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT64)
    }

    // set value in host container
    pub fn set_value(&self, val: i64) {
        set_bytes(self.obj_id, self.key_id, TYPE_INT64, &val.to_le_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> i64 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT64);
        i64::from_le_bytes(bytes.try_into().expect("invalid i64 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int64
pub struct ScMutableInt64Array {
    pub(crate) obj_id: i32,
}

impl ScMutableInt64Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_int64(&self, index: i32) -> ScMutableInt64 {
        ScMutableInt64 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableInt64Array {
        ScImmutableInt64Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// map proxy for mutable map
#[derive(Clone, Copy)]
pub struct ScMutableMap {
    pub(crate) obj_id: i32,
}

impl ScMutableMap {
    pub fn call_func(&self, key_id: Key32, params: &[u8]) -> Vec<u8> {
        call_func(self.obj_id, key_id, params)
    }

    // construct a new map on the host and return a map proxy for it
    pub fn new() -> ScMutableMap {
        let maps = ROOT.get_map_array(&KEY_MAPS);
        maps.get_map(maps.length())
    }

    // empty the map
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for mutable ScAddress field specified by key
    pub fn get_address<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableAddress {
        ScMutableAddress { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableAddressArray specified by key
    pub fn get_address_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableAddressArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_ADDRESS | TYPE_ARRAY);
        ScMutableAddressArray { obj_id: arr_id }
    }

    // get value proxy for mutable ScAgentID field specified by key
    pub fn get_agent_id<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableAgentID {
        ScMutableAgentID { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableAgentIDArray specified by key
    pub fn get_agent_id_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableAgentIDArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_AGENT_ID | TYPE_ARRAY);
        ScMutableAgentIDArray { obj_id: arr_id }
    }

    // get value proxy for mutable Bool field specified by key
    pub fn get_bool<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableBool {
        ScMutableBool { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableBoolArray specified by key
    pub fn get_bool_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableBoolArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_BOOL | TYPE_ARRAY);
        ScMutableBoolArray { obj_id: arr_id }
    }

    // get value proxy for mutable bytes array field specified by key
    pub fn get_bytes<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableBytes {
        ScMutableBytes { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableBytesArray specified by key
    pub fn get_bytes_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableBytesArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_BYTES | TYPE_ARRAY);
        ScMutableBytesArray { obj_id: arr_id }
    }

    // get value proxy for mutable ScChainID field specified by key
    pub fn get_chain_id<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableChainID {
        ScMutableChainID { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableChainIDArray specified by key
    pub fn get_chain_id_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableChainIDArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_CHAIN_ID | TYPE_ARRAY);
        ScMutableChainIDArray { obj_id: arr_id }
    }

    // get value proxy for mutable ScColor field specified by key
    pub fn get_color<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableColor {
        ScMutableColor { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableColorArray specified by key
    pub fn get_color_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableColorArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_COLOR | TYPE_ARRAY);
        ScMutableColorArray { obj_id: arr_id }
    }

    // get value proxy for mutable ScHash field specified by key
    pub fn get_hash<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableHash {
        ScMutableHash { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableHashArray specified by key
    pub fn get_hash_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableHashArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_HASH | TYPE_ARRAY);
        ScMutableHashArray { obj_id: arr_id }
    }

    // get value proxy for mutable ScHname field specified by key
    pub fn get_hname<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableHname {
        ScMutableHname { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableHnameArray specified by key
    pub fn get_hname_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableHnameArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_HNAME | TYPE_ARRAY);
        ScMutableHnameArray { obj_id: arr_id }
    }

    // get value proxy for mutable Int8 field specified by key
    pub fn get_int8<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt8 {
        ScMutableInt8 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableInt8Array specified by key
    pub fn get_int8_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt8Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT8 | TYPE_ARRAY);
        ScMutableInt8Array { obj_id: arr_id }
    }

    // get value proxy for mutable Int16 field specified by key
    pub fn get_int16<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt16 {
        ScMutableInt16 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableInt16Array specified by key
    pub fn get_int16_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt16Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT16 | TYPE_ARRAY);
        ScMutableInt16Array { obj_id: arr_id }
    }

    // get value proxy for mutable Int32 field specified by key
    pub fn get_int32<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt32 {
        ScMutableInt32 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableInt32Array specified by key
    pub fn get_int32_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt32Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT32 | TYPE_ARRAY);
        ScMutableInt32Array { obj_id: arr_id }
    }

    // get value proxy for mutable Int64 field specified by key
    pub fn get_int64<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt64 {
        ScMutableInt64 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableInt64Array specified by key
    pub fn get_int64_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableInt64Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT64 | TYPE_ARRAY);
        ScMutableInt64Array { obj_id: arr_id }
    }

    // get map proxy for ScMutableMap specified by key
    pub fn get_map<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableMap {
        let map_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_MAP);
        ScMutableMap { obj_id: map_id }
    }

    // get array proxy for ScMutableMapArray specified by key
    pub fn get_map_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableMapArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_MAP | TYPE_ARRAY);
        ScMutableMapArray { obj_id: arr_id }
    }

    // get value proxy for mutable ScRequestID field specified by key
    pub fn get_request_id<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableRequestID {
        ScMutableRequestID { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableRequestIDArray specified by key
    pub fn get_request_id_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableRequestIDArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_REQUEST_ID | TYPE_ARRAY);
        ScMutableRequestIDArray { obj_id: arr_id }
    }

    // get value proxy for mutable UTF-8 text string field specified by key
    pub fn get_string<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableString {
        ScMutableString { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableStringArray specified by key
    pub fn get_string_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableStringArray {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_STRING | TYPE_ARRAY);
        ScMutableStringArray { obj_id: arr_id }
    }

    // get value proxy for mutable Uint8 field specified by key
    pub fn get_uint8<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint8 {
        ScMutableUint8 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableUint8Array specified by key
    pub fn get_uint8_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint8Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT8 | TYPE_ARRAY);
        ScMutableUint8Array { obj_id: arr_id }
    }

    // get value proxy for mutable Uint16 field specified by key
    pub fn get_uint16<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint16 {
        ScMutableUint16 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableUint16Array specified by key
    pub fn get_uint16_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint16Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT16 | TYPE_ARRAY);
        ScMutableUint16Array { obj_id: arr_id }
    }

    // get value proxy for mutable Uint32 field specified by key
    pub fn get_uint32<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint32 {
        ScMutableUint32 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableUint32Array specified by key
    pub fn get_uint32_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint32Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT32 | TYPE_ARRAY);
        ScMutableUint32Array { obj_id: arr_id }
    }

    // get value proxy for mutable Uint64 field specified by key
    pub fn get_uint64<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint64 {
        ScMutableUint64 { obj_id: self.obj_id, key_id: key.get_key_id() }
    }

    // get array proxy for ScMutableUint64Array specified by key
    pub fn get_uint64_array<T: MapKey + ?Sized>(&self, key: &T) -> ScMutableUint64Array {
        let arr_id = get_object_id(self.obj_id, key.get_key_id(), TYPE_INT64 | TYPE_ARRAY);
        ScMutableUint64Array { obj_id: arr_id }
    }

    // get immutable version of map proxy
    pub fn immutable(&self) -> ScImmutableMap {
        ScImmutableMap { obj_id: self.obj_id }
    }

    pub fn map_id(&self) -> i32 {
        self.obj_id
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of maps
pub struct ScMutableMapArray {
    pub(crate) obj_id: i32,
}

impl ScMutableMapArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_map(&self, index: i32) -> ScMutableMap {
        let map_id = get_object_id(self.obj_id, Key32(index), TYPE_MAP);
        ScMutableMap { obj_id: map_id }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableMapArray {
        ScImmutableMapArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScRequestID in host container
pub struct ScMutableRequestID {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableRequestID {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableRequestID {
        ScMutableRequestID { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_REQUEST_ID)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_REQUEST_ID)
    }

    // set value in host container
    pub fn set_value(&self, val: &ScRequestID) {
        set_bytes(self.obj_id, self.key_id, TYPE_REQUEST_ID, val.to_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> ScRequestID {
        ScRequestID::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_REQUEST_ID))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScRequestID
pub struct ScMutableRequestIDArray {
    pub(crate) obj_id: i32,
}

impl ScMutableRequestIDArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_request_id(&self, index: i32) -> ScMutableRequestID {
        ScMutableRequestID { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableRequestIDArray {
        ScImmutableRequestIDArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable UTF-8 text string in host container
pub struct ScMutableString {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableString {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableString {
        ScMutableString { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_STRING)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_STRING)
    }

    // set value in host container
    pub fn set_value(&self, val: &str) {
        set_bytes(self.obj_id, self.key_id, TYPE_STRING, val.as_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value()
    }

    // retrieve value from host container
    pub fn value(&self) -> String {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_STRING);
        unsafe { String::from_utf8_unchecked(bytes) }
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of UTF-8 text string
pub struct ScMutableStringArray {
    pub(crate) obj_id: i32,
}

impl ScMutableStringArray {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_string(&self, index: i32) -> ScMutableString {
        ScMutableString { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableStringArray {
        ScImmutableStringArray { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint8 in host container
pub struct ScMutableUint8 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableUint8 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableUint8 {
        ScMutableUint8 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT8)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT8)
    }

    // set value in host container
    pub fn set_value(&self, val: u8) {
        let bytes = [val];
        set_bytes(self.obj_id, self.key_id, TYPE_INT8, &bytes);
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> u8 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT8);
        bytes[0]
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Uint8
pub struct ScMutableUint8Array {
    pub(crate) obj_id: i32,
}

impl ScMutableUint8Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_uint8(&self, index: i32) -> ScMutableUint8 {
        ScMutableUint8 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableUint8Array {
        ScImmutableUint8Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint16 in host container
pub struct ScMutableUint16 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableUint16 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableUint16 {
        ScMutableUint16 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT16)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT16)
    }

    // set value in host container
    pub fn set_value(&self, val: u16) {
        set_bytes(self.obj_id, self.key_id, TYPE_INT16, &val.to_le_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> u16 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT16);
        u16::from_le_bytes(bytes.try_into().expect("invalid u16 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Uint16
pub struct ScMutableUint16Array {
    pub(crate) obj_id: i32,
}

impl ScMutableUint16Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_uint16(&self, index: i32) -> ScMutableUint16 {
        ScMutableUint16 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableUint16Array {
        ScImmutableUint16Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint32 in host container
pub struct ScMutableUint32 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableUint32 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableUint32 {
        ScMutableUint32 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT32)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT32)
    }

    // set value in host container
    pub fn set_value(&self, val: u32) {
        set_bytes(self.obj_id, self.key_id, TYPE_INT32, &val.to_le_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> u32 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT32);
        u32::from_le_bytes(bytes.try_into().expect("invalid u32 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Uint32
pub struct ScMutableUint32Array {
    pub(crate) obj_id: i32,
}

impl ScMutableUint32Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_uint32(&self, index: i32) -> ScMutableUint32 {
        ScMutableUint32 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableUint32Array {
        ScImmutableUint32Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint64 in host container
pub struct ScMutableUint64 {
    obj_id: i32,
    key_id: Key32,
}

impl ScMutableUint64 {
    pub fn new(obj_id: i32, key_id: Key32) -> ScMutableUint64 {
        ScMutableUint64 { obj_id, key_id }
    }

    // delete value from host container
    pub fn delete(&self)  {
        del_key(self.obj_id, self.key_id, TYPE_INT64)
    }

    // check if value exists in host container
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_INT64)
    }

    // set value in host container
    pub fn set_value(&self, val: u64) {
        set_bytes(self.obj_id, self.key_id, TYPE_INT64, &val.to_le_bytes());
    }

    // human-readable string representation
    pub fn to_string(&self) -> String {
        self.value().to_string()
    }

    // retrieve value from host container
    pub fn value(&self) -> u64 {
        let bytes = get_bytes(self.obj_id, self.key_id, TYPE_INT64);
        u64::from_le_bytes(bytes.try_into().expect("invalid ui64 length"))
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint64
pub struct ScMutableUint64Array {
    pub(crate) obj_id: i32,
}

impl ScMutableUint64Array {
    // empty the array
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    pub fn get_uint64(&self, index: i32) -> ScMutableUint64 {
        ScMutableUint64 { obj_id: self.obj_id, key_id: Key32(index) }
    }

    // get immutable version of array proxy
    pub fn immutable(&self) -> ScImmutableUint64Array {
        ScImmutableUint64Array { obj_id: self.obj_id }
    }

    // number of items in array
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }
}
