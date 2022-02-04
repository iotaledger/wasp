// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn bytes_decode(dec: &mut WasmDecoder) -> Vec<u8> {
    dec.bytes().to_vec()
}

pub fn bytes_encode(enc: &mut WasmEncoder, value: &[u8])  {
    enc.bytes(value);
}

pub fn bytes_from_bytes(buf: &[u8]) -> Vec<u8> {
    buf.to_vec()
}

pub fn bytes_to_bytes(value: &[u8]) -> Vec<u8> {
    value.to_vec()
}

pub fn bytes_to_string(value: &[u8]) -> String {
    // TODO standardize human readable string
    base58_encode(value)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableBytes {
    proxy: Proxy
}

impl ScImmutableBytes {
    pub fn new(proxy: Proxy) -> ScImmutableBytes {
        ScImmutableBytes { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        bytes_to_string(&self.value())
    }

    pub fn value(&self) -> Vec<u8> {
        bytes_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScBytes in host container
pub struct ScMutableBytes {
    proxy: Proxy
}

impl ScMutableBytes {
    pub fn new(proxy: Proxy) -> ScMutableBytes {
        ScMutableBytes { proxy }
    }

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &[u8]) {
        self.proxy.set(&bytes_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        bytes_to_string(&self.value())
    }

    pub fn value(&self) -> Vec<u8> {
        bytes_from_bytes(&self.proxy.get())
    }
}
