// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_AGENT_ID_NIL: u8 = 0;
pub const SC_AGENT_ID_ADDRESS: u8 = 1;
pub const SC_AGENT_ID_CONTRACT: u8 = 2;
pub const SC_AGENT_ID_ETHEREUM: u8 = 3;
const NIL_AGENT_ID_STRING: &str = "-";

#[derive(PartialEq, Clone)]
pub struct ScAgentID {
    kind: u8,
    address: ScAddress,
    hname: ScHname,
}

impl ScAgentID {
    pub fn new(address: &ScAddress, hname: ScHname) -> ScAgentID {
        ScAgentID {
            kind: SC_AGENT_ID_CONTRACT,
            address: address.clone(),
            hname: hname,
        }
    }

    pub fn from_address(address: &ScAddress) -> ScAgentID {
        let mut kind = SC_AGENT_ID_ADDRESS;
        if address.id[0] == SC_ADDRESS_ALIAS {
            kind = SC_AGENT_ID_CONTRACT;
        }
        ScAgentID {
            kind: kind,
            address: address.clone(),
            hname: ScHname(0),
        }
    }

    pub fn address(&self) -> ScAddress {
        self.address.clone()
    }

    pub fn hname(&self) -> ScHname {
        self.hname
    }

    pub fn is_address(&self) -> bool {
        self.kind == SC_AGENT_ID_ADDRESS
    }

    pub fn is_contracts(&self) -> bool {
        self.kind == SC_AGENT_ID_CONTRACT
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
    agent_id_from_bytes(&dec.bytes())
}

pub fn agent_id_encode(enc: &mut WasmEncoder, value: &ScAgentID) {
    enc.bytes(&agent_id_to_bytes(value));
}

pub fn agent_id_from_bytes(buf: &[u8]) -> ScAgentID {
    if buf.len() == 0 {
        return ScAgentID {
            kind: SC_AGENT_ID_NIL,
            address: address_from_bytes(buf),
            hname: ScHname(0),
        };
    }
    match buf[0] {
        SC_AGENT_ID_ADDRESS => {
            let buf: &[u8] = &buf[1..];
            if buf.len() != SC_LENGTH_ALIAS && buf.len() != SC_LENGTH_ED25519 {
                panic("invalid AgentID length: address agendID");
            }
            return ScAgentID::from_address(&address_from_bytes(&buf));
        }
        SC_AGENT_ID_CONTRACT => {
            let buf: &[u8] = &buf[1..];
            if buf.len() != SC_CHAIN_ID_LENGTH + SC_HNAME_LENGTH {
                panic("invalid AgentID length: contract agendID");
            }
            let chain_id = chain_id_from_bytes(&buf[..SC_CHAIN_ID_LENGTH]);
            let hname = hname_from_bytes(&buf[SC_CHAIN_ID_LENGTH..]);
            return ScAgentID::new(&chain_id.address(), hname);
        }
        SC_AGENT_ID_ETHEREUM => panic("AgentIDFromBytes: unsupported ScAgentIDEthereum"),
        SC_AGENT_ID_NIL => {}
        _ => panic("AgentIDFromBytes: invalid AgentID type"),
    }
    ScAgentID {
        kind: SC_AGENT_ID_NIL,
        address: address_from_bytes(&[]),
        hname: ScHname(0),
    }
}

pub fn agent_id_to_bytes(value: &ScAgentID) -> Vec<u8> {
    let mut buf: Vec<u8> = Vec::new();
    buf.push(value.kind);
    match value.kind {
        SC_AGENT_ID_ADDRESS => {
            buf.extend_from_slice(&address_to_bytes(&value.address));
        }
        SC_AGENT_ID_CONTRACT => {
            buf.extend_from_slice(&address_to_bytes(&value.address)[1..]);
            buf.extend_from_slice(&hname_to_bytes(value.hname));
        }
        SC_AGENT_ID_ETHEREUM => panic("AgentIDToBytes: unsupported ScAgentIDEthereum"),
        SC_AGENT_ID_NIL => (),
        _ => panic("AgentIDToBytes: invalid AgentID type"),
    }
    buf
}

pub fn agent_id_from_string(value: &str) -> ScAgentID {
    //TODO ScAgentIDEthereum
    if value.eq(NIL_AGENT_ID_STRING) {
        return agent_id_from_bytes(&[]);
    }

    let parts: Vec<&str> = value.split("@").collect();
    match parts.len() {
        1 => return ScAgentID::from_address(&address_from_string(&parts[0])),
        2 => {
            return ScAgentID::new(
                &address_from_string(&parts[1]),
                hname_from_string(&parts[0]),
            )
        }
        _ => {
            panic("invalid AgentID string");
            return agent_id_from_bytes(&[]);
        }
    }
}

pub fn agent_id_to_string(value: &ScAgentID) -> String {
    //TODO ScAgentIDEthereum
    match value.kind {
        SC_AGENT_ID_ADDRESS => {
            return value.address().to_string();
        }
        SC_AGENT_ID_CONTRACT => {
            return value.hname().to_string() + "@" + &value.address().to_string();
        }
        SC_AGENT_ID_ETHEREUM => panic("AgentIDToString: unsupported ScAgentIDEthereum"),
        SC_AGENT_ID_NIL => (),
        _ => panic("AgentIDToString: invalid AgentID type"),
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableAgentID {
    proxy: Proxy,
}

impl ScImmutableAgentID {
    pub fn new(proxy: Proxy) -> ScImmutableAgentID {
        ScImmutableAgentID { proxy }
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

// value proxy for mutable ScAgentID in host container
pub struct ScMutableAgentID {
    proxy: Proxy,
}

impl ScMutableAgentID {
    pub fn new(proxy: Proxy) -> ScMutableAgentID {
        ScMutableAgentID { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScAgentID) {
        self.proxy.set(&agent_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        agent_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScAgentID {
        agent_id_from_bytes(&self.proxy.get())
    }
}
