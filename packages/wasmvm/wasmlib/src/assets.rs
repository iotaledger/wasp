// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::collections::{BTreeMap, HashSet};

use crate::*;

const HAS_BASE_TOKENS: u8 = 0x80;
const HAS_NATIVE_TOKENS: u8 = 0x40;
const HAS_NFTS: u8 = 0x20;

#[derive(Clone)]
pub struct ScAssets {
    base_tokens: u64,
    native_tokens: BTreeMap<Vec<u8>, ScBigInt>,
    nfts: HashSet<ScNftID>,
}

impl ScAssets {
    pub fn new(buf: &[u8]) -> ScAssets {
        let mut assets = ScAssets::new_base_tokens(0);
        if buf.len() == 0 {
            return assets;
        }

        let mut dec = WasmDecoder::new(buf);
        let flags = uint8_decode(&mut dec);
        if flags == 0x00 {
            return assets;
        }
        if (flags & HAS_BASE_TOKENS) != 0 {
            assets.base_tokens = dec.vlu_decode(64);
        }
        if (flags & HAS_NATIVE_TOKENS) != 0 {
            let size = dec.vlu_decode(16);
            for _i in 0..size {
                let token_id = token_id_decode(&mut dec);
                let amount = big_int_decode(&mut dec);
                assets.native_tokens.insert(token_id.to_bytes(), amount);
            }
        }
        if (flags & HAS_NFTS) != 0 {
            let size = dec.vlu_decode(16);
            for _i in 0..size {
                let nft_id = nft_id_decode(&mut dec);
                assets.nfts.insert(nft_id);
            }
        }
        assets
    }

    pub fn new_base_tokens(base_token_num: u64) -> ScAssets {
        return ScAssets {
            base_tokens: base_token_num,
            native_tokens: BTreeMap::new(),
            nfts: HashSet::new(),
        };
    }

    pub fn balances(&self) -> ScBalances {
        ScBalances {
            assets: self.clone(),
        }
    }

    pub fn is_empty(&self) -> bool {
        if self.base_tokens != 0 {
            return false;
        }
        for (_key, val) in self.native_tokens.iter() {
            if !val.is_zero() {
                return false;
            }
        }
        self.nfts.len() == 0
    }

    pub fn nft_ids(&self) -> Vec<ScNftID> {
        let mut nfts: Vec<ScNftID> = Vec::new();
        for nft in &self.nfts {
            nfts.push(nft.clone());
        }
        nfts.sort();
        nfts
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        if self.is_empty() {
            return vec![0; 1];
        }

        let mut flags = 0 as u8;
        if self.base_tokens != 0 {
            flags |= HAS_BASE_TOKENS;
        }
        if self.native_tokens.len() != 0 {
            flags |= HAS_NATIVE_TOKENS;
        }
        if self.nfts.len() != 0 {
            flags |= HAS_NFTS;
        }
        uint8_encode(&mut enc, flags);
        
        if (flags & HAS_BASE_TOKENS) != 0 {
            enc.vlu_encode(self.base_tokens);
        }
        if (flags & HAS_NATIVE_TOKENS) != 0 {
            enc.vlu_encode(self.native_tokens.len() as u64);
            for token_id in self.token_ids() {
                token_id_encode(&mut enc, &token_id);
                let amount = self.native_tokens.get(&token_id.to_bytes());
                big_int_encode(&mut enc, amount.unwrap());
            }
        }
        if (flags & HAS_NFTS) != 0 {
            enc.vlu_encode(self.nfts.len() as u64);
            for nft_id in self.nft_ids() {
                nft_id_encode(&mut enc, &nft_id);
            }
        }
        enc.buf()
    }

    pub fn token_ids(&self) -> Vec<ScTokenID> {
        let mut tokens: Vec<ScTokenID> = Vec::new();
        for (key, _val) in &self.native_tokens {
            tokens.push(token_id_from_bytes(key));
        }
        tokens.sort();
        tokens
    }
}

#[derive(Clone)]
pub struct ScBalances {
    assets: ScAssets,
}

impl ScBalances {
    fn new() -> ScBalances {
        ScBalances {
            assets: ScAssets::new(&[]),
        }
    }

    pub fn balance(&self, token: &ScTokenID) -> ScBigInt {
        let key = token.to_bytes();
        if !self.assets.native_tokens.contains_key(&key) {
            return ScBigInt::new();
        }
        self.assets.native_tokens.get(&key).unwrap().clone()
    }

    pub fn base_tokens(&self) -> u64 {
        self.assets.base_tokens
    }

    pub fn is_empty(&self) -> bool {
        self.assets.is_empty()
    }

    pub fn nft_ids(&self) -> Vec<ScNftID> {
        let mut nft_ids = Vec::new();
        for nft_id in self.assets.nfts.iter() {
            nft_ids.push(*nft_id);
        }
        return nft_ids;
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        self.assets.to_bytes()
    }

    pub fn token_ids(&self) -> Vec<ScTokenID> {
        self.assets.token_ids()
    }
}

#[derive(Clone)]
pub struct ScTransfer {
    balances: ScBalances,
}

impl ScTransfer {
    pub fn new() -> ScTransfer {
        ScTransfer {
            balances: ScBalances::new(),
        }
    }

    pub fn from_balances(balances: &ScBalances) -> ScTransfer {
        let mut transfer = ScTransfer::base_tokens(balances.base_tokens());
        for token_id in balances.token_ids().iter() {
            transfer.set(token_id, &balances.balance(token_id))
        }
        for nft_id in balances.nft_ids().iter() {
            transfer.add_nft(nft_id);
        }
        transfer
    }

    pub fn base_tokens(amount: u64) -> ScTransfer {
        let mut transfer = ScTransfer::new();
        transfer.balances.assets.base_tokens = amount;
        transfer
    }

    pub fn nft(nft_id: &ScNftID) -> ScTransfer {
        let mut transfer = ScTransfer::new();
        transfer.add_nft(nft_id);
        transfer
    }

    pub fn tokens(token_id: &ScTokenID, amount: &ScBigInt) -> ScTransfer {
        let mut transfer = ScTransfer::new();
        transfer.set(token_id, amount);
        transfer
    }

    pub fn add_nft(&mut self, nft_id: &ScNftID) {
        self.balances.assets.nfts.insert(nft_id.clone());
    }

    pub fn set(&mut self, token: &ScTokenID, amount: &ScBigInt) {
        self.balances
            .assets
            .native_tokens
            .insert(token.to_bytes(), amount.clone());
    }
}

impl std::ops::Deref for ScTransfer {
    type Target = ScBalances;
    fn deref(&self) -> &Self::Target {
        &self.balances
    }
}
