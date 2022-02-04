// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_INT32_LENGTH: usize = 4;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn int32_decode(dec: &mut WasmDecoder) -> i32 {
    dec.vli_decode(32) as i32
}

pub fn int32_encode(enc: &mut WasmEncoder, value: i32)  {
    enc.vli_encode(value as i64);
}

pub fn int32_from_bytes(buf: &[u8]) -> i32 {
    if buf.len() == 0 {
        return 0;
    }
    if buf.len() != SC_INT32_LENGTH {
        panic("invalid Int32 length");
    }
    i32::from_le_bytes(buf.try_into().expect("WTF?"))
}

pub fn int32_to_bytes(value: i32) -> Vec<u8> {
    value.to_le_bytes().to_vec()
}

pub fn int32_to_string(value: i32) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableInt32 {
    proxy: Proxy
}

impl ScImmutableInt32 {
    pub fn new(proxy: Proxy) -> ScImmutableInt32 {
        ScImmutableInt32 { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        int32_to_string(self.value())
    }

    pub fn value(&self) -> i32 {
        int32_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable i32 in host container
pub struct ScMutableInt32 {
    proxy: Proxy
}

impl ScMutableInt32 {
    pub fn new(proxy: Proxy) -> ScMutableInt32 {
        ScMutableInt32 { proxy }
    }

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: i32) {
        self.proxy.set(&int32_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        int32_to_string(self.value())
    }

    pub fn value(&self) -> i32 {
        int32_from_bytes(&self.proxy.get())
    }
}
