// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::collections::HashMap;

use crate::host::*;
use crate::wasmtypes::IKvStore;

pub struct ScImmutableDict<'a> {
    dict: &'a HashMap<Vec<u8>,Vec<u8>>,
}

impl ScImmutableDict<'_> {
    pub fn from_dict(dict: &ScDict)-> ScImmutableDict {
        ScImmutableDict{ dict: &dict.dict }
    }

    pub fn exists(&self, key: &[u8]) -> bool {
        self.dict.contains_key(key)
    }

    pub fn get(&self, key: &[u8]) -> Option<&Vec<u8>> {
        self.dict.get(key)
    }
}

pub struct ScDict {
    dict: HashMap<Vec<u8>,Vec<u8>>,
}

impl ScDict {
    pub fn new(buf: &[u8]) -> ScDict {
        ScDict { dict: HashMap::new(), }
    }
}

impl IKvStore for ScDict {
    fn delete(& mut self, key: &[u8]) {
        self.dict.remove(key);
    }

    fn exists(&self, key: &[u8]) -> bool {
        self.dict.contains_key(key)
    }

    fn get(&self, key: &[u8]) -> Vec<u8> {
        if !self.dict.contains_key(key) {
            return Vec::new();
        }
        self.dict.get(key).unwrap().to_vec()
    }

    fn set(& mut self, key: &[u8], value: &[u8]) {
        self.dict.insert(key.to_vec(), value.to_vec());
    }
}
