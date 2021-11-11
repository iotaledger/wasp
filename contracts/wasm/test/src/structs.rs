// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

#![allow(dead_code)]

use wasmlib::*;
use wasmlib::host::*;

pub struct TestStruct {
    pub description : String, 
    pub id          : i32, 
}

impl TestStruct {
    pub fn from_bytes(bytes: &[u8]) -> TestStruct {
        let mut decode = BytesDecoder::new(bytes);
        TestStruct {
            description : decode.string(),
            id          : decode.int32(),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut encode = BytesEncoder::new();
		encode.string(&self.description);
		encode.int32(self.id);
        return encode.data();
    }
}

pub struct ImmutableTestStruct {
    pub(crate) obj_id: i32,
    pub(crate) key_id: Key32,
}

impl ImmutableTestStruct {
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BYTES)
    }

    pub fn value(&self) -> TestStruct {
        TestStruct::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_BYTES))
    }
}

pub struct MutableTestStruct {
    pub(crate) obj_id: i32,
    pub(crate) key_id: Key32,
}

impl MutableTestStruct {
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BYTES)
    }

    pub fn set_value(&self, value: &TestStruct) {
        set_bytes(self.obj_id, self.key_id, TYPE_BYTES, &value.to_bytes());
    }

    pub fn value(&self) -> TestStruct {
        TestStruct::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_BYTES))
    }
}
