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
pub struct ImmutableGetNumRecordsParams {
    pub(crate) id: i32,
}

impl ImmutableGetNumRecordsParams {
    pub fn contract_hname(&self) -> ScImmutableHname {
        ScImmutableHname::new(self.id, PARAM_CONTRACT_HNAME.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct MutableGetNumRecordsParams {
    pub(crate) id: i32,
}

impl MutableGetNumRecordsParams {
    pub fn contract_hname(&self) -> ScMutableHname {
        ScMutableHname::new(self.id, PARAM_CONTRACT_HNAME.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct ImmutableGetRecordsParams {
    pub(crate) id: i32,
}

impl ImmutableGetRecordsParams {
    pub fn contract_hname(&self) -> ScImmutableHname {
        ScImmutableHname::new(self.id, PARAM_CONTRACT_HNAME.get_key_id())
    }

    pub fn from_ts(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, PARAM_FROM_TS.get_key_id())
    }

    pub fn max_last_records(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, PARAM_MAX_LAST_RECORDS.get_key_id())
    }

    pub fn to_ts(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, PARAM_TO_TS.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct MutableGetRecordsParams {
    pub(crate) id: i32,
}

impl MutableGetRecordsParams {
    pub fn contract_hname(&self) -> ScMutableHname {
        ScMutableHname::new(self.id, PARAM_CONTRACT_HNAME.get_key_id())
    }

    pub fn from_ts(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, PARAM_FROM_TS.get_key_id())
    }

    pub fn max_last_records(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, PARAM_MAX_LAST_RECORDS.get_key_id())
    }

    pub fn to_ts(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, PARAM_TO_TS.get_key_id())
    }
}
