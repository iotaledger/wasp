// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::cell::RefCell;
use std::collections::BTreeMap;

use crate::host::*;
use crate::*;

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
            let size = uint32_from_bytes(&dec.fixed_bytes(SC_UINT32_LENGTH));
            for _i in 0..size {
                let key_buf = dec.fixed_bytes(SC_UINT16_LENGTH);
                let key_len = uint16_from_bytes(&key_buf);
                let key = dec.fixed_bytes(key_len as usize);
                let val_buf = dec.fixed_bytes(SC_UINT32_LENGTH);
                let val_len = uint32_from_bytes(&val_buf);
                let val = dec.fixed_bytes(val_len as usize);
                self.set(&key, &val);
            }
        }
    }

    pub fn from_bytes(input: &[u8]) -> Result<ScDict, String> {
        return Err("not impl".to_string());
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let dict = self.dict.borrow();
        if dict.len() == 0 {
            return vec![0; SC_UINT32_LENGTH];
        }

        let mut enc = WasmEncoder::new();
        enc.fixed_bytes(&uint32_to_bytes(dict.len() as u32), SC_UINT32_LENGTH);
        for (key, val) in dict.iter() {
            enc.fixed_bytes(&uint16_to_bytes(key.len() as u16), SC_UINT16_LENGTH);
            enc.fixed_bytes(key, key.len());
            enc.fixed_bytes(&uint32_to_bytes(val.len() as u32), SC_UINT32_LENGTH);
            enc.fixed_bytes(val, val.len());
        }
        return enc.buf();
    }
}
