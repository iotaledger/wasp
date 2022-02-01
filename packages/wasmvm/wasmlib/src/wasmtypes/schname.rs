// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::wasmtypes::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_HNAME_LENGTH: usize = 4;

#[derive(PartialEq, Clone)]
pub struct ScHname {
    id: [u8; SC_HNAME_LENGTH],
}

impl ScHname {
    pub fn from_bytes(buf: &[u8]) -> ScHname {
        hname_from_bytes(buf)
    }

    pub fn to_bytes(&self) -> &[u8] {
        hname_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        hname_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn hname_decode(dec: &mut WasmDecoder) -> ScHname {
    hname_from_bytes_unchecked(dec.fixed_bytes(SC_HNAME_LENGTH))
}

pub fn hname_encode(enc: &mut WasmEncoder, value: &ScHname)  {
    enc.fixed_bytes(&value.to_bytes(), SC_HNAME_LENGTH);
}

pub fn hname_from_bytes(buf: &[u8]) -> ScHname {
    ScHname { id: buf.try_into().expect("invalid Hname length") }
}

pub fn hname_to_bytes(value: &ScHname) -> &[u8] {
    &value.id
}

pub fn hname_to_string(value: &ScHname) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

fn hname_from_bytes_unchecked(buf: &[u8]) -> ScHname {
    ScHname { id: buf.try_into().expect("invalid Hname length") }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableHname<'a> {
    proxy: Proxy<'a>,
}

impl ScImmutableHname<'_> {
    pub fn new(proxy: Proxy) -> ScImmutableHname {
        ScImmutableHname { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        hname_to_string(&self.value())
    }

    pub fn value(&self) -> ScHname {
        hname_from_bytes(self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHname in host container
pub struct ScMutableHname<'a> {
    proxy: Proxy<'a>,
}

impl ScMutableHname<'_> {
    pub fn new(proxy: Proxy) -> ScMutableHname {
        ScMutableHname { proxy }
    }

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScHname) {
        self.proxy.set(hname_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        hname_to_string(&self.value())
    }

    pub fn value(&self) -> ScHname {
        hname_from_bytes(self.proxy.get())
    }
}
