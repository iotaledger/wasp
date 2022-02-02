// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::sandbox::*;
use crate::wasmtypes::*;

pub trait IKvStore {
    fn delete(&mut self, key: &[u8]);
    fn exists(&self, key: &[u8]) -> bool;
    fn get(&self, key: &[u8]) -> Vec<u8>;
    fn set(&mut self, key: &[u8], value: &[u8]);
}

pub struct Proxy<'a> {
    key: Vec<u8>,
    kv_store: &'a dyn IKvStore,
}

impl Proxy<'_> {
    pub fn new(kv_store: &dyn IKvStore) -> Proxy {
        Proxy { kv_store, key: Vec::new() }
    }

    pub fn append(&mut self) -> Proxy {
        let length = self.length();
        self.expand(length + 1);
        self.element(length)
    }

    pub fn clear_array(&mut self) {
        let mut length = self.length();
        while length != 0 {
            length -= 1;
            self.element(length).delete();
        }

        // clear the length counter
        self.delete();
    }

    pub fn clear_map(&mut self) {
        // TODO clear prefix

        // clear the length counter
        self.delete();
    }

    pub fn delete(&mut self) {
        self.kv_store.delete(&self.key);
    }

    fn element(&self, index: u32) -> Proxy {
        let mut enc = WasmEncoder::new();
        uint32_encode(&mut enc, index);
        // 0x23 is '#'
        self.sub(0x23, &enc.buf())
    }

    pub fn exists(&self) -> bool {
        self.kv_store.exists(&self.key)
    }

    pub fn expand(&mut self, length: u32)     {
        // update the length counter
        let mut enc = WasmEncoder::new();
        uint32_encode(&mut enc, length);
        self.set(&enc.buf());
    }

    pub fn get(&self) -> Vec<u8> {
        self.kv_store.get(&self.key)
    }

    pub fn index(&self, index: u32) -> Proxy {
        let size = self.length();
        if index >= size {
            if index == size {
                panic("invalid index: use append");
            }
            panic("invalid index");
        }
        self.element(index)
    }

    pub fn key(&self, key: &[u8]) -> Proxy {
        // 0x2e is '.'
        self.sub(0x2e, key)
    }

    pub fn length(&self) -> u32 {
        let buf = self.get();
        if buf.len() == 0 {
            return 0;
        }
        let mut dec = WasmDecoder::new(&buf);
        uint32_decode(&mut dec)
    }

    pub fn root(&self, key: &str) -> Proxy {
        Proxy { kv_store: self.kv_store, key: string_to_bytes(key).to_vec() }
    }

    pub fn set(&mut self, value: &[u8]) {
        self.kv_store.set(&self.key, value);
    }

    fn sub(&self, sep: u8, key: &[u8]) -> Proxy {
        if self.key.len() == 0 {
            return Proxy { kv_store: self.kv_store, key: key.to_vec() };
        }
        let mut buf: Vec<u8> = self.key.clone();
        buf.push(sep);
        buf.extend_from_slice(key);
        Proxy { kv_store: self.kv_store, key: buf }
    }
}
