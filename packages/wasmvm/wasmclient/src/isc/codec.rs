pub use crate::types::*;
use bech32::*;
use iota_client;
pub const BECH32_PREFIX: &'static str = "smr";

pub fn bech32_decode(input: &str) -> Result<Address, String> {
    let (_hrp, data, _v) = bech32::decode(&input).unwrap();
    return Ok(data.iter().map(|&e| e.to_u8()).collect());
}

pub fn bech32_encode(addr: &Address) -> String {
    return bech32::encode(BECH32_PREFIX, addr.to_base32(), Variant::Bech32).unwrap();
}

use crypto::hashes::{blake2b::Blake2b256, Digest};

pub fn hname_bytes(name: &str) -> Vec<u8> {
    let hash = Blake2b256::digest(name.as_bytes());
    let mut slice = &hash[0..4];
    let hname = wasmlib::uint32_from_bytes(slice);
    if hname == 0 || hname == 0xffff {
        slice = &hash[4..8];
    }
    return slice.to_vec();
}
