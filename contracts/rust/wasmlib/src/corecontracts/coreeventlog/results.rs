// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

#![allow(dead_code)]
#![allow(unused_imports)]

use crate::*;
use crate::corecontracts::coreeventlog::*;
use crate::host::*;

#[derive(Clone, Copy)]
pub struct ImmutableGetNumRecordsResults {
    pub(crate) id: i32,
}

impl ImmutableGetNumRecordsResults {
    pub fn num_records(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, RESULT_NUM_RECORDS.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct MutableGetNumRecordsResults {
    pub(crate) id: i32,
}

impl MutableGetNumRecordsResults {
    pub fn num_records(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, RESULT_NUM_RECORDS.get_key_id())
    }
}

pub struct ArrayOfImmutableBytes {
    pub(crate) obj_id: i32,
}

impl ArrayOfImmutableBytes {
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }

    pub fn get_bytes(&self, index: i32) -> ScImmutableBytes {
        ScImmutableBytes::new(self.obj_id, Key32(index))
    }
}

#[derive(Clone, Copy)]
pub struct ImmutableGetRecordsResults {
    pub(crate) id: i32,
}

impl ImmutableGetRecordsResults {
    pub fn records(&self) -> ArrayOfImmutableBytes {
        let arr_id = get_object_id(self.id, RESULT_RECORDS.get_key_id(), TYPE_ARRAY | TYPE_BYTES);
        ArrayOfImmutableBytes { obj_id: arr_id }
    }
}

pub struct ArrayOfMutableBytes {
    pub(crate) obj_id: i32,
}

impl ArrayOfMutableBytes {
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }

    pub fn get_bytes(&self, index: i32) -> ScMutableBytes {
        ScMutableBytes::new(self.obj_id, Key32(index))
    }
}

#[derive(Clone, Copy)]
pub struct MutableGetRecordsResults {
    pub(crate) id: i32,
}

impl MutableGetRecordsResults {
    pub fn records(&self) -> ArrayOfMutableBytes {
        let arr_id = get_object_id(self.id, RESULT_RECORDS.get_key_id(), TYPE_ARRAY | TYPE_BYTES);
        ArrayOfMutableBytes { obj_id: arr_id }
    }
}
