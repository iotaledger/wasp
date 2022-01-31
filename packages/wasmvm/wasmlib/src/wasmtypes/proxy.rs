// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::wasmtypes::*;

pub trait IKvStore {
    fn delete(&self, key: &[u8]);
    fn exists(&self, key: &[u8]) -> bool;
    fn get(&self, key: &[u8]) -> &[u8];
    fn set(&self, key: &[u8], value: &[u8]);
}

pub struct Proxy<'a>  {
    key: Vec<u8>,
    kv_store: &'a dyn IKvStore,
}

impl Proxy<'_> {
    pub fn new(kv_store: &dyn IKvStore) -> Proxy {
        Proxy { key:Vec::new(), kv_store }
    }

    pub fn delete(&self) {
        self.kv_store.delete(&self.key);
    }

    pub fn exists(&self) -> bool {
        self.kv_store.exists(&self.key)
    }

    pub fn get(&self) -> &[u8] {
        self.kv_store.get(&self.key)
    }

    pub fn root(&self, key:&str) -> Proxy {
        Proxy { key: key.as_bytes().to_vec(), kv_store: self.kv_store }
    }

    pub fn set(&self, value: &[u8]) {
        self.kv_store.set(&self.key, value);
    }
}
