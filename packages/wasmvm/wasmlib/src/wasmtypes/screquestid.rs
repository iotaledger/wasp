// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::wasmtypes::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_REQUEST_ID_LENGTH: usize = 34;

#[derive(PartialEq, Clone)]
pub struct ScRequestId {
    id: [u8; SC_REQUEST_ID_LENGTH],
}

impl ScRequestId {
    pub fn from_bytes(buf: &[u8]) -> ScRequestId {
        request_id_from_bytes(buf)
    }

    pub fn to_bytes(&self) -> &[u8] {
        request_id_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        request_id_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn request_id_decode(dec: &mut WasmDecoder) -> ScRequestId {
    request_id_from_bytes_unchecked(dec.fixed_bytes(SC_REQUEST_ID_LENGTH))
}

pub fn request_id_encode(enc: &mut WasmEncoder, value: &ScRequestId)  {
    enc.fixed_bytes(&value.to_bytes(), SC_REQUEST_ID_LENGTH);
}

pub fn request_id_from_bytes(buf: &[u8]) -> ScRequestId {
    ScRequestId { id: buf.try_into().expect("invalid RequestId length") }
}

pub fn request_id_to_bytes(value: &ScRequestId) -> &[u8] {
    &value.id
}

pub fn request_id_to_string(value: &ScRequestId) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

fn request_id_from_bytes_unchecked(buf: &[u8]) -> ScRequestId {
    ScRequestId { id: buf.try_into().expect("invalid RequestId length") }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableRequestId<'a> {
    proxy: Proxy<'a>,
}

impl ScImmutableRequestId<'_> {
    pub fn new(proxy: Proxy) -> ScImmutableRequestId {
        ScImmutableRequestId { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        request_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScRequestId {
        request_id_from_bytes(self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScRequestId in host container
pub struct ScMutableRequestId<'a> {
    proxy: Proxy<'a>,
}

impl ScMutableRequestId<'_> {
    pub fn new(proxy: Proxy) -> ScMutableRequestId {
        ScMutableRequestId { proxy }
    }

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScRequestId) {
        self.proxy.set(request_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        request_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScRequestId {
        request_id_from_bytes(self.proxy.get())
    }
}
