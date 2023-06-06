// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::*;

pub const SC_UINT16_LENGTH: usize = 2;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn uint16_decode(dec: &mut WasmDecoder) -> u16 {
    uint16_from_bytes(&dec.fixed_bytes(SC_UINT16_LENGTH))
}

pub fn uint16_encode(enc: &mut WasmEncoder, value: u16) {
    enc.fixed_bytes(&uint16_to_bytes(value), SC_UINT16_LENGTH);
}

pub fn uint16_from_bytes(buf: &[u8]) -> u16 {
    let len = buf.len();
    if len == 0 {
        return 0;
    }
    if len != SC_UINT16_LENGTH {
        panic("invalid Uint16 length");
    }
    u16::from_le_bytes(buf.try_into().expect("WTF?"))
}

pub fn uint16_to_bytes(value: u16) -> Vec<u8> {
    value.to_le_bytes().to_vec()
}

pub fn uint16_from_string(value: &str) -> u16 {
    value.parse::<u16>().unwrap()
}

pub fn uint16_to_string(value: u16) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableUint16 {
    proxy: Proxy,
}

impl ScImmutableUint16 {
    pub fn new(proxy: Proxy) -> ScImmutableUint16 {
        ScImmutableUint16 { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        uint16_to_string(self.value())
    }

    pub fn value(&self) -> u16 {
        uint16_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable u16 in host container
pub struct ScMutableUint16 {
    proxy: Proxy,
}

impl ScMutableUint16 {
    pub fn new(proxy: Proxy) -> ScMutableUint16 {
        ScMutableUint16 { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: u16) {
        self.proxy.set(&uint16_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        uint16_to_string(self.value())
    }

    pub fn value(&self) -> u16 {
        uint16_from_bytes(&self.proxy.get())
    }
}
