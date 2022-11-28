// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {ScDict} from "./dict";
import {ScTokenID, tokenIDDecode, tokenIDEncode, tokenIDFromBytes} from "./wasmtypes/sctokenid";
import {uint64Decode, uint64Encode} from "./wasmtypes/scuint64";
import {bigIntDecode, bigIntEncode, ScBigInt} from "./wasmtypes/scbigint";
import {nftIDDecode, nftIDEncode, ScNftID} from "./wasmtypes/scnftid";
import {WasmDecoder, WasmEncoder} from "./wasmtypes/codec";
import {uint32Decode, uint32Encode} from "./wasmtypes/scuint32";

export class ScAssets {
    baseTokens: u64 = 0n;
    nftIDs: Set<ScNftID> = new Set();
    tokens: Map<string, ScBigInt> = new Map();

    public constructor(buf: Uint8Array | null) {
        if (buf === null || buf.length == 0) {
            return this;
        }
        const dec = new WasmDecoder(buf);
        this.baseTokens = uint64Decode(dec);

        let size = uint32Decode(dec);
        for (let i: u32 = 0; i < size; i++) {
            const tokenID = tokenIDDecode(dec);
            const amount = bigIntDecode(dec);
            this.tokens.set(ScDict.toKey(tokenID.id), amount);
        }

        size = uint32Decode(dec);
        for (let i: u32 = 0; i < size; i++) {
            const nftID = nftIDDecode(dec);
            this.nftIDs.add(nftID)
        }
    }

    public balances(): ScBalances {
        return new ScBalances(this);
    }

    public isEmpty(): bool {
        if (this.baseTokens != 0n) {
            return false;
        }
        const values = [...this.tokens.values()];
        for (let i = 0; i < values.length; i++) {
            if (!values[i].isZero()) {
                return false;
            }
        }
        return this.nftIDs.size == 0;
    }

    public toBytes(): Uint8Array {
        const enc = new WasmEncoder();
        uint64Encode(enc, this.baseTokens);

        let tokenIDs = this.tokenIDs();
        uint32Encode(enc, tokenIDs.length as u32);
        for (let i = 0; i < tokenIDs.length; i++) {
            const tokenID = tokenIDs[i]
            tokenIDEncode(enc, tokenID);
            const mapKey = ScDict.toKey(tokenID.id);
            const amount = this.tokens.get(mapKey)!;
            bigIntEncode(enc, amount);
        }

        uint32Encode(enc, this.nftIDs.size as u32);
        let arr = [...this.nftIDs.values()];
        for (let i = 0; i < arr.length; i++) {
            let nftID = arr[i];
            nftIDEncode(enc, nftID);
        }
        return enc.buf()
    }

    public tokenIDs(): ScTokenID[] {
        let tokenIDs: ScTokenID[] = [];
        const keys = [...this.tokens.keys()].sort();
        for (let i = 0; i < keys.length; i++) {
            const keyBytes = ScDict.fromKey(keys[i]);
            const tokenID = tokenIDFromBytes(keyBytes);
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

    public balance(tokenID: ScTokenID): ScBigInt {
        const mapKey = ScDict.toKey(tokenID.id);
        if (!this.assets.tokens.has(mapKey)) {
            return new ScBigInt();
        }
        return this.assets.tokens.get(mapKey)!;
    }

    public baseTokens(): u64 {
        return this.assets.baseTokens;
    }

    public isEmpty(): bool {
        return this.assets.isEmpty();
    }

    public nftIDs(): Set<ScNftID> {
        return this.assets.nftIDs;
    }

    public toBytes(): Uint8Array {
        return this.assets.toBytes();
    }

    public tokenIDs(): ScTokenID[] {
        return this.assets.tokenIDs();
    }
}

export class ScTransfer extends ScBalances {
    public constructor() {
        super(new ScAssets(null));
    }

    public static fromBalances(balances: ScBalances): ScTransfer {
        const transfer = ScTransfer.baseTokens(balances.baseTokens());
        const tokenIDs = balances.tokenIDs();
        for (let i = 0; i < tokenIDs.length; i++) {
            const tokenID = tokenIDs[i];
            transfer.set(tokenID, balances.balance(tokenID));
        }
        const nftIDs = [...balances.nftIDs().values()];
        for (let i = 0; i < nftIDs.length; i++) {
            transfer.addNFT(nftIDs[i]);
        }
        return transfer;
    }

    public static baseTokens(amount: u64): ScTransfer {
        const transfer = new ScTransfer();
        transfer.assets.baseTokens = amount;
        return transfer;
    }

    public static nft(nftID: ScNftID): ScTransfer {
        const transfer = new ScTransfer();
        transfer.addNFT(nftID);
        return transfer;
    }

    public static tokens(tokenID: ScTokenID, amount: ScBigInt): ScTransfer {
        const transfer = new ScTransfer();
        transfer.set(tokenID, amount);
        return transfer;
    }

    public addNFT(nftID: ScNftID): void {
        this.assets.nftIDs.add(nftID);
    }

    public set(tokenID: ScTokenID, amount: ScBigInt): void {
        const mapKey = ScDict.toKey(tokenID.id);
        this.assets.tokens.set(mapKey, amount);
    }
}
