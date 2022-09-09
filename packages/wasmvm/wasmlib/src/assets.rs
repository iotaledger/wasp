// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::collections::{BTreeMap, HashSet};

use crate::*;

#[derive(Clone)]
pub struct ScAssets {
    base_tokens: u64,
    nft_ids: HashSet<ScNftID>,
    tokens: BTreeMap<Vec<u8>, ScBigInt>,
}

impl ScAssets {
    pub fn new(buf: &[u8]) -> ScAssets {
        let mut assets = ScAssets {
            base_tokens: 0,
            nft_ids: HashSet::new(),
            tokens: BTreeMap::new(),
        };
        if buf.len() == 0 {
            return assets;
        }

        let mut dec = WasmDecoder::new(buf);
        assets.base_tokens = uint64_decode(&mut dec);

        let size = uint32_decode(&mut dec);
        for _i in 0..size {
            let token_id = token_id_decode(&mut dec);
            let amount = big_int_decode(&mut dec);
            assets.tokens.insert(token_id.to_bytes(), amount);
        }

        let size = uint32_decode(&mut dec);
        for _i in 0..size {
            let nft_id = nft_id_decode(&mut dec);
            assets.nft_ids.insert(nft_id);
        }
        assets
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
        for (_key, val) in self.tokens.iter() {
            if !val.is_zero() {
                return false;
            }
        }
        self.nft_ids.len() == 0
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        uint64_encode(&mut enc, self.base_tokens);

        uint32_encode(&mut enc, self.tokens.len() as u32);
        for (token_id, amount) in self.tokens.iter() {
            enc.fixed_bytes(token_id, SC_TOKEN_ID_LENGTH);
            big_int_encode(&mut enc, amount);
        }

        uint32_encode(&mut enc, self.nft_ids.len() as u32);
        for nft_id in self.nft_ids.iter() {
            nft_id_encode(&mut enc, &nft_id);
        }
        return enc.buf();
    }

    pub fn token_ids(&self) -> Vec<ScTokenID> {
        let mut tokens: Vec<ScTokenID> = Vec::new();
        for (key, _val) in self.tokens.iter() {
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
        if !self.assets.tokens.contains_key(&key) {
            return ScBigInt::new();
        }
        self.assets.tokens.get(&key).unwrap().clone()
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
            .tokens
            .insert(token.to_bytes(), amount.clone());
    }
}

impl std::ops::Deref for ScTransfer {
    type Target = ScBalances;
    fn deref(&self) -> &Self::Target {
        &self.balances
    }
}
