// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_INT16_LENGTH: usize = 2;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn int16_decode(dec: &mut WasmDecoder) -> i16 {
    dec.vli_decode(16) as i16
}

pub fn int16_encode(enc: &mut WasmEncoder, value: i16) {
    enc.vli_encode(value as i64);
}

pub fn int16_from_bytes(buf: &[u8]) -> i16 {
    if buf.len() == 0 {
        return 0;
    }
    if buf.len() != SC_INT16_LENGTH {
        panic("invalid Int16 length");
    }
    i16::from_le_bytes(buf.try_into().expect("WTF?"))
}

pub fn int16_to_bytes(value: i16) -> Vec<u8> {
    value.to_le_bytes().to_vec()
}

pub fn int16_from_string(value: &str) -> i16 {
    value.parse::<i16>().unwrap()
}

pub fn int16_to_string(value: i16) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableInt16 {
    proxy: Proxy,
}

impl ScImmutableInt16 {
    pub fn new(proxy: Proxy) -> ScImmutableInt16 {
        ScImmutableInt16 { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        int16_to_string(self.value())
    }

    pub fn value(&self) -> i16 {
        int16_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable i16 in host container
pub struct ScMutableInt16 {
    proxy: Proxy,
}

impl ScMutableInt16 {
    pub fn new(proxy: Proxy) -> ScMutableInt16 {
        ScMutableInt16 { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: i16) {
        self.proxy.set(&int16_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        int16_to_string(self.value())
    }

    pub fn value(&self) -> i16 {
        int16_from_bytes(&self.proxy.get())
    }
}
