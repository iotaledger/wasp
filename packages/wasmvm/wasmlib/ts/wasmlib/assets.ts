// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {ScDict} from "./dict";

export class ScAssets {
    baseTokens: u64 = 0;
    nftIDs: wasmtypes.ScNftID[] = [];
    tokens: Map<string, wasmtypes.ScBigInt> = new Map();

    public constructor(buf: u8[]) {
        if (buf.length == 0) {
            return this;
        }
        const dec = new wasmtypes.WasmDecoder(buf);
        this.baseTokens = wasmtypes.uint64Decode(dec);

        let size = wasmtypes.uint32Decode(dec);
        for (let i: u32 = 0; i < size; i++) {
            const tokenID = wasmtypes.tokenIDDecode(dec);
            const amount = wasmtypes.bigIntDecode(dec);
            this.tokens.set(ScDict.toKey(tokenID.id), amount);
        }

        size = wasmtypes.uint32Decode(dec);
        for (let i: u32 = 0; i < size; i++) {
            const nftID = wasmtypes.nftIDDecode(dec);
            this.nftIDs.push(nftID)
        }
    }

    public balances(): ScBalances {
        return new ScBalances(this);
    }

    public isEmpty(): bool {
        if (this.baseTokens != 0) {
            return false;
        }
        const values = this.tokens.values();
        for (let i = 0; i < values.length; i++) {
            if (!values[i].isZero()) {
                return false;
            }
        }
        return this.nftIDs.length == 0;
    }
    
    public toBytes(): u8[] {
        const enc = new wasmtypes.WasmEncoder();
        wasmtypes.uint64Encode(enc, this.baseTokens);

        let tokenIDs = this.tokenIDs();
        wasmtypes.uint32Encode(enc, tokenIDs.length as u32);
        for (let i = 0; i < tokenIDs.length; i++) {
            const tokenID = tokenIDs[i]
            wasmtypes.tokenIDEncode(enc, tokenID);
            const mapKey = ScDict.toKey(tokenID.id);
            const amount = this.tokens.get(mapKey);
            wasmtypes.bigIntEncode(enc, amount);
        }

        wasmtypes.uint32Encode(enc, this.nftIDs.length as u32);
        for (let i = 0; i < this.nftIDs.length; i++) {
            const nftID = this.nftIDs[i]
            wasmtypes.nftIDEncode(enc, nftID);
        }
        return enc.buf()
    }
    
    public tokenIDs(): wasmtypes.ScTokenID[] {
        let tokenIDs: wasmtypes.ScTokenID[] = [];
        const keys = this.tokens.keys().sort();
        for (let i = 0; i < keys.length; i++) {
            const keyBytes = ScDict.fromKey(keys[i]);
            const tokenID = wasmtypes.tokenIDFromBytes(keyBytes);
            tokenIDs.push(tokenID);
        }
        return tokenIDs;
    }
}

export class ScBalances {
    assets: ScAssets;

    constructor(assets: ScAssets) {
        this.assets = assets;
    }

    public balance(tokenID: wasmtypes.ScTokenID): wasmtypes.ScBigInt {
        const mapKey = ScDict.toKey(tokenID.id);
        if (!this.assets.tokens.has(mapKey)) {
            return new wasmtypes.ScBigInt();
        }
        return this.assets.tokens.get(mapKey);
    }

    public baseTokens(): u64 {
        return this.assets.baseTokens;
    }

    public isEmpty(): bool {
        return this.assets.isEmpty();
    }

    public nftIDs(): wasmtypes.ScNftID[] {
        return this.assets.nftIDs;
    }

    public toBytes(): u8[] {
        return this.assets.toBytes();
    }

    public tokenIDs(): wasmtypes.ScTokenID[] {
        return this.assets.tokenIDs();
    }
}

export class ScTransfer extends ScBalances{
    public constructor() {
        super(new ScAssets([]));
    }

    public static fromBalances(balances: ScBalances): ScTransfer {
        const transfer = ScTransfer.baseTokens(balances.baseTokens());
        const tokenIDs = balances.tokenIDs();
        for (let i = 0; i < tokenIDs.length; i++) {
            const tokenID = tokenIDs[i];
            transfer.set(tokenID, balances.balance(tokenID));
        }
        const nftIDs = balances.nftIDs();
        for (let i = 0; i < nftIDs.length; i++) {
            const nftID = nftIDs[i];
            transfer.addNFT(nftID);
        }
        return transfer;
    }

    public static baseTokens(amount: u64): ScTransfer {
        const transfer = new ScTransfer();
        transfer.assets.baseTokens = amount;
        return transfer;
    }

    public static nft(nftID: wasmtypes.ScNftID): ScTransfer {
        const transfer = new ScTransfer();
        transfer.addNFT(nftID);
        return transfer;
    }

    public static tokens(tokenID: wasmtypes.ScTokenID, amount: wasmtypes.ScBigInt): ScTransfer {
        const transfer = new ScTransfer();
        transfer.set(tokenID, amount);
        return transfer;
    }

    public addNFT(nftID: wasmtypes.ScNftID): void {
        this.assets.nftIDs.push(nftID);
    }

    public set(tokenID: wasmtypes.ScTokenID, amount: wasmtypes.ScBigInt): void {
        const mapKey = ScDict.toKey(tokenID.id);
        this.assets.tokens.set(mapKey, amount);
    }
}
