// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_TOKEN_ID_LENGTH: usize = 50;

#[derive(PartialEq, Clone, Copy, Eq, Hash)]
pub struct ScTokenID {
    id: [u8; SC_TOKEN_ID_LENGTH],
}

impl ScTokenID {
    pub fn to_bytes(&self) -> Vec<u8> {
        token_id_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        token_id_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn token_id_decode(dec: &mut WasmDecoder) -> ScTokenID {
    let buf = dec.fixed_bytes(SC_TOKEN_ID_LENGTH);
    ScTokenID { id: buf.try_into().expect("WTF?") }
}

pub fn token_id_encode(enc: &mut WasmEncoder, value: &ScTokenID) {
    enc.fixed_bytes(&value.id, SC_TOKEN_ID_LENGTH);
}

pub fn token_id_from_bytes(buf: &[u8]) -> ScTokenID {
    if buf.len() == 0 {
        return ScTokenID { id: [0; SC_TOKEN_ID_LENGTH] };
    }
    if buf.len() != SC_TOKEN_ID_LENGTH {
        panic("invalid TokenID length");
    }
    ScTokenID { id: buf.try_into().expect("WTF?") }
}

pub fn token_id_to_bytes(value: &ScTokenID) -> Vec<u8> {
    value.id.to_vec()
}

pub fn token_id_to_string(value: &ScTokenID) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableTokenID {
    proxy: Proxy,
}

impl ScImmutableTokenID {
    pub fn new(proxy: Proxy) -> ScImmutableTokenID {
        ScImmutableTokenID { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        token_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScTokenID {
        token_id_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScTokenID in host container
pub struct ScMutableTokenID {
    proxy: Proxy,
}

impl ScMutableTokenID {
    pub fn new(proxy: Proxy) -> ScMutableTokenID {
        ScMutableTokenID { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScTokenID) {
        self.proxy.set(&token_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        token_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScTokenID {
        token_id_from_bytes(&self.proxy.get())
    }
}
