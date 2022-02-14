// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const NIL_AGENT_ID: u8 = 0xff;

#[derive(PartialEq, Clone)]
pub struct ScAgentID {
    address: ScAddress,
    hname: ScHname,
}

const NIL_AGENT: ScAgentID = ScAgentID {
    address: ScAddress { id: [0; SC_ADDRESS_LENGTH] },
    hname: ScHname(0),
};

impl ScAgentID {
    pub fn new(address: &ScAddress, hname: ScHname) -> ScAgentID {
        ScAgentID {
            address: address_from_bytes(&address.to_bytes()),
            hname: hname_from_bytes(&hname.to_bytes()),
        }
    }

    pub fn address(&self) -> ScAddress {
        self.address.clone()
    }

    pub fn hname(&self) -> ScHname {
        self.hname
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

// note: only alias address can have a non-zero hname
// so there is no need to encode it when it is always zero

pub fn agent_id_decode(dec: &mut WasmDecoder) -> ScAgentID {
    if dec.peek() == SC_ADDRESS_ALIAS {
        return ScAgentID { address: address_decode(dec), hname: hname_decode(dec) };
    }
    ScAgentID { address: address_decode(dec), hname: ScHname(0) }
}

pub fn agent_id_encode(enc: &mut WasmEncoder, value: &ScAgentID) {
    address_encode(enc, &value.address());
    if value.address.to_bytes()[0] == SC_ADDRESS_ALIAS {
        hname_encode(enc, value.hname());
    }
}

pub fn agent_id_from_bytes(buf: &[u8]) -> ScAgentID {
    if buf.len() == 0 {
        return ScAgentID {
            address: address_from_bytes(buf),
            hname: ScHname(0),
        };
    }
    match buf[0] {
        SC_ADDRESS_ALIAS => {
            if buf.len() != SC_LENGTH_ALIAS + SC_HNAME_LENGTH {
                panic("invalid AgentID length: Alias address");
            }
            return ScAgentID {
                address: address_from_bytes(&buf[..SC_LENGTH_ALIAS]),
                hname: hname_from_bytes(&buf[SC_LENGTH_ALIAS..]),
            };
        }
        SC_ADDRESS_ED25519 => {
            if buf.len() != SC_LENGTH_ED25519 {
                panic("invalid AgentID length: Ed25519 address");
            }
            return ScAgentID {
                address: address_from_bytes(buf),
                hname: ScHname(0),
            };
        }
        SC_ADDRESS_NFT => {
            if buf.len() != SC_LENGTH_NFT {
                panic("invalid AgentID length: NFT address");
            }
            return ScAgentID {
                address: address_from_bytes(buf),
                hname: ScHname(0),
            };
        }
        NIL_AGENT_ID => {
            if buf.len() != 1 {
                panic("invalid AgentID length: nil AgentID");
            }
        }
        _ =>
            panic("invalid AgentID address type"),
    }
    ScAgentID {
        address: address_from_bytes(&[]),
        hname: ScHname(0),
    }
}

pub fn agent_id_to_bytes(value: &ScAgentID) -> Vec<u8> {
    if *value == NIL_AGENT {
        return [NIL_AGENT_ID].to_vec();
    }
    let mut buf = address_to_bytes(&value.address);
    if buf[0] == SC_ADDRESS_ALIAS {
        buf.extend_from_slice(&hname_to_bytes(value.hname));
    }
    buf
}

pub fn agent_id_to_string(value: &ScAgentID) -> String {
    // TODO standardize human readable string
    value.address.to_string() + "::" + &value.hname.to_string()
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
