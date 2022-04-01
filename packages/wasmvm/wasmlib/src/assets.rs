// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::collections::BTreeMap;

use crate::*;

#[derive(Clone)]
pub struct ScAssets {
    iotas: u64,
    nfts: Vec<ScNftID>,
    tokens: BTreeMap<Vec<u8>, ScBigInt>,
}

impl ScAssets {
    pub fn new(buf: &[u8]) -> ScAssets {
        let mut assets = ScAssets {
            iotas: 0,
            nfts: Vec::new(),
            tokens: BTreeMap::new(),
        };
        if buf.len() == 0 {
            return assets;
        }

        let mut dec = WasmDecoder::new(buf);
        assets.iotas = uint64_decode(&mut dec);

        let size = uint32_decode(&mut dec);
        for _i in 0..size {
            let token_id = token_id_decode(&mut dec);
            let amount = big_int_decode(&mut dec);
            assets.tokens.insert(token_id.to_bytes(), amount);
        }

        let size = uint32_decode(&mut dec);
        for _i in 0..size {
            let nft_id = nft_id_decode(&mut dec);
            assets.nfts.push(nft_id);
        }
        assets
    }

    pub fn balances(&self) -> ScBalances {
        ScBalances { assets: self.clone() }
    }

    pub fn is_empty(&self) -> bool {
        if self.iotas != 0 {
            return false;
        }
        for (_key, val) in self.tokens.iter() {
            if !val.is_zero() {
                return false;
            }
        }
        self.nfts.len() == 0
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
        uint64_encode(&mut enc, self.iotas);

        uint32_encode(&mut enc, self.tokens.len() as u32);
        for (token_id, amount) in self.tokens.iter() {
            enc.fixed_bytes(token_id, SC_TOKEN_ID_LENGTH);
            big_int_encode(&mut enc, amount);
        }

        uint32_encode(&mut enc, self.nfts.len() as u32);
        for nft_id in self.nfts.iter() {
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
        ScBalances { assets: ScAssets::new(&[]) }
    }

    pub fn balance(&self, token: &ScTokenID) -> ScBigInt {
        let key = token.to_bytes();
        if !self.assets.tokens.contains_key(&key) {
            return ScBigInt::new();
        }
        self.assets.tokens.get(&key).unwrap().clone()
    }

    pub fn iotas(&self) -> u64 {
        self.assets.iotas
    }

    pub fn is_empty(&self) -> bool {
        self.assets.is_empty()
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
        ScTransfer { balances: ScBalances::new() }
    }

    pub fn from_balances(balances: ScBalances) -> ScTransfer {
        let mut transfer = ScTransfer::new();
        let token_ids = balances.token_ids();
        for i in 0..token_ids.len() {
            let token_id = token_ids.get(i).unwrap();
            transfer.set(token_id, &balances.balance(token_id))
        }
        transfer
    }

    pub fn iotas(amount: u64) -> ScTransfer {
        let mut transfer = ScTransfer::new();
        transfer.balances.assets.iotas = amount;
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
        self.balances.assets.nfts.push(nft_id.clone());
    }

    pub fn set(&mut self, token: &ScTokenID, amount: &ScBigInt) {
        self.balances.assets.tokens.insert(token.to_bytes(), amount.clone());
    }
}

impl std::ops::Deref for ScTransfer {
    type Target = ScBalances;
    fn deref(&self) -> &Self::Target {
        &self.balances
    }
}