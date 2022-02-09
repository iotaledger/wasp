// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_HNAME_LENGTH: usize = 4;

#[derive(PartialEq, Clone, Copy)]
pub struct ScHname(pub u32);

impl ScHname {
    pub fn new(name: &str]) -> ScHname {
        hname_from_bytes(&host::sandbox(FN_UTILS_HASH_NAME, &string_to_bytes(name)))
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        hname_to_bytes(*self)
    }

    pub fn to_string(&self) -> String {
        hname_to_string(*self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn hname_decode(dec: &mut WasmDecoder) -> ScHname {
    hname_from_bytes(&dec.fixed_bytes(SC_HNAME_LENGTH))
}

pub fn hname_encode(enc: &mut WasmEncoder, value: ScHname) {
    enc.fixed_bytes(&hname_to_bytes(value), SC_HNAME_LENGTH);
}

pub fn hname_from_bytes(buf: &[u8]) -> ScHname {
    if buf.len() == 0 {
        return ScHname(0);
    }
    if buf.len() != SC_HNAME_LENGTH {
        panic("invalid Hname length");
    }
    ScHname(u32::from_le_bytes(buf.try_into().expect("WTF?")))
}

pub fn hname_to_bytes(value: ScHname) -> Vec<u8> {
    value.0.to_le_bytes().to_vec()
}

pub fn hname_to_string(value: ScHname) -> String {
    let hexa = "0123456789abcdef".as_bytes();
    let mut res = [0u8; 8];
    let mut val = value.0;
    for n in 0..8 {
        res[7 - n] = hexa[val as usize & 0x0f];
        val >>= 4;
    }
    string_from_bytes(&res)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableHname {
    proxy: Proxy,
}

impl ScImmutableHname {
    pub fn new(proxy: Proxy) -> ScImmutableHname {
        ScImmutableHname { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        hname_to_string(self.value())
    }

    pub fn value(&self) -> ScHname {
        hname_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHname in host container
pub struct ScMutableHname {
    proxy: Proxy,
}

impl ScMutableHname {
    pub fn new(proxy: Proxy) -> ScMutableHname {
        ScMutableHname { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: ScHname) {
        self.proxy.set(&hname_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        hname_to_string(self.value())
    }

    pub fn value(&self) -> ScHname {
        hname_from_bytes(&self.proxy.get())
    }
}
