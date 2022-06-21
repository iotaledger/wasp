// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_HASH_LENGTH: usize = 32;

#[derive(PartialEq, Clone)]
pub struct ScHash {
    id: [u8; SC_HASH_LENGTH],
}

impl ScHash {
    pub fn to_bytes(&self) -> Vec<u8> {
        hash_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        hash_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn hash_decode(dec: &mut WasmDecoder) -> ScHash {
    hash_from_bytes_unchecked(&dec.fixed_bytes(SC_HASH_LENGTH))
}

pub fn hash_encode(enc: &mut WasmEncoder, value: &ScHash) {
    enc.fixed_bytes(&value.id, SC_HASH_LENGTH);
}

pub fn hash_from_bytes(buf: &[u8]) -> ScHash {
    if buf.len() == 0 {
        return ScHash {
            id: [0; SC_HASH_LENGTH],
        };
    }
    if buf.len() != SC_HASH_LENGTH {
        panic("invalid Hash length");
    }
    ScHash {
        id: buf.try_into().expect("WTF?"),
    }
}

pub fn hash_to_bytes(value: &ScHash) -> Vec<u8> {
    value.id.to_vec()
}

pub fn hash_from_string(value: &str) -> ScHash {
    hash_from_bytes(&hex_decode(value))
}

pub fn hash_to_string(value: &ScHash) -> String {
    hex_encode(&value.id)
}

fn hash_from_bytes_unchecked(buf: &[u8]) -> ScHash {
    ScHash {
        id: buf.try_into().expect("invalid Hash length"),
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableHash {
    proxy: Proxy,
}

impl ScImmutableHash {
    pub fn new(proxy: Proxy) -> ScImmutableHash {
        ScImmutableHash { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        hash_to_string(&self.value())
    }

    pub fn value(&self) -> ScHash {
        hash_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHash in host container
pub struct ScMutableHash {
    proxy: Proxy,
}

impl ScMutableHash {
    pub fn new(proxy: Proxy) -> ScMutableHash {
        ScMutableHash { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScHash) {
        self.proxy.set(&hash_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        hash_to_string(&self.value())
    }

    pub fn value(&self) -> ScHash {
        hash_from_bytes(&self.proxy.get())
    }
}
