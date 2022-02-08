// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_CHAIN_ID_LENGTH: usize = 20;

#[derive(PartialEq, Clone)]
pub struct ScChainID {
    id: [u8; SC_CHAIN_ID_LENGTH],
}

impl ScChainID {
    pub fn new(buf: &[u8]) -> ScChainID {
        chain_id_from_bytes(buf)
    }

    pub fn address(&self) -> ScAddress {
        let mut buf = [0_u8; SC_ADDRESS_LENGTH];
        buf[0] = SC_ADDRESS_ALIAS;
        buf[1..SC_CHAIN_ID_LENGTH+1].copy_from_slice(&self.id);
        address_from_bytes(&buf)
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        chain_id_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        chain_id_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn chain_id_decode(dec: &mut WasmDecoder) -> ScChainID {
    chain_id_from_bytes_unchecked(&dec.fixed_bytes(SC_CHAIN_ID_LENGTH))
}

pub fn chain_id_encode(enc: &mut WasmEncoder, value: &ScChainID) {
    enc.fixed_bytes(&value.to_bytes(), SC_CHAIN_ID_LENGTH);
}

pub fn chain_id_from_bytes(buf: &[u8]) -> ScChainID {
    if buf.len() == 0 {
        let mut chain_id = ScChainID { id: [0; SC_CHAIN_ID_LENGTH] };
        chain_id.id[0] = SC_ADDRESS_ALIAS;
        return chain_id;
    }
    if buf.len() != SC_CHAIN_ID_LENGTH {
        panic("invalid ChainID length");
    }
    if buf[0] != SC_ADDRESS_ALIAS {
        panic("invalid ChainID: not an alias address");
    }
    ScChainID { id: buf.try_into().expect("WTF?") }
}

pub fn chain_id_to_bytes(value: &ScChainID) -> Vec<u8> {
    value.id.to_vec()
}

pub fn chain_id_to_string(value: &ScChainID) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

fn chain_id_from_bytes_unchecked(buf: &[u8]) -> ScChainID {
    ScChainID { id: buf.try_into().expect("invalid ChainID length") }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableChainID {
    proxy: Proxy,
}

impl ScImmutableChainID {
    pub fn new(proxy: Proxy) -> ScImmutableChainID {
        ScImmutableChainID { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        chain_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScChainID {
        chain_id_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScChainID in host container
pub struct ScMutableChainID {
    proxy: Proxy,
}

impl ScMutableChainID {
    pub fn new(proxy: Proxy) -> ScMutableChainID {
        ScMutableChainID { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScChainID) {
        self.proxy.set(&chain_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        chain_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScChainID {
        chain_id_from_bytes(&self.proxy.get())
    }
}
