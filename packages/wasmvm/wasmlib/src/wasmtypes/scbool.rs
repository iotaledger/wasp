// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_BOOL_LENGTH: usize = 1;
pub const SC_BOOL_FALSE: u8 = 0x00;
pub const SC_BOOL_TRUE: u8 = 0xff;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn bool_decode(dec: &mut WasmDecoder) -> bool {
    dec.byte() != SC_BOOL_FALSE
}

pub fn bool_encode(enc: &mut WasmEncoder, value: bool) {
    if value {
        enc.byte(SC_BOOL_TRUE);
        return;
    }
    enc.byte(SC_BOOL_FALSE);
}

pub fn bool_from_bytes(buf: &[u8]) -> bool {
    if buf.len() == 0 {
        return false;
    }
    if buf.len() != SC_BOOL_LENGTH {
        panic("invalid Bool length");
    }
    if buf[0] == SC_BOOL_FALSE {
        return false;
    }
    if buf[0] != SC_BOOL_TRUE {
        panic("invalid Bool value");
    }
    return true;
}

pub fn bool_to_bytes(value: bool) -> Vec<u8> {
    if value {
        return [SC_BOOL_TRUE].to_vec();
    }
    [SC_BOOL_FALSE].to_vec()
}

pub fn bool_from_string(value: &str) -> bool {
    match value {
        "0" => return false,
        "1" => return true,
        _ => panic("invalid Bool string")
    }
    false
}

pub fn bool_to_string(value: bool) -> String {
    if value {
        return "1".to_string();
    }
    "0".to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableBool {
    proxy: Proxy,
}

impl ScImmutableBool {
    pub fn new(proxy: Proxy) -> ScImmutableBool {
        ScImmutableBool { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        bool_to_string(self.value())
    }

    pub fn value(&self) -> bool {
        bool_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable bool in host container
pub struct ScMutableBool {
    proxy: Proxy,
}

impl ScMutableBool {
    pub fn new(proxy: Proxy) -> ScMutableBool {
        ScMutableBool { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: bool) {
        self.proxy.set(&bool_to_bytes(value));
    }

    pub fn to_string(&self) -> String {
        bool_to_string(self.value())
    }

    pub fn value(&self) -> bool {
        bool_from_bytes(&self.proxy.get())
    }
}
