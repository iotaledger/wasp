// Code generated by schema tool; DO NOT EDIT.

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

#![allow(dead_code)]
#![allow(unused_imports)]

use crate::*;

#[derive(Clone)]
pub struct CallRequest {
    // caller assets that the call is allowed to access
    pub allowance : Vec<u8>,
    pub contract  : ScHname,
    pub function  : ScHname,
    pub params    : Vec<u8>,
}

impl CallRequest {
    pub fn from_bytes(bytes: &[u8]) -> CallRequest {
        let mut dec = WasmDecoder::new(bytes);
        CallRequest {
            allowance : bytes_decode(&mut dec),
            contract  : hname_decode(&mut dec),
            function  : hname_decode(&mut dec),
            params    : bytes_decode(&mut dec),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        bytes_encode(&mut enc, &self.allowance);
        hname_encode(&mut enc, self.contract);
        hname_encode(&mut enc, self.function);
        bytes_encode(&mut enc, &self.params);
        enc.buf()
    }
}

#[derive(Clone)]
pub struct ImmutableCallRequest {
    pub(crate) proxy: Proxy,
}

impl ImmutableCallRequest {
    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn value(&self) -> CallRequest {
        CallRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct MutableCallRequest {
    pub(crate) proxy: Proxy,
}

impl MutableCallRequest {
    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &CallRequest) {
        self.proxy.set(&value.to_bytes());
    }

    pub fn value(&self) -> CallRequest {
        CallRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct DeployRequest {
    pub description : String,
    pub name        : String,
    pub params      : Vec<u8>,
    pub prog_hash   : ScHash,
}

impl DeployRequest {
    pub fn from_bytes(bytes: &[u8]) -> DeployRequest {
        let mut dec = WasmDecoder::new(bytes);
        DeployRequest {
            description : string_decode(&mut dec),
            name        : string_decode(&mut dec),
            params      : bytes_decode(&mut dec),
            prog_hash   : hash_decode(&mut dec),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        string_encode(&mut enc, &self.description);
        string_encode(&mut enc, &self.name);
        bytes_encode(&mut enc, &self.params);
        hash_encode(&mut enc, &self.prog_hash);
        enc.buf()
    }
}

#[derive(Clone)]
pub struct ImmutableDeployRequest {
    pub(crate) proxy: Proxy,
}

impl ImmutableDeployRequest {
    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn value(&self) -> DeployRequest {
        DeployRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct MutableDeployRequest {
    pub(crate) proxy: Proxy,
}

impl MutableDeployRequest {
    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &DeployRequest) {
        self.proxy.set(&value.to_bytes());
    }

    pub fn value(&self) -> DeployRequest {
        DeployRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct PostRequest {
    // caller assets that the call is allowed to access
    pub allowance : Vec<u8>,
    pub chain_id  : ScChainID,
    pub contract  : ScHname,
    pub delay     : u32,
    pub function  : ScHname,
    pub params    : Vec<u8>,
    // assets that are transferred into caller account
    pub transfer  : Vec<u8>,
}

impl PostRequest {
    pub fn from_bytes(bytes: &[u8]) -> PostRequest {
        let mut dec = WasmDecoder::new(bytes);
        PostRequest {
            allowance : bytes_decode(&mut dec),
            chain_id  : chain_id_decode(&mut dec),
            contract  : hname_decode(&mut dec),
            delay     : uint32_decode(&mut dec),
            function  : hname_decode(&mut dec),
            params    : bytes_decode(&mut dec),
            transfer  : bytes_decode(&mut dec),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        bytes_encode(&mut enc, &self.allowance);
        chain_id_encode(&mut enc, &self.chain_id);
        hname_encode(&mut enc, self.contract);
        uint32_encode(&mut enc, self.delay);
        hname_encode(&mut enc, self.function);
        bytes_encode(&mut enc, &self.params);
        bytes_encode(&mut enc, &self.transfer);
        enc.buf()
    }
}

#[derive(Clone)]
pub struct ImmutablePostRequest {
    pub(crate) proxy: Proxy,
}

impl ImmutablePostRequest {
    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn value(&self) -> PostRequest {
        PostRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct MutablePostRequest {
    pub(crate) proxy: Proxy,
}

impl MutablePostRequest {
    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &PostRequest) {
        self.proxy.set(&value.to_bytes());
    }

    pub fn value(&self) -> PostRequest {
        PostRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct SendRequest {
    pub address  : ScAddress,
    pub transfer : Vec<u8>,
}

impl SendRequest {
    pub fn from_bytes(bytes: &[u8]) -> SendRequest {
        let mut dec = WasmDecoder::new(bytes);
        SendRequest {
            address  : address_decode(&mut dec),
            transfer : bytes_decode(&mut dec),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        address_encode(&mut enc, &self.address);
        bytes_encode(&mut enc, &self.transfer);
        enc.buf()
    }
}

#[derive(Clone)]
pub struct ImmutableSendRequest {
    pub(crate) proxy: Proxy,
}

impl ImmutableSendRequest {
    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn value(&self) -> SendRequest {
        SendRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct MutableSendRequest {
    pub(crate) proxy: Proxy,
}

impl MutableSendRequest {
    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &SendRequest) {
        self.proxy.set(&value.to_bytes());
    }

    pub fn value(&self) -> SendRequest {
        SendRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct TransferRequest {
    pub agent_id : ScAgentID,
    pub transfer : Vec<u8>,
}

impl TransferRequest {
    pub fn from_bytes(bytes: &[u8]) -> TransferRequest {
        let mut dec = WasmDecoder::new(bytes);
        TransferRequest {
            agent_id : agent_id_decode(&mut dec),
            transfer : bytes_decode(&mut dec),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        agent_id_encode(&mut enc, &self.agent_id);
        bytes_encode(&mut enc, &self.transfer);
        enc.buf()
    }
}

#[derive(Clone)]
pub struct ImmutableTransferRequest {
    pub(crate) proxy: Proxy,
}

impl ImmutableTransferRequest {
    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn value(&self) -> TransferRequest {
        TransferRequest::from_bytes(&self.proxy.get())
    }
}

#[derive(Clone)]
pub struct MutableTransferRequest {
    pub(crate) proxy: Proxy,
}

impl MutableTransferRequest {
    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &TransferRequest) {
        self.proxy.set(&value.to_bytes());
    }

    pub fn value(&self) -> TransferRequest {
        TransferRequest::from_bytes(&self.proxy.get())
    }
}
