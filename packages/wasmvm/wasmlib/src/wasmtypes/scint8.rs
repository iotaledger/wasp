// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_INT8_LENGTH: usize = 1;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn int8_decode(dec: &mut WasmDecoder) -> i8 {
    dec.byte() as i8
}

pub fn int8_encode(enc: &mut WasmEncoder, value: i8) {
    enc.byte(value as u8);
}

pub fn int8_from_bytes(buf: &[u8]) -> i8 {
    if buf.len() == 0 {
        return 0;
    }
    if buf.len() != SC_INT8_LENGTH {
        panic("invalid Int8 length");
    }
    buf[0] as i8
}

pub fn int8_to_bytes(value: i8) -> Vec<u8> {
    [value as u8].to_vec()
}

pub fn int8_to_string(value: i8) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableInt8 {
    proxy: Proxy,
}

impl ScImmutableInt8 {
    pub fn new(proxy: Proxy) -> ScImmutableInt8 {
        ScImmutableInt8 { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        int8_to_string(self.value())
    }

    pub fn value(&self) -> i8 {
        int8_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable i8 in host container
pub struct ScMutableInt8 {
    proxy: Proxy,
}

impl ScMutableInt8 {
    pub fn new(proxy: Proxy) -> ScMutableInt8 {
        ScMutableInt8 { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: i8) {
        self.proxy.set(&int8_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        int8_to_string(self.value())
    }

    pub fn value(&self) -> i8 {
        int8_from_bytes(&self.proxy.get())
    }
}
