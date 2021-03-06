// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

#![allow(dead_code)]
#![allow(unused_imports)]

use crate::*;
use crate::corecontracts::coreroot::*;
use crate::host::*;

#[derive(Clone, Copy)]
pub struct ImmutableFindContractResults {
    pub(crate) id: i32,
}

impl ImmutableFindContractResults {
    pub fn data(&self) -> ScImmutableBytes {
        ScImmutableBytes::new(self.id, RESULT_DATA.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct MutableFindContractResults {
    pub(crate) id: i32,
}

impl MutableFindContractResults {
    pub fn data(&self) -> ScMutableBytes {
        ScMutableBytes::new(self.id, RESULT_DATA.get_key_id())
    }
}

pub struct MapHnameToImmutableBytes {
    pub(crate) obj_id: i32,
}

impl MapHnameToImmutableBytes {
    pub fn get_bytes(&self, key: &ScHname) -> ScImmutableBytes {
        ScImmutableBytes::new(self.obj_id, key.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct ImmutableGetChainInfoResults {
    pub(crate) id: i32,
}

impl ImmutableGetChainInfoResults {
    pub fn chain_id(&self) -> ScImmutableChainID {
        ScImmutableChainID::new(self.id, RESULT_CHAIN_ID.get_key_id())
    }

    pub fn chain_owner_id(&self) -> ScImmutableAgentID {
        ScImmutableAgentID::new(self.id, RESULT_CHAIN_OWNER_ID.get_key_id())
    }

    pub fn contract_registry(&self) -> MapHnameToImmutableBytes {
        let map_id = get_object_id(self.id, RESULT_CONTRACT_REGISTRY.get_key_id(), TYPE_MAP);
        MapHnameToImmutableBytes { obj_id: map_id }
    }

    pub fn default_owner_fee(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, RESULT_DEFAULT_OWNER_FEE.get_key_id())
    }

    pub fn default_validator_fee(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, RESULT_DEFAULT_VALIDATOR_FEE.get_key_id())
    }

    pub fn description(&self) -> ScImmutableString {
        ScImmutableString::new(self.id, RESULT_DESCRIPTION.get_key_id())
    }

    pub fn fee_color(&self) -> ScImmutableColor {
        ScImmutableColor::new(self.id, RESULT_FEE_COLOR.get_key_id())
    }
}

pub struct MapHnameToMutableBytes {
    pub(crate) obj_id: i32,
}

impl MapHnameToMutableBytes {
    pub fn clear(&self) {
        clear(self.obj_id)
    }

    pub fn get_bytes(&self, key: &ScHname) -> ScMutableBytes {
        ScMutableBytes::new(self.obj_id, key.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct MutableGetChainInfoResults {
    pub(crate) id: i32,
}

impl MutableGetChainInfoResults {
    pub fn chain_id(&self) -> ScMutableChainID {
        ScMutableChainID::new(self.id, RESULT_CHAIN_ID.get_key_id())
    }

    pub fn chain_owner_id(&self) -> ScMutableAgentID {
        ScMutableAgentID::new(self.id, RESULT_CHAIN_OWNER_ID.get_key_id())
    }

    pub fn contract_registry(&self) -> MapHnameToMutableBytes {
        let map_id = get_object_id(self.id, RESULT_CONTRACT_REGISTRY.get_key_id(), TYPE_MAP);
        MapHnameToMutableBytes { obj_id: map_id }
    }

    pub fn default_owner_fee(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, RESULT_DEFAULT_OWNER_FEE.get_key_id())
    }

    pub fn default_validator_fee(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, RESULT_DEFAULT_VALIDATOR_FEE.get_key_id())
    }

    pub fn description(&self) -> ScMutableString {
        ScMutableString::new(self.id, RESULT_DESCRIPTION.get_key_id())
    }

    pub fn fee_color(&self) -> ScMutableColor {
        ScMutableColor::new(self.id, RESULT_FEE_COLOR.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct ImmutableGetFeeInfoResults {
    pub(crate) id: i32,
}

impl ImmutableGetFeeInfoResults {
    pub fn fee_color(&self) -> ScImmutableColor {
        ScImmutableColor::new(self.id, RESULT_FEE_COLOR.get_key_id())
    }

    pub fn owner_fee(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, RESULT_OWNER_FEE.get_key_id())
    }

    pub fn validator_fee(&self) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.id, RESULT_VALIDATOR_FEE.get_key_id())
    }
}

#[derive(Clone, Copy)]
pub struct MutableGetFeeInfoResults {
    pub(crate) id: i32,
}

impl MutableGetFeeInfoResults {
    pub fn fee_color(&self) -> ScMutableColor {
        ScMutableColor::new(self.id, RESULT_FEE_COLOR.get_key_id())
    }

    pub fn owner_fee(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, RESULT_OWNER_FEE.get_key_id())
    }

    pub fn validator_fee(&self) -> ScMutableInt64 {
        ScMutableInt64::new(self.id, RESULT_VALIDATOR_FEE.get_key_id())
    }
}
