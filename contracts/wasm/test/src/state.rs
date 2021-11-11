// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use wasmlib::host::*;

use crate::*;
use crate::keys::*;
use crate::structs::*;

pub struct ArrayOfImmutableTestStruct {
	pub(crate) obj_id: i32,
}

impl ArrayOfImmutableTestStruct {
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }

	pub fn get_test_struct(&self, index: i32) -> ImmutableTestStruct {
		ImmutableTestStruct { obj_id: self.obj_id, key_id: Key32(index) }
	}
}

#[derive(Clone, Copy)]
pub struct ImmutabletestState {
    pub(crate) id: i32,
}

impl ImmutabletestState {
    pub fn test_structs(&self) -> ArrayOfImmutableTestStruct {
		let arr_id = get_object_id(self.id, idx_map(IDX_STATE_TEST_STRUCTS), TYPE_ARRAY | TYPE_BYTES);
		ArrayOfImmutableTestStruct { obj_id: arr_id }
	}
}

pub struct ArrayOfMutableTestStruct {
	pub(crate) obj_id: i32,
}

impl ArrayOfMutableTestStruct {
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }

	pub fn get_test_struct(&self, index: i32) -> MutableTestStruct {
		MutableTestStruct { obj_id: self.obj_id, key_id: Key32(index) }
	}
}

#[derive(Clone, Copy)]
pub struct MutabletestState {
    pub(crate) id: i32,
}

impl MutabletestState {
    pub fn test_structs(&self) -> ArrayOfMutableTestStruct {
		let arr_id = get_object_id(self.id, idx_map(IDX_STATE_TEST_STRUCTS), TYPE_ARRAY | TYPE_BYTES);
		ArrayOfMutableTestStruct { obj_id: arr_id }
	}
}
