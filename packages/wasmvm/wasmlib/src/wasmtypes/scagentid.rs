// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;
use crate::sandbox::*;
use crate::wasmtypes::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_AGENT_ID_LENGTH: usize = 37;

#[derive(PartialEq, Clone)]
pub struct ScAgentID {
    address: ScAddress,
    hname: ScHname,
}

impl ScAgentID {
    pub fn new(address: &ScAddress, hname: &ScHname) -> ScAgentID {
        ScAgentID { address: address_from_bytes(&address.to_bytes()), hname: hname_from_bytes(&hname.to_bytes()) }
    }

    pub fn from_bytes(buf: &[u8]) -> ScAgentID {
        agent_id_from_bytes(buf)
    }

    pub fn address(&self) -> &ScAddress {
        &self.address
    }

    pub fn hname(&self) -> &ScHname {
        &self.hname
    }

    pub fn is_address(&self) -> bool {
        self.hname == ScHname(0)
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        agent_id_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        agent_id_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn agent_id_decode(dec: &mut WasmDecoder) -> ScAgentID {
    ScAgentID { address: address_decode(dec), hname: hname_decode(dec) }
}

pub fn agent_id_encode(enc: &mut WasmEncoder, value: &ScAgentID) {
    address_encode(enc, value.address());
    hname_encode(enc, value.hname());
}

pub fn agent_id_from_bytes(buf: &[u8]) -> ScAgentID {
    if buf.len() == 0 {
        return ScAgentID { address: address_from_bytes(buf), hname: hname_from_bytes(buf) };
    }
    if buf.len() != SC_AGENT_ID_LENGTH {
        panic("invalid AgentId length");
    }
    // max ledgerstate.AliasAddressType
    if buf[0] > 2 {
        panic("invalid AgentID: address type > 2");
    }
    ScAgentID {
        address: address_from_bytes(&buf[..SC_ADDRESS_LENGTH]),
        hname: hname_from_bytes(&buf[SC_ADDRESS_LENGTH..]),
    }
}

pub fn agent_id_to_bytes(value: &ScAgentID) -> Vec<u8> {
    let mut enc = WasmEncoder::new();
    agent_id_encode(&mut enc, value);
    enc.buf()
}

pub fn agent_id_to_string(value: &ScAgentID) -> String {
    // TODO standardize human readable string
    value.address.to_string() + "::" + &value.hname.to_string()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableAgentId<'a> {
    proxy: Proxy<'a>,
}

impl ScImmutableAgentId<'_> {
    pub fn new(proxy: Proxy) -> ScImmutableAgentId {
        ScImmutableAgentId { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        agent_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScAgentID {
        agent_id_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScAgentId in host container
pub struct ScMutableAgentId<'a> {
    proxy: Proxy<'a>,
}

impl ScMutableAgentId<'_> {
    pub fn new(proxy: Proxy) -> ScMutableAgentId {
        ScMutableAgentId { proxy }
    }

    pub fn delete(&mut self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&mut self, value: &ScAgentID) {
        self.proxy.set(&agent_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        agent_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScAgentID {
        agent_id_from_bytes(&self.proxy.get())
    }
}
