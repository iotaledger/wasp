// types encapsulating immutable host objects
// ScImmutableInt          : refers to immutable integer on host
// ScImmutableMap          : refers to immutable map of immutable values on host
// ScImmutableString       : refers to immutable string on host
// ScImmutableIntArray     : refers to immutable array of immutable integers on host
// ScImmutableMapArray     : refers to immutable array of immutable maps on host
// ScImmutableStringArray  : refers to immutable array of immutable strings on host

use super::host::get_int;
use super::host::get_key;
use super::host::get_object;
use super::host::get_string;
use super::host::ScType;
use super::keys::key_length;

#[derive(Copy, Clone)]
pub struct ScImmutableInt {
    obj_id: i32,
    key_id: i32,
}

impl ScImmutableInt {
    pub fn value(&self) -> i64 {
        get_int(self.obj_id, self.key_id)
    }
}

#[derive(Copy, Clone)]
pub struct ScImmutableIntArray {
    obj_id: i32
}

impl ScImmutableIntArray {
    pub(crate) fn new(obj_id: i32) -> ScImmutableIntArray {
        ScImmutableIntArray { obj_id }
    }

    // index 0..length(), exclusive
    pub fn get_int(&self, index: i32) -> ScImmutableInt {
        ScImmutableInt { obj_id: self.obj_id, key_id: index }
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}

#[derive(Copy, Clone)]
pub struct ScImmutableMap {
    obj_id: i32
}

impl ScImmutableMap {
    pub(crate) fn new(obj_id: i32) -> ScImmutableMap {
        ScImmutableMap { obj_id }
    }

    pub fn get_int(&self, key: &str) -> ScImmutableInt {
        ScImmutableInt { obj_id: self.obj_id, key_id: get_key(key) }
    }

    pub fn get_int_array(&self, key: &str) -> ScImmutableIntArray {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeIntArray);
        ScImmutableIntArray { obj_id }
    }

    pub fn get_map(&self, key: &str) -> ScImmutableMap {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeMap);
        ScImmutableMap { obj_id }
    }

    pub fn get_map_array(&self, key: &str) -> ScImmutableMapArray {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeMapArray);
        ScImmutableMapArray { obj_id }
    }

    pub fn get_string(&self, key: &str) -> ScImmutableString {
        ScImmutableString { obj_id: self.obj_id, key_id: get_key(key) }
    }

    pub fn get_string_array(&self, key: &str) -> ScImmutableStringArray {
        let obj_id = get_object(self.obj_id, get_key(key), ScType::TypeStringArray);
        ScImmutableStringArray { obj_id }
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}

#[derive(Copy, Clone)]
pub struct ScImmutableMapArray {
    obj_id: i32
}

impl ScImmutableMapArray {
    pub(crate) fn new(obj_id: i32) -> ScImmutableMapArray {
        ScImmutableMapArray { obj_id }
    }

    // index 0..length(), exclusive
    pub fn get_map(&self, index: i32) -> ScImmutableMap {
        let obj_id = get_object(self.obj_id, index, ScType::TypeMap);
        ScImmutableMap { obj_id }
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}

#[derive(Copy, Clone)]
pub struct ScImmutableString {
    obj_id: i32,
    key_id: i32,
}

impl ScImmutableString {
    pub fn value(&self) -> String {
        String::from(get_string(self.obj_id, self.key_id))
    }
}

#[derive(Copy, Clone)]
pub struct ScImmutableStringArray {
    obj_id: i32
}

impl ScImmutableStringArray {
    pub(crate) fn new(obj_id: i32) -> ScImmutableStringArray {
        ScImmutableStringArray { obj_id }
    }

    // index 0..length(), exclusive
    pub fn get_string(&self, index: i32) -> ScImmutableString {
        ScImmutableString { obj_id: self.obj_id, key_id: index }
    }

    pub fn length(&self) -> i32 {
        get_int(self.obj_id, key_length()) as i32
    }
}
