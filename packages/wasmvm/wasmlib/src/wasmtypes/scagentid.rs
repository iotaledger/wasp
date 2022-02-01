// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::wasmtypes::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_AGENT_ID_LENGTH: usize = 37;

#[derive(PartialEq, Clone)]
pub struct ScAgentId {
    id: [u8; SC_AGENT_ID_LENGTH],
}

impl ScAgentId {
    pub fn from_bytes(buf: &[u8]) -> ScAgentId {
        agent_id_from_bytes(buf)
    }

    pub fn to_bytes(&self) -> &[u8] {
        agent_id_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        agent_id_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn agent_id_decode(dec: &mut WasmDecoder) -> ScAgentId {
    agent_id_from_bytes_unchecked(dec.fixed_bytes(SC_AGENT_ID_LENGTH))
}

pub fn agent_id_encode(enc: &mut WasmEncoder, value: &ScAgentId)  {
    enc.fixed_bytes(&value.to_bytes(), SC_AGENT_ID_LENGTH);
}

pub fn agent_id_from_bytes(buf: &[u8]) -> ScAgentId {
    ScAgentId { id: buf.try_into().expect("invalid AgentId length") }
}

pub fn agent_id_to_bytes(value: &ScAgentId) -> &[u8] {
    &value.id
}

pub fn agent_id_to_string(value: &ScAgentId) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

fn agent_id_from_bytes_unchecked(buf: &[u8]) -> ScAgentId {
    ScAgentId { id: buf.try_into().expect("invalid AgentId length") }
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

    pub fn value(&self) -> ScAgentId {
        agent_id_from_bytes(self.proxy.get())
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

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScAgentId) {
        self.proxy.set(agent_id_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        agent_id_to_string(&self.value())
    }

    pub fn value(&self) -> ScAgentId {
        agent_id_from_bytes(self.proxy.get())
    }
}
