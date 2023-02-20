// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crypto::hashes::{blake2b::Blake2b256, Digest};
use wasmlib::*;

use crate::keypair::KeyPair;

pub const MAX_GAS_PER_BLOCK: u64 = 1_000_000_000;
pub const MIN_GAS_PER_REQUEST: u64 = 10000;
pub const MAX_GAS_PER_REQUEST: u64 = MAX_GAS_PER_BLOCK / 20;
pub const MAX_GAS_EXTERNAL_VIEW_CALL: u64 = MIN_GAS_PER_REQUEST;

#[derive(Clone)]
pub struct OffLedgerRequest {
    chain_id: ScChainID,
    contract: ScHname,
    entry_point: ScHname,
    params: Vec<u8>,
    signature_scheme: OffLedgerSignatureScheme,
    nonce: u64,
    allowance: ScAssets,
    gas_budget: u64,
}

#[derive(Clone)]
pub struct OffLedgerSignatureScheme {
    key_pair: KeyPair,
    pub signature: Vec<u8>,
}

impl OffLedgerSignatureScheme {
    pub fn new(key_pair: &KeyPair) -> Self {
        return OffLedgerSignatureScheme {
            key_pair: key_pair.clone(),
            signature: Vec::new(),
        };
    }
}

impl OffLedgerRequest {
    pub fn new(
        chain_id: &ScChainID,
        contract: &ScHname,
        entry_point: &ScHname,
        params: &[u8],
        nonce: u64,
    ) -> Self {
        return OffLedgerRequest {
            chain_id: chain_id.clone(),
            contract: contract.clone(),
            entry_point: entry_point.clone(),
            params: params.to_vec(),
            signature_scheme: OffLedgerSignatureScheme::new(&KeyPair::new(&[])),
            nonce: nonce,
            allowance: ScAssets::new(&[]),
            gas_budget: MAX_GAS_PER_REQUEST,
        };
    }

    pub fn essence(&self) -> Vec<u8> {
        let mut data: Vec<u8> = vec![1];
        data.extend(self.chain_id.to_bytes());
        data.extend(self.contract.to_bytes());
        data.extend(self.entry_point.to_bytes());
        data.extend(self.params.clone());
        data.extend(uint64_to_bytes(self.nonce));
        data.extend(uint64_to_bytes(self.gas_budget));
        let pub_key = self.signature_scheme.key_pair.public_key.to_bytes();
        data.push(pub_key.len() as u8);
        data.extend(pub_key);
        data.extend(self.allowance.to_bytes());
        return data;
    }

    pub fn id(&self) -> ScRequestID {
        // req id is hash of req bytes with output index zero
        let mut hash = Blake2b256::digest(self.to_bytes()).to_vec();
        hash.push(0);
        hash.push(0);
        return request_id_from_bytes(&hash);
    }

    pub fn sign(&self, key_pair: &KeyPair) -> Self {
        let mut req = OffLedgerRequest::new(
            &self.chain_id,
            &self.contract,
            &self.entry_point,
            &self.params,
            self.nonce,
        );
        req.signature_scheme = OffLedgerSignatureScheme::new(&key_pair);
        req.signature_scheme.signature = key_pair.sign(&req.essence());
        return req;
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut b = self.essence();
        let sig = &self.signature_scheme.signature;
        b.extend(uint16_to_bytes(sig.len() as u16));
        b.extend(sig);
        return b;
    }

    pub fn with_allowance(&mut self, allowance: &ScAssets) -> &Self {
        self.allowance = allowance.clone();
        return self;
    }
}
