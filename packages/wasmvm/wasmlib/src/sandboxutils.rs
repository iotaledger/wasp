// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::*;
use crate::host::*;

pub struct ScSandboxUtils {}

impl ScSandboxUtils {
    // decodes the specified bech32-encoded string value to its original address
    pub fn bech32_decode(&self, bech32: &str) -> ScAddress {
        address_from_bytes(&sandbox(FN_UTILS_BECH32_DECODE, &string_to_bytes(bech32)))
    }

    // encodes the specified address to a bech32-encoded string
    pub fn bech32_encode(&self, addr: &ScAddress) -> String {
        string_from_bytes(&sandbox(FN_UTILS_BECH32_ENCODE, &address_to_bytes(addr)))
    }

    pub fn bls_address_from_pub_key(&self, pub_key: &[u8]) -> ScAddress {
        address_from_bytes(&sandbox(FN_UTILS_BLS_ADDRESS, pub_key))
    }

    pub fn bls_aggregate_signatures(&self, pub_keys: &[&[u8]], sigs: &[&[u8]]) -> Vec<Vec<u8>> {
        let mut enc = WasmEncoder::new();
        uint32_encode(&mut enc, pub_keys.len() as u32);
        for i in 0..pub_keys.len() {
            enc.bytes(pub_keys[i]);
        }
        uint32_encode(&mut enc, sigs.len() as u32);
        for i in 0..sigs.len() {
            enc.bytes(sigs[i]);
        }
        let res = sandbox(FN_UTILS_BLS_AGGREGATE, &enc.buf());
        let mut dec = WasmDecoder::new(&res);
        return [dec.bytes(), dec.bytes()].to_vec();
    }

    pub fn bls_valid_signature(&self, data: &[u8], pub_key: &[u8], signature: &[u8]) -> bool {
        let mut enc = WasmEncoder::new();
        enc.bytes(data);
        enc.bytes(pub_key);
        enc.bytes(signature);
        bool_from_bytes(&sandbox(FN_UTILS_BLS_VALID, &enc.buf()))
    }

    pub fn ed25519_address_from_pub_key(&self, pub_key: &[u8]) -> ScAddress {
        address_from_bytes(&sandbox(FN_UTILS_ED25519_ADDRESS, pub_key))
    }

    pub fn ed25519_valid_signature(&self, data: &[u8], pub_key: &[u8], signature: &[u8]) -> bool {
        let mut enc = WasmEncoder::new();
        enc.bytes(data);
        enc.bytes(pub_key);
        enc.bytes(signature);
        bool_from_bytes(&sandbox(FN_UTILS_ED25519_VALID, &enc.buf()))
    }

    // hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
    pub fn hash_blake2b(&self, value: &[u8]) -> ScHash {
        hash_from_bytes(&sandbox(FN_UTILS_HASH_BLAKE2B, value))
    }

    // hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
    pub fn hash_name(&self, name: &str) -> ScHname {
        hname_from_bytes(&sandbox(FN_UTILS_HASH_NAME, &string_to_bytes(name)))
    }

    // hashes the specified value bytes using sha3 hashing and returns the resulting 32-byte hash
    pub fn hash_sha3(&self, value: &[u8]) -> ScHash {
        hash_from_bytes(&sandbox(FN_UTILS_HASH_SHA3, value))
    }
}