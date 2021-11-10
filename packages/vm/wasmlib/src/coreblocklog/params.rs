// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

#![allow(dead_code)]
#![allow(unused_imports)]

use crate::*;
use crate::coreblocklog::*;
use crate::host::*;

#[derive(Clone, Copy)]
pub struct ImmutableGetBlockInfoParams {
    pub(crate) id: i32,
}

impl ImmutableGetBlockInfoParams {
    pub fn block_index(&self) -> ScImmutableInt32 {
		ScImmutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableGetBlockInfoParams {
    pub(crate) id: i32,
}

impl MutableGetBlockInfoParams {
    pub fn block_index(&self) -> ScMutableInt32 {
		ScMutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct ImmutableGetEventsForBlockParams {
    pub(crate) id: i32,
}

impl ImmutableGetEventsForBlockParams {
    pub fn block_index(&self) -> ScImmutableInt32 {
		ScImmutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableGetEventsForBlockParams {
    pub(crate) id: i32,
}

impl MutableGetEventsForBlockParams {
    pub fn block_index(&self) -> ScMutableInt32 {
		ScMutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct ImmutableGetEventsForContractParams {
    pub(crate) id: i32,
}

impl ImmutableGetEventsForContractParams {
    pub fn contract_hname(&self) -> ScImmutableHname {
		ScImmutableHname::new(self.id, PARAM_CONTRACT_HNAME.get_key_id())
	}

    pub fn from_block(&self) -> ScImmutableInt32 {
		ScImmutableInt32::new(self.id, PARAM_FROM_BLOCK.get_key_id())
	}

    pub fn to_block(&self) -> ScImmutableInt32 {
		ScImmutableInt32::new(self.id, PARAM_TO_BLOCK.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableGetEventsForContractParams {
    pub(crate) id: i32,
}

impl MutableGetEventsForContractParams {
    pub fn contract_hname(&self) -> ScMutableHname {
		ScMutableHname::new(self.id, PARAM_CONTRACT_HNAME.get_key_id())
	}

    pub fn from_block(&self) -> ScMutableInt32 {
		ScMutableInt32::new(self.id, PARAM_FROM_BLOCK.get_key_id())
	}

    pub fn to_block(&self) -> ScMutableInt32 {
		ScMutableInt32::new(self.id, PARAM_TO_BLOCK.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct ImmutableGetEventsForRequestParams {
    pub(crate) id: i32,
}

impl ImmutableGetEventsForRequestParams {
    pub fn request_id(&self) -> ScImmutableRequestID {
		ScImmutableRequestID::new(self.id, PARAM_REQUEST_ID.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableGetEventsForRequestParams {
    pub(crate) id: i32,
}

impl MutableGetEventsForRequestParams {
    pub fn request_id(&self) -> ScMutableRequestID {
		ScMutableRequestID::new(self.id, PARAM_REQUEST_ID.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct ImmutableGetRequestIDsForBlockParams {
    pub(crate) id: i32,
}

impl ImmutableGetRequestIDsForBlockParams {
    pub fn block_index(&self) -> ScImmutableInt32 {
		ScImmutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableGetRequestIDsForBlockParams {
    pub(crate) id: i32,
}

impl MutableGetRequestIDsForBlockParams {
    pub fn block_index(&self) -> ScMutableInt32 {
		ScMutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct ImmutableGetRequestReceiptParams {
    pub(crate) id: i32,
}

impl ImmutableGetRequestReceiptParams {
    pub fn request_id(&self) -> ScImmutableRequestID {
		ScImmutableRequestID::new(self.id, PARAM_REQUEST_ID.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableGetRequestReceiptParams {
    pub(crate) id: i32,
}

impl MutableGetRequestReceiptParams {
    pub fn request_id(&self) -> ScMutableRequestID {
		ScMutableRequestID::new(self.id, PARAM_REQUEST_ID.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct ImmutableGetRequestReceiptsForBlockParams {
    pub(crate) id: i32,
}

impl ImmutableGetRequestReceiptsForBlockParams {
    pub fn block_index(&self) -> ScImmutableInt32 {
		ScImmutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableGetRequestReceiptsForBlockParams {
    pub(crate) id: i32,
}

impl MutableGetRequestReceiptsForBlockParams {
    pub fn block_index(&self) -> ScMutableInt32 {
		ScMutableInt32::new(self.id, PARAM_BLOCK_INDEX.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct ImmutableIsRequestProcessedParams {
    pub(crate) id: i32,
}

impl ImmutableIsRequestProcessedParams {
    pub fn request_id(&self) -> ScImmutableRequestID {
		ScImmutableRequestID::new(self.id, PARAM_REQUEST_ID.get_key_id())
	}
}

#[derive(Clone, Copy)]
pub struct MutableIsRequestProcessedParams {
    pub(crate) id: i32,
}

impl MutableIsRequestProcessedParams {
    pub fn request_id(&self) -> ScMutableRequestID {
		ScMutableRequestID::new(self.id, PARAM_REQUEST_ID.get_key_id())
	}
}
