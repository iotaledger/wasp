// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_REQUEST_ID_LENGTH: usize = 34;
pub const REQUEST_ID_SEPARATOR: &str = "-";

#[derive(PartialEq, Clone)]
pub struct ScRequestID {
    id: [u8; SC_REQUEST_ID_LENGTH],
}

impl ScRequestID {
    pub fn to_bytes(&self) -> Vec<u8> {
        request_id_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        request_id_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn request_id_decode(dec: &mut WasmDecoder) -> ScRequestID {
    request_id_from_bytes_unchecked(&dec.fixed_bytes(SC_REQUEST_ID_LENGTH))
}

pub fn request_id_encode(enc: &mut WasmEncoder, value: &ScRequestID) {
    enc.fixed_bytes(&value.id, SC_REQUEST_ID_LENGTH);
}

pub fn request_id_from_bytes(buf: &[u8]) -> ScRequestID {
    if buf.len() == 0 {
        return ScRequestID {
            id: [0; SC_REQUEST_ID_LENGTH],
        };
    }
    if buf.len() != SC_REQUEST_ID_LENGTH {
        panic("invalid RequestID length");
    }
    // final uint16 output index must be > ledgerstate.MaxOutputCount
    if buf[SC_REQUEST_ID_LENGTH - 2] > 127 || buf[SC_REQUEST_ID_LENGTH - 1] != 0 {
        panic("invalid RequestID: output index > 127");
    }
    ScRequestID {
        id: buf.try_into().expect("WTF?"),
    }
}

pub fn request_id_to_bytes(value: &ScRequestID) -> Vec<u8> {
    value.id.to_vec()
}

pub fn request_id_from_string(value: &str) -> ScRequestID {
    request_id_from_bytes(&hex_decode(value))
}

pub fn request_id_to_string(value: &ScRequestID) -> String {
    hex_encode(&value.id)
}

fn request_id_from_bytes_unchecked(buf: &[u8]) -> ScRequestID {
    ScRequestID {
        id: buf.try_into().expect("invalid RequestID length"),
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableRequestID {
    proxy: Proxy,
}

impl ScImmutableRequestID {
    pub fn new(proxy: Proxy) -> ScImmutableRequestID {
        ScImmutableRequestID { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        request_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScRequestID {
        request_id_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScRequestID in host container
pub struct ScMutableRequestID {
    proxy: Proxy,
}

impl ScMutableRequestID {
    pub fn new(proxy: Proxy) -> ScMutableRequestID {
        ScMutableRequestID { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScRequestID) {
        self.proxy.set(&request_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        request_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScRequestID {
        request_id_from_bytes(&self.proxy.get())
    }
}
