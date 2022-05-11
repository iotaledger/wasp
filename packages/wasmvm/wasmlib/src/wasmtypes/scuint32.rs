// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_UINT32_LENGTH: usize = 4;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn uint32_decode(dec: &mut WasmDecoder) -> u32 {
    dec.vlu_decode(32) as u32
}

pub fn uint32_encode(enc: &mut WasmEncoder, value: u32) {
    enc.vlu_encode(value as u64);
}

pub fn uint32_from_bytes(buf: &[u8]) -> u32 {
    if buf.len() == 0 {
        return 0;
    }
    if buf.len() != SC_UINT32_LENGTH {
        panic("invalid Uint32 length");
    }
    u32::from_le_bytes(buf.try_into().expect("WTF?"))
}

pub fn uint32_to_bytes(value: u32) -> Vec<u8> {
    value.to_le_bytes().to_vec()
}

pub fn uint32_from_string(value: &str) -> u32 {
    value.parse::<u32>().unwrap()
}

pub fn uint32_to_string(value: u32) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableUint32 {
    proxy: Proxy,
}

impl ScImmutableUint32 {
    pub fn new(proxy: Proxy) -> ScImmutableUint32 {
        ScImmutableUint32 { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        uint32_to_string(self.value())
    }

    pub fn value(&self) -> u32 {
        uint32_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable u32 in host container
pub struct ScMutableUint32 {
    proxy: Proxy,
}

impl ScMutableUint32 {
    pub fn new(proxy: Proxy) -> ScMutableUint32 {
        ScMutableUint32 { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: u32) {
        self.proxy.set(&uint32_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        uint32_to_string(self.value())
    }

    pub fn value(&self) -> u32 {
        uint32_from_bytes(&self.proxy.get())
    }
}
