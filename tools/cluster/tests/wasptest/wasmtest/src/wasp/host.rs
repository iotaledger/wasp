// all enum values should exactly match the counterpart values on the host!
#[derive(Copy, Clone)]
#[repr(C)]
pub enum ScType {
    TypeInt,
    TypeIntArray,
    TypeMap,
    TypeMapArray,
    TypeString,
    TypeStringArray,
}

// all token colors must be encoded as a 64-byte hex string,
// except for the following two special cases:
// default color, encoded as "iota" (COLOR_IOTA)
// new color, encoded as "new" (COLOR_NEW)

// any host function that gets called once the current request has
// entered an error state will immediately return without action.
// Any return value will be zero or empty string in that case
#[link(wasm_import_module = "waspRust")]
#[no_mangle]
extern {
    pub fn hostGetInt(obj_id: i32, key_id: i32) -> i64;
    pub fn hostGetKey(key: &str) -> i32;
    pub fn hostGetObject(obj_id: i32, key_id: i32, type_id: ScType) -> i32;
    pub fn hostGetString(obj_id: i32, key_id: i32) -> &'static str;
    pub fn hostSetInt(obj_id: i32, key_id: i32, value: i64);
    pub fn hostSetString(obj_id: i32, key_id: i32, value: &str);
}

pub fn get_int(obj_id: i32, key_id: i32) -> i64 {
    unsafe {
        hostGetInt(obj_id, key_id)
    }
}

pub fn get_key(key: &str) -> i32 {
    unsafe {
        hostGetKey(key)
    }
}

pub fn get_object(obj_id: i32, key_id: i32, type_id: ScType) -> i32 {
    unsafe {
        hostGetObject(obj_id, key_id, type_id)
    }
}

pub fn get_string(obj_id: i32, key_id: i32) -> &'static str {
    unsafe {
        hostGetString(obj_id, key_id)
    }
}

pub fn set_int(obj_id: i32, key_id: i32, value: i64) {
    unsafe {
        hostSetInt(obj_id, key_id, value)
    }
}

pub fn set_string(obj_id: i32, key_id: i32, value: &str) {
    unsafe {
        hostSetString(obj_id, key_id, value)
    }
}
