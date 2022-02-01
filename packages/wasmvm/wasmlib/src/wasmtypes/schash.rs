// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::wasmtypes::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_HASH_LENGTH: usize = 32;

#[derive(PartialEq, Clone)]
pub struct ScHash {
    id: [u8; SC_HASH_LENGTH],
}

impl ScHash {
    pub fn from_bytes(buf: &[u8]) -> ScHash {
        hash_from_bytes(buf)
    }

    pub fn to_bytes(&self) -> &[u8] {
        hash_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        hash_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn hash_decode(dec: &mut WasmDecoder) -> ScHash {
    hash_from_bytes_unchecked(dec.fixed_bytes(SC_HASH_LENGTH))
}

pub fn hash_encode(enc: &mut WasmEncoder, value: &ScHash)  {
    enc.fixed_bytes(&value.to_bytes(), SC_HASH_LENGTH);
}

pub fn hash_from_bytes(buf: &[u8]) -> ScHash {
    ScHash { id: buf.try_into().expect("invalid Hash length") }
}

pub fn hash_to_bytes(value: &ScHash) -> &[u8] {
    &value.id
}

pub fn hash_to_string(value: &ScHash) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

fn hash_from_bytes_unchecked(buf: &[u8]) -> ScHash {
    ScHash { id: buf.try_into().expect("invalid Hash length") }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableHash<'a> {
    proxy: Proxy<'a>,
}

impl ScImmutableHash<'_> {
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
        hash_from_bytes(self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHash in host container
pub struct ScMutableHash<'a> {
    proxy: Proxy<'a>,
}

impl ScMutableHash<'_> {
    pub fn new(proxy: Proxy) -> ScMutableHash {
        ScMutableHash { proxy }
    }

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScHash) {
        self.proxy.set(hash_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        hash_to_string(&self.value())
    }

    pub fn value(&self) -> ScHash {
        hash_from_bytes(self.proxy.get())
    }
}
