// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crypto::{
    hashes::{blake2b::Blake2b256, Digest},
    signatures::ed25519,
    signatures::ed25519::Signature,
};
use wasmlib::*;

pub struct KeyPair {
    private_key: ed25519::SecretKey,
    pub public_key: ed25519::PublicKey,
}

impl KeyPair {
    pub fn new(seed_bytes: &[u8]) -> KeyPair {
        let mut seed = [0; 32];
        if seed_bytes.len() != 0 {
            seed.copy_from_slice(seed_bytes);
        }
        let key = ed25519::SecretKey::from_bytes(seed);
        let pub_key = key.public_key();
        return KeyPair {
            private_key: key,
            public_key: pub_key,
        };
    }

    pub fn address(&self) -> ScAddress {
        let mut addr: Vec<u8> = Vec::with_capacity(SC_LENGTH_ED25519);
        addr.push(SC_ADDRESS_ED25519);
        let hash = Blake2b256::digest(self.public_key.to_bytes());
        addr.extend(&hash[..]);
        return address_from_bytes(&addr);
    }

    pub fn sign(&self, data: &[u8]) -> Vec<u8> {
        return self.private_key.sign(data).to_bytes().to_vec();
    }

    pub fn verify(&self, data: &[u8], sig: &[u8]) -> bool {
        let mut sig_data = [0; 64];
        sig_data.copy_from_slice(sig);
        self.public_key.verify(&Signature::from_bytes(sig_data), data)
    }

    pub fn sub_seed(seed: &[u8], n: u64) -> Vec<u8> {
        let index_bytes = uint64_to_bytes(n);
        let mut hash_of_index_bytes = Blake2b256::digest(index_bytes.to_owned());
        for i in 0..seed.len() {
            hash_of_index_bytes[i] ^= seed[i];
        }
        hash_of_index_bytes.to_vec()
    }

    pub fn from_sub_seed(seed: &[u8], n: u64) -> KeyPair {
        let sub_seed = KeyPair::sub_seed(seed, n);
        return KeyPair::new(&sub_seed);
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

impl PartialEq for KeyPair {
    fn eq(&self, other: &Self) -> bool {
        return self.private_key.as_slice() == other.private_key.as_slice()
            && self.public_key == other.public_key;
    }
}

#[cfg(test)]
mod tests {
    use wasmlib::{bytes_from_string, bytes_to_string};

    use crate::keypair::KeyPair;

    const MYSEED: &str = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3";

    #[test]
    fn keypair_clone() {
        let my_seed = bytes_from_string(&MYSEED);
        let pair1 = KeyPair::new(&my_seed);
        let pair2 = pair1.clone();

        println!("Publ1: {}", bytes_to_string(&pair1.public_key.to_bytes()));
        println!("Publ2: {}", bytes_to_string(&pair2.public_key.to_bytes()));
        println!("Priv1: {}", bytes_to_string(&pair1.private_key.to_bytes()));
        println!("Priv2: {}", bytes_to_string(&pair2.private_key.to_bytes()));
        assert_eq!(bytes_to_string(&pair1.public_key.to_bytes()), bytes_to_string(&pair2.public_key.to_bytes()));
        assert_eq!(bytes_to_string(&pair2.private_key.to_bytes()), bytes_to_string(&pair2.private_key.to_bytes()));
    }

    #[test]
    fn keypair_construct() {
        let my_seed = bytes_from_string(&MYSEED);
        let pair = KeyPair::new(&my_seed);
        println!("Publ: {}", bytes_to_string(&pair.public_key.to_bytes()));
        println!("Priv: {}", bytes_to_string(&pair.private_key.to_bytes()));
        assert_eq!(bytes_to_string(&pair.public_key.to_bytes()), "0x30adc0bd555d56ed51895528e47dcb403e36e0026fe49b6ae59e9adcea5f9a87");
        assert_eq!(bytes_to_string(&pair.private_key.to_bytes()), "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3");
    }

    #[test]
    fn keypair_from_sub_seed_0() {
        let my_seed = bytes_from_string(&MYSEED);
        let pair = KeyPair::from_sub_seed(&my_seed, 0);
        println!("Publ: {}", bytes_to_string(&pair.public_key.to_bytes()));
        println!("Priv: {}", bytes_to_string(&pair.private_key.to_bytes()));
        assert_eq!(bytes_to_string(&pair.public_key.to_bytes()), "0x40a757d26f6ef94dccee5b4f947faa78532286fe18117f2150a80acf2a95a8e2");
        assert_eq!(bytes_to_string(&pair.private_key.to_bytes()), "0x24642f47bd363fbd4e05f13ed6c60b04c8a4cf1d295f76fc16917532bc4cd0af");
    }

    #[test]
    fn keypair_from_sub_seed_1() {
        let my_seed = bytes_from_string(&MYSEED);
        let pair = KeyPair::from_sub_seed(&my_seed, 1);
        println!("Publ: {}", bytes_to_string(&pair.public_key.to_bytes()));
        println!("Priv: {}", bytes_to_string(&pair.private_key.to_bytes()));
        assert_eq!(bytes_to_string(&pair.public_key.to_bytes()), "0x120d2b26fc1b1d53bb916b8a277bcc2efa09e92c95be1a8fd5c6b3adbc795679");
        assert_eq!(bytes_to_string(&pair.private_key.to_bytes()), "0xb83d28550d9ee5651796eeb36027e737f0d79495b56d3d8931c716f2141017c8");
    }

    #[test]
    fn keypair_sign_and_verify() {
        let my_seed = bytes_from_string(&MYSEED);
        let pair = KeyPair::new(&my_seed);
        let signed_seed = pair.sign(&my_seed);
        println!("Seed: {}", bytes_to_string(&my_seed));
        println!("Sign: {}", bytes_to_string(&signed_seed));
        assert_eq!(bytes_to_string(&signed_seed), "0xa9571cc0c8612a63feaa325372a33c2f4ff6c414def18eb85ce4afe9b7cf01b84dba089278ca992e76fad8a50a76e3bf157216c445a404dc9e0424c250640906");
        assert!(pair.verify(&my_seed, &signed_seed));
    }

    #[test]
    fn keypair_sub_seed_0() {
        let my_seed = bytes_from_string(&MYSEED);
        let sub_seed = KeyPair::sub_seed(&my_seed, 0);
        println!("Seed: {}", bytes_to_string(&sub_seed));
        assert_eq!(bytes_to_string(&sub_seed), "0x24642f47bd363fbd4e05f13ed6c60b04c8a4cf1d295f76fc16917532bc4cd0af");
    }

    #[test]
    fn keypair_sub_seed_1() {
        let my_seed = bytes_from_string(&MYSEED);
        let sub_seed = KeyPair::sub_seed(&my_seed, 1);
        println!("Seed: {}", bytes_to_string(&sub_seed));
        assert_eq!(bytes_to_string(&sub_seed), "0xb83d28550d9ee5651796eeb36027e737f0d79495b56d3d8931c716f2141017c8");
    }
}