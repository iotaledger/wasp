// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn string_decode(dec: &mut WasmDecoder) -> String {
    let length = uint16_decode(dec);
    string_from_bytes(&dec.fixed_bytes(length as usize))
}

pub fn string_encode(enc: &mut WasmEncoder, value: &str) {
    let buf = string_to_bytes(value);
    uint16_encode(enc, buf.len() as u16);
    enc.fixed_bytes(&buf, buf.len());
}

pub fn string_from_bytes(buf: &[u8]) -> String {
    unsafe {
        String::from_utf8_unchecked(buf.to_vec()).to_string()
    }
}

pub fn string_to_bytes(value: &str) -> Vec<u8> {
    value.as_bytes().to_vec()
}

pub fn string_from_string(value: &str) -> String {
    value.to_string()
}

pub fn string_to_string(value: &str) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableString {
    proxy: Proxy,
}

impl ScImmutableString {
    pub fn new(proxy: Proxy) -> ScImmutableString {
        ScImmutableString { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        self.value()
    }

    pub fn value(&self) -> String {
        string_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScString in host container
pub struct ScMutableString {
    proxy: Proxy,
}

impl ScMutableString {
    pub fn new(proxy: Proxy) -> ScMutableString {
        ScMutableString { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &str) {
        self.proxy.set(&string_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        self.value()
    }

    pub fn value(&self) -> String {
        string_from_bytes(&self.proxy.get())
    }
}
