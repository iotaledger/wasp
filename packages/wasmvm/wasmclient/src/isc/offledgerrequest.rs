// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::keypair;
use wasmlib::*;

pub trait OffLedgerRequest {
    fn new(
        chain_id: &ScChainID,
        contract: &ScHname,
        entry_point: &ScHname,
        params: &ScDict,
        nonce: u64,
    ) -> Self;
    fn with_nonce(&mut self, nonce: u64) -> &Self;
    fn with_gas_budget(&mut self, gas_budget: u64) -> &Self;
    fn with_allowance(&mut self, allowance: &ScAssets) -> &Self;
    fn sign(&self, key: &keypair::KeyPair) -> Self;
}

#[derive(Clone)]
pub struct OffLedgerRequestData {
    chain_id: ScChainID,
    contract: ScHname,
    entry_point: ScHname,
    params: ScDict,
    signature_scheme: Option<OffLedgerSignatureScheme>, // None if unsigned
    nonce: u64,
    allowance: ScAssets,
    gas_budget: u64,
}

#[derive(Clone)]
pub struct OffLedgerSignatureScheme {
    key_pair: keypair::KeyPair,
    pub signature: Vec<u8>,
}

impl OffLedgerSignatureScheme {
    pub fn new(key_pair: &keypair::KeyPair) -> Self {
        return OffLedgerSignatureScheme {
            key_pair: key_pair.clone(),
            signature: Vec::new(),
        };
    }
}

impl OffLedgerRequest for OffLedgerRequestData {
    fn new(
        chain_id: &ScChainID,
        contract: &ScHname,
        entry_point: &ScHname,
        params: &ScDict,
        nonce: u64,
    ) -> Self {
        return OffLedgerRequestData {
            chain_id: chain_id.clone(),
            contract: contract.clone(),
            entry_point: entry_point.clone(),
            params: params.clone(),
            signature_scheme: None,
            nonce: nonce,
            allowance: ScAssets::new(&Vec::new()),
            gas_budget: super::gas::MAX_GAS_PER_REQUEST,
        };
    }
    fn with_nonce(&mut self, nonce: u64) -> &Self {
        self.nonce = nonce;
        return self;
    }
    fn with_gas_budget(&mut self, gas_budget: u64) -> &Self {
        self.gas_budget = gas_budget;
        return self;
    }
    fn with_allowance(&mut self, allowance: &ScAssets) -> &Self {
        self.allowance = allowance.clone();
        return self;
    }
    fn sign<'a>(&self, key_pair: &'a keypair::KeyPair) -> Self {
        let mut req = OffLedgerRequestData::new(
            &self.chain_id,
            &self.contract,
            &self.entry_point,
            &self.params,
            self.nonce,
        );
        let mut scheme = OffLedgerSignatureScheme::new(&key_pair.to_owned());
        scheme.signature = key_pair.clone().sign(&self.essence()).clone();
        req.signature_scheme = Some(scheme);
        return req;
    }
}

use crypto::hashes::{blake2b::Blake2b256, Digest};

impl OffLedgerRequestData {
    pub fn id(&self) -> ScRequestID {
        let hash = Blake2b256::digest(self.to_bytes());
        return wasmlib::request_id_from_bytes(&hash.to_vec());
    }

    pub fn essence(&self) -> Vec<u8> {
        let mut data: Vec<u8> = vec![1];
        data.append(self.chain_id.to_bytes().as_mut());
        data.append(self.contract.to_bytes().as_mut());
        data.append(self.entry_point.to_bytes().as_mut());
        data.append(self.params.to_bytes().as_mut());
        data.append(wasmlib::uint64_to_bytes(self.nonce).as_mut());
        data.append(wasmlib::uint64_to_bytes(self.gas_budget).as_mut());
        let scheme = match self.signature_scheme.clone() {
            Some(val) => val.clone(),
            None => {
                panic!("signature_scheme is not given")
            }
        };
        let mut public_key = scheme.key_pair.public_key.to_bytes().to_vec();
        data.push(public_key.len() as u8);
        data.append(&mut public_key);
        data.append(self.allowance.to_bytes().as_mut());
        return data;
    }
    pub fn to_bytes(&self) -> Vec<u8> {
        let mut b = self.essence();
        b.append(
            &mut self
                .signature_scheme
                .clone()
                .unwrap()
                .signature
                .to_owned()
                .to_vec(),
        );
        return b;
    }
    pub fn with_allowance(&mut self, allowance: &ScAssets) {
        self.allowance = allowance.clone();
    }
}
