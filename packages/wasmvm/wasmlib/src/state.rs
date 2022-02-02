// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::host::*;
use crate::wasmtypes::IKvStore;

pub struct ScImmutableState {}

impl ScImmutableState {
    pub fn exists(&self, key: &[u8]) -> bool {
        state_exists(key)
    }

    pub fn get(&self, key: &[u8]) -> Vec<u8> {
        state_get(key)
    }
}

pub struct ScState {}

impl IKvStore for ScState {
    fn delete(&mut self, key: &[u8]) {
        state_delete(key);
    }

    fn exists(&self, key: &[u8]) -> bool {
        state_exists(key)
    }

    fn get(&self, key: &[u8]) -> Vec<u8> {
        state_get(key)
    }

    fn set(&mut self, key: &[u8], value: &[u8]) {
        state_set(key, value);
    }
}
