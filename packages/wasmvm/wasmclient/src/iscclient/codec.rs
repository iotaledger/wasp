// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use bech32::{FromBase32, ToBase32, Variant};
use crypto::hashes::{blake2b::Blake2b256, Digest};
use serde::{Deserialize, Serialize};
use tiny_keccak::{Hasher, Keccak};
use wasmlib::*;

pub type Result<T> = std::result::Result<T, String>;

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JsonItem {
    key: String,
    value: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JsonDict {
    items: Vec<JsonItem>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub(crate) struct APICallViewRequest {
    pub(crate) arguments: JsonDict,
    #[serde(rename = "contractHName")]
    pub(crate) contract_hname: String,
    #[serde(rename = "functionHName")]
    pub(crate) function_hname: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub(crate) struct APIOffLedgerRequest {
    #[serde(rename = "chainId")]
    pub(crate) chain_id: String,
    pub(crate) request: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JsonResponse {
    #[serde(rename = "Items")]
    pub(crate) items: Vec<JsonItem>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JsonError {
    #[serde(rename = "Error")]
    pub(crate) error: String,
    #[serde(rename = "Message")]
    pub(crate) message: String,
}

pub(crate) struct Codec {}

impl Codec {
    fn bech32_decode(input: &str) -> Result<(String, ScAddress)> {
        let (hrp, data, _v) = match bech32::decode(&input) {
            Ok(v) => v,
            Err(_) => return Err(String::from(format!("invalid bech32 string: {}", input))),
        };
        let buf = match Vec::<u8>::from_base32(&data) {
            Ok(b) => b,
            Err(e) => return Err(e.to_string()),
        };
        return Ok((hrp, address_from_bytes(&buf)));
    }

    fn bech32_encode(hrp: &str, addr: &ScAddress) -> Result<String> {
        match bech32::encode(hrp, addr.to_bytes().to_base32(), Variant::Bech32) {
            Ok(v) => Ok(v),
            Err(e) => Err(e.to_string()),
        }
    }

    fn hash_keccak(buf: &[u8]) -> Vec<u8> {
        let mut keccak = Keccak::v256();
        keccak.update(buf);
        let mut output: [u8; 32] = [0; 32];
        keccak.finalize(&mut output);
        output.to_vec()
    }

    fn hash_name(name: &str) -> Vec<u8> {
        let hash = Blake2b256::digest(name.as_bytes());
        for i in (0..hash.len()).step_by(SC_HNAME_LENGTH) {
            let slice = &hash[i..i + SC_HNAME_LENGTH];
            let hname = uint32_from_bytes(slice);
            if hname != 0 {
                return slice.to_vec();
            }
        }
        // astronomically unlikely to end up here
        return uint32_to_bytes(1);
    }

    pub fn hrp_for_client() -> String {
        unsafe {
            HRP_FOR_CLIENT.clone()
        }
    }

    pub fn json_decode(dict: JsonResponse) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        let items_num = dict.items.len();
        enc.vlu_encode(items_num as u64);
        for i in 0..items_num {
            let item = dict.items[i].clone();
            let key = hex_decode(&item.key);
            let val = hex_decode(&item.value);
            enc.vlu_encode(key.len() as u64);
            enc.fixed_bytes(&key, key.len());
            enc.vlu_encode(val.len() as u64);
            enc.fixed_bytes(&val, val.len());
        }
        return enc.buf();
    }

    pub fn json_encode(buf: &[u8]) -> JsonDict {
        let mut dec = WasmDecoder::new(buf);
        let size = dec.vlu_decode(32);
        let mut dict = JsonDict {
            items: Vec::with_capacity(size as usize),
        };
        for _ in 0..size {
            let key_len = dec.vlu_decode(32);
            let key = dec.fixed_bytes(key_len as usize);
            let val_len = dec.vlu_decode(32);
            let val = dec.fixed_bytes(val_len as usize);
            let item = JsonItem {
                key: hex_encode(&key),
                value: hex_encode(&val),
            };
            dict.items.push(item);
        }
        return dict;
    }
}

static mut HRP_FOR_CLIENT: String = String::new();

fn client_bech32_decode(bech32: &str) -> ScAddress {
    match Codec::bech32_decode(&bech32) {
        Ok((hrp, addr)) => unsafe {
            if hrp != HRP_FOR_CLIENT {
                panic(&("invalid protocol prefix: ".to_owned() + &hrp));
                return address_from_bytes(&[]);
            }
            addr
        },
        Err(e) => {
            panic(&e.to_string());
            address_from_bytes(&[])
        }
    }
}

fn client_bech32_encode(addr: &ScAddress) -> String {
    unsafe {
        match Codec::bech32_encode(&HRP_FOR_CLIENT, &addr) {
            Ok(v) => v,
            Err(e) => {
                panic(&e.to_string());
                String::new()
            }
        }
    }
}

fn client_hash_keccak(buf: &[u8]) -> ScHash {
    hash_from_bytes(&Codec::hash_keccak(buf))
}

fn client_hash_name(name: &str) -> ScHname {
    hname_from_bytes(&Codec::hash_name(name))
}

pub(crate) fn set_sandbox_wrappers(chain_id: &str) -> Result<()> {
    unsafe {
        // local client implementations for some sandbox  functions
        BECH32_DECODE = client_bech32_decode;
        BECH32_ENCODE = client_bech32_encode;
        HASH_KECCAK = client_hash_keccak;
        HASH_NAME = client_hash_name;
    }

    // set the network prefix for the current network
    match Codec::bech32_decode(chain_id) {
        Ok((hrp, _)) => unsafe {
            if HRP_FOR_CLIENT.len() != 0 && HRP_FOR_CLIENT != hrp {
                panic!("WasmClient can only connect to one Tangle network per app");
            }
            HRP_FOR_CLIENT = hrp;
        },
        Err(e) => return Err(e),
    };
    Ok(())
}