// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::wasmtypes::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn string_decode(dec: &mut WasmDecoder) -> String {
    string_from_bytes(&dec.bytes())
}

pub fn string_encode(enc: &mut WasmEncoder, value: &str) {
    enc.bytes(&string_to_bytes(value));
}

pub fn string_from_bytes(buf: &[u8]) -> String {
    String::from_utf8_lossy(buf).to_string()
}

pub fn string_to_bytes(value: &str) -> Vec<u8> {
    value.to_vec()
}

pub fn string_to_string(value: &str) -> String {
    value.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableString<'a> {
    proxy: Proxy<'a>,
}

impl ScImmutableString<'_> {
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
pub struct ScMutableString<'a> {
    proxy: Proxy<'a>,
}

impl ScMutableString<'_> {
    pub fn new(proxy: Proxy) -> ScMutableString {
        ScMutableString { proxy }
    }

    pub fn delete(&mut self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&mut self, value: &str) {
        self.proxy.set(&string_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        self.value()
    }

    pub fn value(&self) -> String {
        string_from_bytes(&self.proxy.get())
    }
}
