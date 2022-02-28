// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {ScDict} from "./dict";

export class ScAssets {
    assets: Map<string, u64> = new Map();

    public constructor(buf: u8[]) {
        if (buf.length != 0) {
            const dec = new wasmtypes.WasmDecoder(buf);
            const size = wasmtypes.uint32FromBytes(dec.fixedBytes(wasmtypes.ScUint32Length));
            for (let i: u32 = 0; i < size; i++) {
                const color = wasmtypes.colorDecode(dec);
                this.assets.set(ScDict.toKey(color.id), wasmtypes.uint64FromBytes(dec.fixedBytes(wasmtypes.ScUint64Length)));
            }
        }
    }

    public balances(): ScBalances {
        return new ScBalances(this);
    }

    public toBytes(): u8[] {
        const keys = this.assets.keys().sort();
        const enc = new wasmtypes.WasmEncoder();
        enc.fixedBytes(wasmtypes.uint32ToBytes(keys.length as u32), wasmtypes.ScUint32Length);
        for (let i = 0; i < keys.length; i++) {
            const mapKey = keys[i]
            const colorId = ScDict.fromKey(mapKey);
            enc.fixedBytes(colorId, wasmtypes.ScColorLength);
            enc.fixedBytes(wasmtypes.uint64ToBytes(this.assets.get(mapKey)), wasmtypes.ScUint64Length);
        }
        return enc.buf()
    }
}

export class ScBalances {
    assets: Map<string, u64>;

    constructor(assets: ScAssets) {
        this.assets = assets.assets;
    }

    public balance(color: wasmtypes.ScColor): u64 {
        const mapKey = ScDict.toKey(color.id);
        if (!this.assets.has(mapKey)) {
            return 0;
        }
        return this.assets.get(mapKey);
    }

    public colors(): wasmtypes.ScColor[] {
        let colors: wasmtypes.ScColor[] = [];
        const keys = this.assets.keys();
        for (let i = 0; i < keys.length; i++) {
            const colorId = ScDict.fromKey(keys[i]);
            colors.push(wasmtypes.colorFromBytes(colorId));
        }
        return colors;
    }
}

export class ScTransfers extends ScAssets {
    public constructor() {
        super([]);
    }

    public static fromBalances(balances: ScBalances): ScTransfers {
        const transfer = new ScTransfers();
        const colors = balances.colors();
        for (let i = 0; i < colors.length; i++) {
            const color = colors[i];
            transfer.set(color, balances.balance(color));
        }
        return transfer;
    }

    public static iotas(amount: u64): ScTransfers {
        return ScTransfers.transfer(wasmtypes.IOTA, amount);
    }

    public static transfer(color: wasmtypes.ScColor, amount: u64): ScTransfers {
        const transfer = new ScTransfers();
        transfer.set(color, amount);
        return transfer;
    }

    public isEmpty(): bool {
        const keys = this.assets.keys();
        for (let i = 0; i < keys.length; i++) {
            if (this.assets.get(keys[i]) != 0) {
                return false;
            }
         }
        return true;
    }

    public set(color: wasmtypes.ScColor, amount: u64): void {
        this.assets.set(ScDict.toKey(color.id), amount);
    }
}
