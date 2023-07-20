// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::cell::RefCell;
use std::collections::BTreeMap;

use crate::*;
use crate::host::*;

pub struct ScImmutableDict {
    dict: ScDict,
}

impl ScImmutableDict {
    pub fn new(dict: ScDict) -> ScImmutableDict {
        ScImmutableDict { dict }
    }

    pub fn exists(&self, key: &[u8]) -> bool {
        self.dict.exists(key)
    }

    pub fn get(&self, key: &[u8]) -> Vec<u8> {
        self.dict.get(key)
    }
}

#[derive(Clone)]
pub struct ScDict {
    dict: RefCell<BTreeMap<Vec<u8>, Vec<u8>>>,
    state: bool,
}

impl ScDict {
    pub fn new(buf: &[u8]) -> ScDict {
        let dict = ScDict {
            dict: RefCell::new(BTreeMap::new()),
            state: false,
        };
        dict.read_bytes(buf);
        dict
    }

    pub fn state() -> ScDict {
        ScDict {
            dict: RefCell::new(BTreeMap::new()),
            state: true,
        }
    }

    pub fn copy(&self, buf: &[u8]) {
        self.dict.borrow_mut().clear();
        self.read_bytes(buf);
    }

    pub fn delete(&self, key: &[u8]) {
        if self.state {
            state_delete(key);
            return;
        }
        self.dict.borrow_mut().remove(key);
    }

    pub fn exists(&self, key: &[u8]) -> bool {
        if self.state {
            return state_exists(key);
        }
        self.dict.borrow().contains_key(key)
    }

    pub fn get(&self, key: &[u8]) -> Vec<u8> {
        if self.state {
            return state_get(key);
        }
        let dict = self.dict.borrow();
        if !dict.contains_key(key) {
            return Vec::new();
        }
        dict.get(key).unwrap().to_vec()
    }

    pub fn set(&self, key: &[u8], value: &[u8]) {
        if self.state {
            state_set(key, value);
            return;
        }
        self.dict.borrow_mut().insert(key.to_vec(), value.to_vec());
    }

    fn read_bytes(&self, buf: &[u8]) {
        if buf.len() != 0 {
            let mut dec = WasmDecoder::new(buf);
            let size = dec.vlu_decode(32);
            for _ in 0..size {
                let key_len = dec.vlu_decode(32) as usize;
                let key = dec.fixed_bytes(key_len);
                let val_len = dec.vlu_decode(32) as usize;
                let val = dec.fixed_bytes(val_len);
                self.set(&key, &val);
            }
        }
    }

    pub fn from_bytes(buf: &[u8]) -> Result<ScDict, String> {
        let dict = ScDict::new(&[]);
        dict.read_bytes(buf);
        return Ok(dict);
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let dict = self.dict.borrow();
        if dict.len() == 0 {
            return vec![0; 1];
        }

        let mut enc = WasmEncoder::new();
        enc.vlu_encode(dict.len() as u64);
        for (key, val) in dict.iter() {
            let key_len = key.len();
            enc.vlu_encode(key_len as u64);
            enc.fixed_bytes(key, key_len);
            let val_len = val.len();
            enc.vlu_encode(val_len as u64);
            enc.fixed_bytes(val, val_len);
        }
        return enc.buf();
    }
}
