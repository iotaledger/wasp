// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_CHAIN_ID_LENGTH: usize = 32;

#[derive(PartialEq, Copy, Clone)]
pub struct ScChainID {
    id: [u8; SC_CHAIN_ID_LENGTH],
}

impl ScChainID {
    pub fn address(&self) -> ScAddress {
        let buf = [SC_ADDRESS_ALIAS];
        address_from_bytes(&[&buf[..], &self.id[..]].concat())
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
    enc.fixed_bytes(&value.id, SC_CHAIN_ID_LENGTH);
}

pub fn chain_id_from_bytes(buf: &[u8]) -> ScChainID {
    if buf.len() == 0 {
        return ScChainID {
            id: [0; SC_CHAIN_ID_LENGTH],
        };
    }
    if buf.len() != SC_CHAIN_ID_LENGTH {
        panic("invalid ChainID length");
    }
    ScChainID {
        id: buf.try_into().expect("WTF?"),
    }
}

pub fn chain_id_to_bytes(value: &ScChainID) -> Vec<u8> {
    value.id.to_vec()
}

pub fn chain_id_from_string(value: &str) -> ScChainID {
    let addr = address_from_string(value);
    if addr.id[0] != SC_ADDRESS_ALIAS {
        panic("invalid ChainID address type");
    }
    chain_id_from_bytes(&addr.id[1..])
}

pub fn chain_id_to_string(value: &ScChainID) -> String {
    address_to_string(&value.address())
}

fn chain_id_from_bytes_unchecked(buf: &[u8]) -> ScChainID {
    ScChainID {
        id: buf.try_into().expect("invalid ChainID length"),
    }
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
