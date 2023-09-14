// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

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
    eth: ScAddress,
}

impl ScAgentID {
    pub fn new(address: &ScAddress, hname: ScHname) -> ScAgentID {
        ScAgentID {
            kind: SC_AGENT_ID_CONTRACT,
            address: address.clone(),
            hname: hname,
            eth: ScAddress {
                id: [0; SC_ADDRESS_LENGTH],
            },
        }
    }

    pub fn for_ethereum(chain: &ScAddress, eth_address: &ScAddress) -> ScAgentID {
        if chain.id[0] != SC_ADDRESS_ALIAS {
            panic("invalid eth AgentID: chain address");
        }
        if eth_address.id[0] != SC_ADDRESS_ETH {
            panic("invalid eth AgentID: eth address");
        }
        ScAgentID {
            kind: SC_AGENT_ID_ETHEREUM,
            address: chain.clone(),
            hname: ScHname(0),
            eth: eth_address.clone(),
        }
    }

    pub fn from_address(address: &ScAddress) -> ScAgentID {
        let kind;
        match address.id[0] {
            SC_ADDRESS_ALIAS => {
                kind = SC_AGENT_ID_CONTRACT;
            }
            SC_ADDRESS_ETH => {
                panic("invalid eth AgentID: need chain address");
                kind = 0;
            }
            _ => {
                kind = SC_AGENT_ID_ADDRESS;
            }
        }

        ScAgentID {
            kind: kind,
            address: address.clone(),
            hname: ScHname(0),
            eth: ScAddress {
                id: [0; SC_ADDRESS_LENGTH],
            },
        }
    }

    pub fn address(&self) -> ScAddress {
        self.address.clone()
    }

    pub fn eth_address(&self) -> ScAddress {
        self.eth.clone()
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
    let len = buf.len();
    if len == 0 {
        return ScAgentID {
            kind: SC_AGENT_ID_NIL,
            address: address_from_bytes(buf),
            hname: ScHname(0),
            eth: ScAddress {
                id: [0; SC_ADDRESS_LENGTH],
            },
        };
    }
    let len = len - 1;
    match buf[0] {
        SC_AGENT_ID_ADDRESS => {
            let buf: &[u8] = &buf[1..];
            if len != SC_LENGTH_ALIAS && len != SC_LENGTH_ED25519 {
                panic("invalid AgentID length: address agentID");
            }
            return ScAgentID::from_address(&address_from_bytes(&buf));
        }
        SC_AGENT_ID_CONTRACT => {
            let buf: &[u8] = &buf[1..];
            if len != SC_CHAIN_ID_LENGTH + SC_HNAME_LENGTH {
                panic("invalid AgentID length: contract agentID");
            }
            let chain_id = chain_id_from_bytes(&buf[..SC_CHAIN_ID_LENGTH]);
            let hname = hname_from_bytes(&buf[SC_CHAIN_ID_LENGTH..]);
            return ScAgentID::new(&chain_id.address(), hname);
        }
        SC_AGENT_ID_ETHEREUM => {
            let buf: &[u8] = &buf[1..];
            if len != SC_CHAIN_ID_LENGTH + SC_LENGTH_ETH {
                panic("invalid AgentID length: eth agentID");
            }
            let chain_id = chain_id_from_bytes(&buf[..SC_CHAIN_ID_LENGTH]);
            let eth_address = address_from_bytes(&buf[SC_CHAIN_ID_LENGTH..]);
            return ScAgentID::for_ethereum(&chain_id.address(), &eth_address);
        }
        SC_AGENT_ID_NIL => {}
        _ => panic("AgentIDFromBytes: invalid AgentID type"),
    }
    ScAgentID {
        kind: SC_AGENT_ID_NIL,
        address: address_from_bytes(&[]),
        hname: ScHname(0),
        eth: ScAddress {
            id: [0; SC_ADDRESS_LENGTH],
        },
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
        SC_AGENT_ID_ETHEREUM => {
            buf.extend_from_slice(&address_to_bytes(&value.address)[1..]);
            buf.extend_from_slice(&address_to_bytes(&value.eth));
        }
        SC_AGENT_ID_NIL => (),
        _ => panic("AgentIDToBytes: invalid AgentID type"),
    }
    buf
}

pub fn agent_id_from_string(value: &str) -> ScAgentID {
    if value.eq(NIL_AGENT_ID_STRING) {
        return agent_id_from_bytes(&[]);
    }

    let parts: Vec<&str> = value.split("@").collect();
    return match parts.len() {
        1 => ScAgentID::from_address(&address_from_string(&parts[0])),
        2 => {
            if !value.starts_with("0x") {
                return ScAgentID::new(
                    &address_from_string(&parts[1]),
                    hname_from_string(&parts[0]),
                );
            }
            return ScAgentID::for_ethereum(
                &address_from_string(&parts[1]),
                &address_from_string(&parts[0]),
            );
        },
        _ => {
            panic("invalid AgentID string");
            agent_id_from_bytes(&[])
        }
    };
}

pub fn agent_id_to_string(value: &ScAgentID) -> String {
    match value.kind {
        SC_AGENT_ID_ADDRESS => {
            return value.address().to_string();
        }
        SC_AGENT_ID_CONTRACT => {
            return value.hname().to_string() + "@" + &value.address().to_string();
        }
        SC_AGENT_ID_ETHEREUM => {
            return value.eth_address().to_string() + "@" + &value.address().to_string();
        }
        SC_AGENT_ID_NIL => {
            return NIL_AGENT_ID_STRING.to_string();
        }
        _ => panic("AgentIDToString: invalid AgentID type"),
    }
    "".to_string()
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
