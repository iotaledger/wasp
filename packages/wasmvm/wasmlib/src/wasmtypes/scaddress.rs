// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_ADDRESS_ALIAS: u8 = 8;
pub const SC_ADDRESS_ED25519: u8 = 0;
pub const SC_ADDRESS_NFT: u8 = 16;
pub const SC_ADDRESS_ETH: u8 = 32;

pub const SC_LENGTH_ALIAS: usize = 33;
pub const SC_LENGTH_ED25519: usize = 33;
pub const SC_LENGTH_NFT: usize = 33;

pub const SC_ADDRESS_LENGTH: usize = SC_LENGTH_ED25519;
pub const SC_ADDRESS_ETH_LENGTH: usize = 21;

#[derive(PartialEq, Clone)]
pub struct ScAddress {
    pub(crate) id: [u8; SC_ADDRESS_LENGTH],
}

impl ScAddress {
    pub fn as_agent_id(&self) -> ScAgentID {
        ScAgentID::from_address(self)
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        address_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        address_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

//TODO address type-dependent encoding/decoding?
pub fn address_decode(dec: &mut WasmDecoder) -> ScAddress {
    let buf = dec.fixed_bytes(SC_ADDRESS_LENGTH);
    ScAddress {
        id: buf.try_into().expect("WTF?"),
    }
}

pub fn address_encode(enc: &mut WasmEncoder, value: &ScAddress) {
    enc.fixed_bytes(&value.id, SC_ADDRESS_LENGTH);
}

pub fn address_from_bytes(buf: &[u8]) -> ScAddress {
    let mut addr = ScAddress {
        id: [0; SC_ADDRESS_LENGTH],
    };
    if buf.len() == 0 {
        return addr;
    }
    match buf[0] {
        SC_ADDRESS_ALIAS => {
            if buf.len() != SC_LENGTH_ALIAS {
                panic("invalid Address length: Alias");
            }
            addr.id[..SC_LENGTH_ALIAS].copy_from_slice(&buf[..SC_LENGTH_ALIAS]);
        }
        SC_ADDRESS_ED25519 => {
            if buf.len() != SC_LENGTH_ED25519 {
                panic("invalid Address length: Ed25519");
            }
            addr.id[..SC_LENGTH_ED25519].copy_from_slice(&buf[..SC_LENGTH_ED25519]);
        }
        SC_ADDRESS_NFT => {
            if buf.len() != SC_LENGTH_NFT {
                panic("invalid Address length: NFT");
            }
            addr.id[..SC_LENGTH_NFT].copy_from_slice(&buf[..SC_LENGTH_NFT]);
        }
        SC_ADDRESS_ETH => {
            if buf.len() != SC_ADDRESS_ETH_LENGTH {
                panic("invalid Address length: Eth");
            }
            addr.id[..SC_ADDRESS_ETH_LENGTH].copy_from_slice(&buf[..SC_ADDRESS_ETH_LENGTH]);
        }
        _ => panic("invalid Address type"),
    }
    addr
}

pub fn address_to_bytes(value: &ScAddress) -> Vec<u8> {
    match value.id[0] {
        SC_ADDRESS_ALIAS => {
            return value.id[..SC_LENGTH_ALIAS].to_vec();
        }
        SC_ADDRESS_ED25519 => {
            return value.id[..SC_LENGTH_ED25519].to_vec();
        }
        SC_ADDRESS_NFT => {
            return value.id[..SC_LENGTH_NFT].to_vec();
        }
        SC_ADDRESS_ETH => {
            return value.id[..SC_ADDRESS_ETH_LENGTH].to_vec();
        }
        _ => panic("unexpected Address type"),
    }
    Vec::new()
}

pub fn address_from_string(value: &str) -> ScAddress {
    if value.find("0x") == Some(0) {
        let mut b = vec![SC_ADDRESS_ETH];
        b.append(&mut hex_decode(&value));
        return address_from_bytes(&b);
    }
    let utils = ScSandboxUtils {};
    utils.bech32_decode(value)
}

pub fn address_to_string(value: &ScAddress) -> String {
    if value.id[0] == SC_ADDRESS_ETH {
        return hex_encode(&value.id[1..SC_ADDRESS_ETH_LENGTH]);
    }
    let utils = ScSandboxUtils {};
    utils.bech32_encode(value)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableAddress {
    proxy: Proxy,
}

impl ScImmutableAddress {
    pub fn new(proxy: Proxy) -> ScImmutableAddress {
        ScImmutableAddress { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        address_to_string(&self.value())
    }

    pub fn value(&self) -> ScAddress {
        address_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScAddress in host container
pub struct ScMutableAddress {
    proxy: Proxy,
}

impl ScMutableAddress {
    pub fn new(proxy: Proxy) -> ScMutableAddress {
        ScMutableAddress { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScAddress) {
        self.proxy.set(&address_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        address_to_string(&self.value())
    }

    pub fn value(&self) -> ScAddress {
        address_from_bytes(&self.proxy.get())
    }
}
