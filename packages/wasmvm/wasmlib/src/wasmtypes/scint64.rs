// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_INT64_LENGTH: usize = 8;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn int64_decode(dec: &mut WasmDecoder) -> i64 {
    dec.vli_decode(64)
}

pub fn int64_encode(enc: &mut WasmEncoder, value: i64) {
    enc.vli_encode(value);
}

pub fn int64_from_bytes(buf: &[u8]) -> i64 {
    if buf.len() == 0 {
        return 0;
    }
    if buf.len() != SC_INT64_LENGTH {
        panic("invalid Int64 length");
    }
    i64::from_le_bytes(buf.try_into().expect("WTF?"))
}

pub fn int64_to_bytes(value: i64) -> Vec<u8> {
    value.to_le_bytes().to_vec()
}

pub fn int64_from_string(value: &str) -> i64 {
    value.parse::<i64>().unwrap()
}

pub fn int64_to_string(value: i64) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableInt64 {
    proxy: Proxy,
}

impl ScImmutableInt64 {
    pub fn new(proxy: Proxy) -> ScImmutableInt64 {
        ScImmutableInt64 { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        int64_to_string(self.value())
    }

    pub fn value(&self) -> i64 {
        int64_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable i64 in host container
pub struct ScMutableInt64 {
    proxy: Proxy,
}

impl ScMutableInt64 {
    pub fn new(proxy: Proxy) -> ScMutableInt64 {
        ScMutableInt64 { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: i64) {
        self.proxy.set(&int64_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        int64_to_string(self.value())
    }

    pub fn value(&self) -> i64 {
        int64_from_bytes(&self.proxy.get())
    }
}
