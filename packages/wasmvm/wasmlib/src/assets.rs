// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::collections::{BTreeMap, HashSet};

use crate::*;

#[derive(Clone)]
pub struct ScAssets {
    base_tokens: u64,
    native_tokens: BTreeMap<Vec<u8>, ScBigInt>,
    nft_ids: HashSet<ScNftID>,
}

impl ScAssets {
    pub fn new(buf: &[u8]) -> ScAssets {
        let mut assets = ScAssets::new_base_tokens(0);
        if buf.len() == 0 {
            return assets;
        }

        let mut dec = WasmDecoder::new(buf);
        let empty = bool_decode(&mut dec);
        if empty {
            return assets;
        }

        assets.base_tokens = uint64_decode(&mut dec);

        let size = uint16_decode(&mut dec);
        for _i in 0..size {
            let token_id = token_id_decode(&mut dec);
            let amount = big_int_decode(&mut dec);
            assets.native_tokens.insert(token_id.to_bytes(), amount);
        }

        let size = uint16_decode(&mut dec);
        for _i in 0..size {
            let nft_id = nft_id_decode(&mut dec);
            assets.nft_ids.insert(nft_id);
        }
        assets
    }

    pub fn new_base_tokens(base_token_num: u64) -> ScAssets {
        return ScAssets {
            base_tokens: base_token_num,
            native_tokens: BTreeMap::new(),
            nft_ids: HashSet::new(),
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
        self.nft_ids.len() == 0
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        let empty = self.is_empty();
        bool_encode(&mut enc, empty);
        if empty {
            return enc.buf();
        }

        uint64_encode(&mut enc, self.base_tokens);

        uint16_encode(&mut enc, self.native_tokens.len() as u16);
        for (token_id, amount) in self.native_tokens.iter() {
            enc.fixed_bytes(token_id, SC_TOKEN_ID_LENGTH);
            big_int_encode(&mut enc, amount);
        }

        uint16_encode(&mut enc, self.nft_ids.len() as u16);
        for nft_id in self.nft_ids.iter() {
            nft_id_encode(&mut enc, &nft_id);
        }
        enc.buf()
    }

    pub fn token_ids(&self) -> Vec<ScTokenID> {
        let mut tokens: Vec<ScTokenID> = Vec::new();
        for (key, _val) in self.native_tokens.iter() {
            tokens.push(token_id_from_bytes(key));
        }
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
        for nft_id in self.assets.nft_ids.iter() {
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
        self.balances.assets.nft_ids.insert(nft_id.clone());
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
