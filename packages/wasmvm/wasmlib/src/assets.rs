// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::collections::BTreeMap;

use crate::*;

fn read_bytes(assets: &mut BTreeMap<Vec<u8>, u64>, buf: &[u8]) {
    if buf.len() != 0 {
        let mut dec = WasmDecoder::new(buf);
        let size = uint32_from_bytes(&dec.fixed_bytes(SC_UINT32_LENGTH));
        for _i in 0..size {
            let color = color_decode(&mut dec);
            let amount_buf = dec.fixed_bytes(SC_UINT64_LENGTH);
            let amount = uint64_from_bytes(&amount_buf);
            assets.insert(color.to_bytes(), amount);
        }
    }
}

#[derive(Clone)]
pub struct ScAssets {
    assets: BTreeMap<Vec<u8>, u64>,
}

impl ScAssets {
    pub fn new(buf: &[u8]) -> ScAssets {
        let mut assets = ScAssets { assets: BTreeMap::new() };
        read_bytes(&mut assets.assets, buf);
        assets
    }

    pub fn balances(&self) -> ScBalances {
        ScBalances { assets: self.assets.clone() }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let dict = &self.assets;
        if dict.len() == 0 {
            return vec![0; SC_UINT32_LENGTH];
        }

        let mut enc = WasmEncoder::new();
        enc.fixed_bytes(&uint32_to_bytes(dict.len() as u32), SC_UINT32_LENGTH);
        for (key, val) in dict.iter() {
            enc.fixed_bytes(key, SC_COLOR_LENGTH);
            enc.fixed_bytes(&uint64_to_bytes(*val), SC_UINT64_LENGTH);
        }
        return enc.buf();
    }
}

#[derive(Clone)]
pub struct ScBalances {
    assets: BTreeMap<Vec<u8>, u64>,
}

impl ScBalances {
    pub fn balance(&self, color: &ScColor) -> u64 {
        let key = color.to_bytes();
        if !self.assets.contains_key(&key) {
            return 0;
        }
        *self.assets.get(&key).unwrap()
    }

    pub fn colors(&self) -> Vec<ScColor> {
        let mut colors: Vec<ScColor> = Vec::new();
        for color in self.assets.keys() {
            colors.push(color_from_bytes(color));
        }
        colors
    }
}

#[derive(Clone)]
pub struct ScTransfers {
    assets: BTreeMap<Vec<u8>, u64>,
}

impl ScTransfers {
    pub fn new(buf: &[u8]) -> ScTransfers {
        let mut assets = ScTransfers { assets: BTreeMap::new() };
        read_bytes(&mut assets.assets, buf);
        assets
    }

    pub fn from_balances(balances: ScBalances) -> ScTransfers {
        let mut transfers = ScTransfers { assets: BTreeMap::new() };
        let colors = balances.colors();
        for i in 0..colors.len() {
            let color = colors.get(i).unwrap();
            transfers.set(color, balances.balance(color))
        }
        transfers
    }

    pub fn iotas(amount: u64) -> ScTransfers {
        ScTransfers::transfer(&ScColor::IOTA, amount)
    }

    pub fn transfer(color: &ScColor, amount: u64) -> ScTransfers {
        let mut transfers = ScTransfers { assets: BTreeMap::new() };
        transfers.set(color, amount);
        transfers
    }

    pub fn as_assets(&self) -> ScAssets {
        ScAssets { assets: self.assets.clone() }
    }

    pub fn balances(&self) -> ScBalances {
        ScBalances { assets: self.assets.clone() }
    }

    pub fn set(&mut self, color: &ScColor, amount: u64) {
        self.assets.insert(color.to_bytes(), amount);
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        self.as_assets().to_bytes()
    }
}
