// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::num::Wrapping;
use std::ops::Add;
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
        let mut enc = WasmEncoder::new();
        self.essence_encode(&mut enc);
        enc.buf()
    }

    fn essence_encode(&self, mut enc: &mut WasmEncoder) {
        enc.byte(1); // requestKindOffLedgerISC
        chain_id_encode(&mut enc, &self.chain_id);
        hname_encode(&mut enc, self.contract);
        hname_encode(&mut enc, self.entry_point);
        enc.fixed_bytes(&self.params, self.params.len());
        enc.vlu_encode(self.nonce);
        let gas_budget = Wrapping(self.gas_budget).add(Wrapping(1)).0;
        enc.vlu_encode(gas_budget);
        let allowance = self.allowance.to_bytes();
        enc.fixed_bytes(&allowance, allowance.len());
    }

    pub fn id(&self) -> ScRequestID {
        // req id is hash of req bytes with output index zero
        let mut hash = Blake2b256::digest(self.to_bytes()).to_vec();
        hash.push(0);
        hash.push(0);
        return request_id_from_bytes(&hash);
    }

    pub fn sign(&mut self, key_pair: &KeyPair) {
        self.signature = OffLedgerSignature::new(&key_pair.public_key);
        let hash = Blake2b256::digest(&self.essence());
        self.signature.signature = key_pair.sign(&hash);
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        self.essence_encode(&mut enc);
        enc.fixed_bytes(&self.signature.public_key.to_bytes(), 32);
        enc.bytes(&self.signature.signature);
        enc.buf()
    }

    pub fn with_allowance(&mut self, allowance: &ScAssets) -> &Self {
        self.allowance = allowance.clone();
        return self;
    }
}
