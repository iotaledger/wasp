// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crypto::{
    hashes::{blake2b::Blake2b256, Digest},
    signatures::ed25519,
};
use wasmlib::*;

use crate::keypair::KeyPair;

#[derive(Clone)]
pub struct OffLedgerRequest {
    chain_id: ScChainID,
    contract: ScHname,
    entry_point: ScHname,
    params: Vec<u8>,
    signature: OffLedgerSignature,
    nonce: u64,
    allowance: ScAssets,
    gas_budget: u64,
}

#[derive(Clone)]
pub struct OffLedgerSignature {
    public_key: ed25519::PublicKey,
    signature: Vec<u8>,
}

impl OffLedgerSignature {
    pub fn new(public_key: &ed25519::PublicKey) -> Self {
        return OffLedgerSignature {
            public_key: public_key.clone(),
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
            signature: OffLedgerSignature::new(&KeyPair::new(&[]).public_key),
            nonce,
            allowance: ScAssets::new(&[]),
            gas_budget: u64::MAX,
        };
    }

    pub fn essence(&self) -> Vec<u8> {
        let mut data: Vec<u8> = vec![1]; // requestKindTagOffLedgerISC
        data.extend(self.chain_id.to_bytes());
        data.extend(self.contract.to_bytes());
        data.extend(self.entry_point.to_bytes());
        data.extend(self.params.clone());
        data.extend(uint64_to_bytes(self.nonce));
        data.extend(uint64_to_bytes(self.gas_budget));
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
        req.signature = OffLedgerSignature::new(&key_pair.public_key);
        req.signature.signature = key_pair.sign(&req.essence());
        return req;
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut data = self.essence();
        let public_key = self.signature.public_key.to_bytes();
        data.extend(uint8_to_bytes(public_key.len() as u8));
        data.extend(public_key);
        let signature = &self.signature.signature;
        data.extend(uint16_to_bytes(signature.len() as u16));
        data.extend(signature);
        return data;
    }

    pub fn with_allowance(&mut self, allowance: &ScAssets) -> &Self {
        self.allowance = allowance.clone();
        return self;
    }
}
