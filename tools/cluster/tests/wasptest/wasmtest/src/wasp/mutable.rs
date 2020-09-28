// types encapsulating mutable host objects
// ScMutableInt          : refers to mutable integer on host
// ScMutableMap          : refers to mutable map of mutable values on host
// ScMutableString       : refers to mutable string on host
// ScMutableIntArray     : refers to mutable array of mutable integers on host
// ScMutableMapArray     : refers to mutable array of mutable maps on host
// ScMutableStringArray  : refers to mutable array of mutable strings on host

use super::host::get_int;
use super::host::get_key;
use super::host::get_object;
use super::host::get_string;
use super::host::set_int;
use super::host::set_string;
use super::host::ScType;
use super::immutable::ScImmutableIntArray;
use super::immutable::ScImmutableMap;
use super::immutable::ScImmutableMapArray;
use super::immutable::ScImmutableStringArray;
use super::keys::key_length;

#[derive(Copy, Clone)]
pub struct ScMutableInt {
    obj_id: i32,
    key_id: i32,
}

impl ScMutableInt {
    pub fn set_value(&self, val: i64) {
        set_int(self.obj_id, self.key_id, val);
    }

    pub fn value(&self) -> i64 {
        get_int(self.obj_id, self.key_id)
    }
}

#[derive(Copy, Clone)]
pub struct ScMutableIntArray {
    obj_id: i32
}

impl ScMutableIntArray {
    pub fn clear(&self) {
        set_int(self.obj_id, key_length(), 0);
    }

    // index 0..length(), when length() a new one is appended
    pub fn get_int(&self, index: i32) -> ScMutableInt {
        ScMutableInt { obj_id: self.obj_id, key_id: index }
    }

    pub fn immutable(&self) -> ScImmutableIntArray {
        ScImmutableIntArray::new(self.obj_id)
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}

#[derive(Copy, Clone)]
pub struct ScMutableMap {
    obj_id: i32
}

impl ScMutableMap {
    pub(crate) fn new_root_map() -> ScMutableMap {
        ScMutableMap { obj_id: 1 }
    }

    pub fn clear(&self) {
        set_int(self.obj_id, key_length(), 0);
    }

    pub fn get_int(&self, key: &str) -> ScMutableInt {
        ScMutableInt { obj_id: self.obj_id, key_id: get_key(key) }
    }

    pub fn get_int_array(&self, key: &str) -> ScMutableIntArray {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeIntArray);
        ScMutableIntArray { obj_id }
    }

    pub fn get_map(&self, key: &str) -> ScMutableMap {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeMap);
        ScMutableMap { obj_id }
    }

    pub fn get_map_array(&self, key: &str) -> ScMutableMapArray {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeMapArray);
        ScMutableMapArray { obj_id }
    }

    pub fn get_string(&self, key: &str) -> ScMutableString {
        ScMutableString { obj_id: self.obj_id, key_id: get_key(key) }
    }

    pub fn get_string_array(&self, key: &str) -> ScMutableStringArray {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeStringArray);
        ScMutableStringArray { obj_id }
    }

    pub fn immutable(&self) -> ScImmutableMap {
        ScImmutableMap::new(self.obj_id)
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}

#[derive(Copy, Clone)]
pub struct ScMutableMapArray {
    obj_id: i32
}

impl ScMutableMapArray {
    pub fn clear(&self) {
        set_int(self.obj_id, key_length(), 0);
    }

    // index 0..length(), when length() a new one is appended
    pub fn get_map(&self, index: i32) -> ScMutableMap {
        let obj_id = get_object(self.obj_id, index, ScType::TypeMap);
        ScMutableMap { obj_id }
    }

    pub fn immutable(&self) -> ScImmutableMapArray {
        ScImmutableMapArray::new(self.obj_id)
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}

#[derive(Copy, Clone)]
pub struct ScMutableString {
    obj_id: i32,
    key_id: i32,
}

impl ScMutableString {
    pub fn set_value(&self, val: &str) {
        set_string(self.obj_id, self.key_id, val);
    }

    pub fn value(&self) -> String {
        String::from(get_string(self.obj_id, self.key_id))
    }
}

#[derive(Copy, Clone)]
pub struct ScMutableStringArray {
    obj_id: i32
}

impl ScMutableStringArray {
    pub fn clear(&self) {
        set_int(self.obj_id, key_length(), 0);
    }

    // index 0..length(), when length() a new one is appended
    pub fn get_string(&self, index: i32) -> ScMutableString {
        ScMutableString { obj_id: self.obj_id, key_id: index }
    }

    pub fn immutable(&self) -> ScImmutableStringArray {
        ScImmutableStringArray::new(self.obj_id)
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}
