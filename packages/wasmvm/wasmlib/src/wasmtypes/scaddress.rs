// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_ADDRESS_ALIAS: u8 = 8;
pub const SC_ADDRESS_ED25519: u8 = 0;
pub const SC_ADDRESS_NFT: u8 = 16;
pub const SC_ADDRESS_ETH: u8 = 32;

pub const SC_LENGTH_ALIAS: usize = 33;
pub const SC_LENGTH_ED25519: usize = 33;
pub const SC_LENGTH_NFT: usize = 33;
pub const SC_LENGTH_ETH: usize = 20;

pub const SC_ADDRESS_LENGTH: usize = SC_LENGTH_ED25519;

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
    let len = buf.len();
    if len == 0 {
        return addr;
    }

    // special case, ETH address has no type byte but different length
    if len == SC_LENGTH_ETH {
        addr.id[0] = SC_ADDRESS_ETH;
        addr.id[1..SC_LENGTH_ETH + 1].copy_from_slice(&buf[..SC_LENGTH_ETH]);
        return addr;
    }

    match buf[0] {
        SC_ADDRESS_ALIAS => {
            if len != SC_LENGTH_ALIAS {
                panic("invalid Address length: Alias");
            }
            addr.id[..SC_LENGTH_ALIAS].copy_from_slice(&buf[..SC_LENGTH_ALIAS]);
        }
        SC_ADDRESS_ED25519 => {
            if len != SC_LENGTH_ED25519 {
                panic("invalid Address length: Ed25519");
            }
            addr.id[..SC_LENGTH_ED25519].copy_from_slice(&buf[..SC_LENGTH_ED25519]);
        }
        SC_ADDRESS_NFT => {
            if len != SC_LENGTH_NFT {
                panic("invalid Address length: NFT");
            }
            addr.id[..SC_LENGTH_NFT].copy_from_slice(&buf[..SC_LENGTH_NFT]);
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
            return value.id[1..SC_LENGTH_ETH + 1].to_vec();
        }
        _ => panic("unexpected Address type"),
    }
    Vec::new()
}

pub fn address_from_string(value: &str) -> ScAddress {
    if !value.starts_with("0x") {
        return bech32_decode(value);
    }

    // ETH address, allow the common "0x0"
    if value == "0x0" {
        return address_from_bytes(&vec![0; SC_LENGTH_ETH]);
    }

    let bytes = hex_decode(value);
    if bytes.len() != SC_LENGTH_ETH {
        panic("invalid ETH address");
    }
    address_from_bytes(&bytes)
}

pub fn address_to_string(value: &ScAddress) -> String {
    if value.id[0] != SC_ADDRESS_ETH {
        return bech32_encode(value);
    }

    let mut hex = string_to_bytes(&hex_encode(&address_to_bytes(value)));
    let hash = hash_keccak(&hex[2..]).to_bytes();
    for i in 2..hex.len() {
        let mut hash_byte = hash[(i - 2) >> 1];
        if (i & 0x01) == 0 {
            hash_byte >>= 4;
        } else {
            hash_byte &= 0x0f;
        }
        if hex[i] > 0x39 && hash_byte > 7 {
            hex[i] -= 32;
        }
    }
    string_from_bytes(&hex)
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
