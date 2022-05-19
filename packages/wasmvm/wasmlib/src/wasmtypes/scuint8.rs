// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_UINT8_LENGTH: usize = 1;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn uint8_decode(dec: &mut WasmDecoder) -> u8 {
    dec.byte()
}

pub fn uint8_encode(enc: &mut WasmEncoder, value: u8) {
    enc.byte(value);
}

pub fn uint8_from_bytes(buf: &[u8]) -> u8 {
    if buf.len() == 0 {
        return 0;
    }
    if buf.len() != SC_UINT8_LENGTH {
        panic("invalid Uint8 length");
    }
    buf[0]
}

pub fn uint8_to_bytes(value: u8) -> Vec<u8> {
    [value].to_vec()
}

pub fn uint8_from_string(value: &str) -> u8 {
    value.parse::<u8>().unwrap()
}

pub fn uint8_to_string(value: u8) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableUint8 {
    proxy: Proxy,
}

impl ScImmutableUint8 {
    pub fn new(proxy: Proxy) -> ScImmutableUint8 {
        ScImmutableUint8 { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        uint8_to_string(self.value())
    }

    pub fn value(&self) -> u8 {
        uint8_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable u8 in host container
pub struct ScMutableUint8 {
    proxy: Proxy,
}

impl ScMutableUint8 {
    pub fn new(proxy: Proxy) -> ScMutableUint8 {
        ScMutableUint8 { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: u8) {
        self.proxy.set(&uint8_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        uint8_to_string(self.value())
    }

    pub fn value(&self) -> u8 {
        uint8_from_bytes(&self.proxy.get())
    }
}
