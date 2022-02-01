// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::wasmtypes::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_CHAIN_ID_LENGTH: usize = 33;

#[derive(PartialEq, Clone)]
pub struct ScChainId {
    id: [u8; SC_CHAIN_ID_LENGTH],
}

impl ScChainId {
    pub fn from_bytes(buf: &[u8]) -> ScChainId {
        chain_id_from_bytes(buf)
    }

    pub fn to_bytes(&self) -> &[u8] {
        chain_id_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        chain_id_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn chain_id_decode(dec: &mut WasmDecoder) -> ScChainId {
    chain_id_from_bytes_unchecked(dec.fixed_bytes(SC_CHAIN_ID_LENGTH))
}

pub fn chain_id_encode(enc: &mut WasmEncoder, value: &ScChainId)  {
    enc.fixed_bytes(&value.to_bytes(), SC_CHAIN_ID_LENGTH);
}

pub fn chain_id_from_bytes(buf: &[u8]) -> ScChainId {
    ScChainId { id: buf.try_into().expect("invalid ChainId length") }
}

pub fn chain_id_to_bytes(value: &ScChainId) -> &[u8] {
    &value.id
}

pub fn chain_id_to_string(value: &ScChainId) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

fn chain_id_from_bytes_unchecked(buf: &[u8]) -> ScChainId {
    ScChainId { id: buf.try_into().expect("invalid ChainId length") }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableChainId<'a> {
    proxy: Proxy<'a>,
}

impl ScImmutableChainId<'_> {
    pub fn new(proxy: Proxy) -> ScImmutableChainId {
        ScImmutableChainId { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        chain_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScChainId {
        chain_id_from_bytes(self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScChainId in host container
pub struct ScMutableChainId<'a> {
    proxy: Proxy<'a>,
}

impl ScMutableChainId<'_> {
    pub fn new(proxy: Proxy) -> ScMutableChainId {
        ScMutableChainId { proxy }
    }

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScChainId) {
        self.proxy.set(chain_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        chain_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScChainId {
        chain_id_from_bytes(self.proxy.get())
    }
}
