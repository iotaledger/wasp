// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {ScDict} from './dict';
import {ScTokenID, ScTokenIDLength, tokenIDDecode, tokenIDEncode, tokenIDFromBytes} from './wasmtypes/sctokenid';
import {ScUint64Length, uint64FromBytes, uint64ToBytes} from './wasmtypes/scuint64';
import {bigIntDecode, bigIntEncode, ScBigInt} from './wasmtypes/scbigint';
import {nftIDDecode, nftIDEncode, ScNftID} from './wasmtypes/scnftid';
import {WasmDecoder, WasmEncoder} from './wasmtypes/codec';
import {uint8Decode, uint8Encode} from "./wasmtypes/scuint8";

const hasBaseTokens: u8 = 0x80;
const hasNativeTokens: u8 = 0x40;
const hasNFTs: u8 = 0x20;

export class ScAssets {
    baseTokens: u64 = 0;
    nativeTokens: Map<string, ScBigInt> = new Map();
    nfts: Map<string, ScNftID> = new Map();

    public constructor(buf: Uint8Array | null) {
        if (buf === null || buf.length == 0) {
            return this;
        }

        const dec = new WasmDecoder(buf);
        const flags = uint8Decode(dec);
        if (flags == 0x00) {
            return this;
        }

        if ((flags & hasBaseTokens) != 0) {
            this.baseTokens = dec.vluDecode(64);
        }
        if ((flags & hasNativeTokens) != 0) {
            let size = dec.vluDecode(16);
            for (; size > 0; size--) {
                const tokenID = tokenIDDecode(dec);
                const amount = bigIntDecode(dec);
                this.nativeTokens.set(ScDict.toKey(tokenID.id), amount);
            }
        }
        if ((flags & hasNFTs) != 0) {
            let size = dec.vluDecode(16);
            for (; size > 0; size--) {
                const nftID = nftIDDecode(dec);
                this.nfts.set(ScDict.toKey(nftID.id), nftID);
            }
        }
    }

    public balances(): ScBalances {
        return new ScBalances(this);
    }

    public isEmpty(): bool {
        if (this.baseTokens != 0) {
            return false;
        }
        const values = this.nativeTokens.values();
        for (let i = 0; i < values.length; i++) {
            if (!values[i].isZero()) {
                return false;
            }
        }
        return this.nfts.size == 0;
    }

    public nftIDs(): ScNftID[] {
        const nftIDs: ScNftID[] = [];
        const keys = this.nfts.keys().sort();
        for (let i = 0; i < keys.length; i++) {
            const nftID = this.nfts.get(keys[i]);
            nftIDs.push(nftID);
        }
        return nftIDs;
    }

    public toBytes(): Uint8Array {
        const enc = new WasmEncoder();
        if (this.isEmpty()) {
            return new Uint8Array(1);
        }

        let flags = 0x00 as u8;
        if (this.baseTokens != 0) {
            flags |= hasBaseTokens;
        }
        if (this.nativeTokens.size != 0) {
            flags |= hasNativeTokens;
        }
        if (this.nfts.size != 0) {
            flags |= hasNFTs;
        }
        uint8Encode(enc, flags);

        if ((flags & hasBaseTokens) != 0) {
            enc.vluEncode(this.baseTokens);
        }
        if ((flags & hasNativeTokens) != 0) {
            const keys = this.nativeTokens.keys().sort();
            enc.vluEncode(keys.length as u64);
            for (let i = 0; i < keys.length; i++) {
                const tokenID = ScDict.fromKey(keys[i]);
                enc.fixedBytes(tokenID, ScTokenIDLength);
                const amount = this.nativeTokens.get(keys[i]);
                bigIntEncode(enc, amount);
            }
        }
        if ((flags & hasNFTs) != 0) {
            const keys = this.nfts.keys().sort();
            enc.vluEncode(keys.length as u64);
            for (let i = 0; i < keys.length; i++) {
                const nftID = this.nfts.get(keys[i]);
                nftIDEncode(enc, nftID);
            }
        }
        return enc.buf();
    }

    public tokenIDs(): ScTokenID[] {
        const tokenIDs: ScTokenID[] = [];
        const keys = this.nativeTokens.keys().sort();
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
        if (!this.assets.nativeTokens.has(mapKey)) {
            return new ScBigInt();
        }
        return this.assets.nativeTokens.get(mapKey);
    }

    public baseTokens(): u64 {
        return this.assets.baseTokens;
    }

    public isEmpty(): bool {
        return this.assets.isEmpty();
    }

    public nftIDs(): ScNftID[] {
        return this.assets.nftIDs();
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
        const nftIDs = balances.nftIDs();
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
        this.assets.nfts.set(ScDict.toKey(nftID.id), nftID);
    }

    public set(tokenID: ScTokenID, amount: ScBigInt): void {
        const mapKey = ScDict.toKey(tokenID.id);
        this.assets.nativeTokens.set(mapKey, amount);
    }
}
