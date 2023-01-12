// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::cell::RefCell;
use std::rc::Rc;

use crate::*;

#[derive(Clone)]
pub struct Proxy {
    key: Vec<u8>,
    pub(crate) kv_store: Rc<ScDict>,
}

impl Proxy {
    pub fn new(kv_store: &Rc<ScDict>) -> Proxy {
        Proxy { kv_store: Rc::clone(kv_store), key: Vec::new() }
    }

    pub fn nil() -> Proxy {
        Proxy::new(&Rc::new(ScDict::new(&[])))
    }

    pub fn link(proxy: &mut Proxy, kv_store: &Rc<ScDict>) {
        proxy.kv_store = Rc::clone(kv_store);
    }

    pub fn append(&self) -> Proxy {
        let length = self.length();
        self.expand(length + 1);
        self.element(length)
    }

    pub fn clear_array(&self) {
        let mut length = self.length();
        while length != 0 {
            length -= 1;
            self.element(length).delete();
        }

        // clear the length counter
        self.delete();
    }

    pub fn clear_map(&self) {
        // TODO clear prefix

        // clear the length counter
        self.delete();
    }

    pub fn delete(&self) {
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

    pub fn expand(&self, length: u32) {
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
        Proxy { kv_store: Rc::clone(&self.kv_store), key: string_to_bytes(key).to_vec() }
    }

    pub fn set(&self, value: &[u8]) {
        self.kv_store.set(&self.key, value);
    }

    fn sub(&self, sep: u8, key: &[u8]) -> Proxy {
        if self.key.len() == 0 {
            return Proxy { kv_store: Rc::clone(&self.kv_store), key: key.to_vec() };
        }
        let mut buf: Vec<u8> = self.key.clone();
        buf.push(sep);
        buf.extend_from_slice(key);
        Proxy { kv_store: Rc::clone(&self.kv_store), key: buf }
    }
}
