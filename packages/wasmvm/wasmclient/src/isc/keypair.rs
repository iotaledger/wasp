// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crypto::hashes::{blake2b::Blake2b256, Digest};
use crypto::signatures::ed25519;
use std::fmt::Debug;
use wasmlib::*;

pub struct KeyPair {
    private_key: ed25519::SecretKey,
    pub public_key: ed25519::PublicKey,
}

impl KeyPair {
    pub fn address(&self) -> ScAddress {
        let mut addr: Vec<u8> = Vec::with_capacity(wasmlib::SC_LENGTH_ED25519);
        addr[0] = wasmlib::SC_ADDRESS_ED25519;
        let hash = Blake2b256::digest(self.public_key.to_bytes());
        addr.copy_from_slice(&hash[..]);
        return wasmlib::address_from_bytes(&addr);
    }
    pub fn sign(&self, data: &[u8]) -> Vec<u8> {
        return self.private_key.sign(data).to_bytes().to_vec();
    }
}

impl Clone for KeyPair {
    fn clone(&self) -> Self {
        return KeyPair {
            private_key: ed25519::SecretKey::from_bytes(self.private_key.to_bytes()),
            public_key: self.public_key.clone(),
        };
    }
}

impl Debug for KeyPair {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> Result<(), std::fmt::Error> {
        f.debug_tuple("KeyPair").field(&self.public_key).finish()
    }
}

impl PartialEq for KeyPair {
    fn eq(&self, other: &Self) -> bool {
        todo!()
    }
}
