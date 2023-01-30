// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use bech32::*;
use crypto::hashes::{blake2b::Blake2b256, Digest};
use serde::{Deserialize, Serialize};
use wasmlib::*;
pub use wasmtypes::*;

use crate::errors;

const BECH32_PREFIX: &'static str = "smr";

pub fn bech32_decode(input: &str) -> errors::Result<(String, ScAddress)> {
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

pub fn bech32_encode(hrp: &str, addr: &ScAddress) -> errors::Result<String> {
    match bech32::encode(hrp, addr.to_bytes().to_base32(), Variant::Bech32) {
        Ok(v) => Ok(v),
        Err(e) => Err(e.to_string()),
    }
}

pub fn hname_bytes(name: &str) -> Vec<u8> {
    let hash = Blake2b256::digest(name.as_bytes());
    let mut slice = &hash[0..4];
    let hname = wasmlib::uint32_from_bytes(slice);
    if hname == 0 || hname == 0xffff {
        slice = &hash[4..8];
    }
    return slice.to_vec();
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JsonItem {
    key: String,
    value: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JsonRequest {
    items: Vec<JsonItem>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JsonResponse {
    items: Vec<JsonItem>,
    message: String,
    status_code: u16,
}

pub fn json_decode(dict: JsonResponse) -> Vec<u8> {
    let mut enc = WasmEncoder::new();
    let items_num = dict.items.len();
    enc.fixed_bytes(&uint32_to_bytes(items_num as u32), SC_UINT32_LENGTH);
    for i in 0..items_num {
        let item = dict.items[i].clone();
        let key = hex_decode(&item.key);
        let val = hex_decode(&item.value);
        enc.fixed_bytes(&uint16_to_bytes(key.len() as u16), SC_UINT16_LENGTH);
        enc.fixed_bytes(&key, key.len());
        enc.fixed_bytes(&uint32_to_bytes(val.len() as u32), SC_UINT32_LENGTH);
        enc.fixed_bytes(&val, val.len());
    }
    return enc.buf();
}

pub fn json_encode(buf: &[u8]) -> JsonRequest {
    let mut dec = WasmDecoder::new(buf);
    let items_num = uint32_from_bytes(&dec.fixed_bytes(SC_UINT32_LENGTH));
    let mut dict = JsonRequest {
        items: Vec::with_capacity(items_num as usize),
    };
    for _ in 0..items_num {
        let key_buf = dec.fixed_bytes(SC_UINT16_LENGTH);
        let key_len = uint16_from_bytes(&key_buf);
        let key = dec.fixed_bytes(key_len as usize);
        let val_buf = dec.fixed_bytes(SC_UINT32_LENGTH);
        let val_len = uint32_from_bytes(&val_buf);
        let val = dec.fixed_bytes(val_len as usize);
        let item = JsonItem {
            key: hex_encode(&key),
            value: hex_encode(&val),
        };
        dict.items.push(item);
    }
    return dict;
}
