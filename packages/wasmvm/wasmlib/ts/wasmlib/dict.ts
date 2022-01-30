// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {log} from "./sandbox";

export class ScDict implements wasmtypes.IKvStore {
    dict: Map<string, u8[]> = new Map();

    static toKey(buf: u8[]): string {
        let key = "";
        for (let i = 0; i < buf.length; i++) {
            key += String.fromCharCode((buf[i] >> 4) + 0x40, (buf[i] & 0x0f) + 0x40);
        }
        //log("toKey(" + wasmtypes.hex(buf) + ")=" + key);
        return key;
    }

    static fromKey(key: string): u8[] {
        let buf: u8[] = new Array(key.length / 2);
        for (let i = 0; i < key.length; i += 2) {
            const b1 = key.charCodeAt(i) as u8;
            const b2 = key.charCodeAt(i + 1) as u8;
            buf[i / 2] = (((b1 - 0x40) << 4) | (b2 - 0x40));
        }
        //log("fromKey(" + key + ")=" + wasmtypes.hex(buf));
        return buf;
    }

    public constructor(buf: u8[]) {
        if (buf.length != 0) {
            const dec = new wasmtypes.WasmDecoder(buf);
            const size = wasmtypes.uint32FromBytes(dec.fixedBytes(wasmtypes.ScUint32Length));
            for (let i: u32 = 0; i < size; i++) {
                const keyLen = wasmtypes.uint16FromBytes(dec.fixedBytes(wasmtypes.ScUint16Length));
                const key = dec.fixedBytes(keyLen as u32);
                const valLen = wasmtypes.uint32FromBytes(dec.fixedBytes(wasmtypes.ScUint32Length));
                const val = dec.fixedBytes(valLen);
                this.set(key, val);
            }
        }
    }

    public asProxy(): wasmtypes.Proxy {
        return new wasmtypes.Proxy(this);
    }

    delete(key: u8[]): void {
        this.dict.delete(ScDict.toKey(key));
    }

    exists(key: u8[]): bool {
        return this.dict.has(ScDict.toKey(key));
    }

    get(key: u8[]): u8[] | null {
        const mapKey = ScDict.toKey(key);
        if (! this.dict.has(mapKey)) {
            return null;
        }
        return this.dict.get(mapKey);
    }

    public immutable(): ScImmutableDict {
        return new ScImmutableDict(this);
    }

    set(key: u8[], value: u8[]): void {
        this.dict.set(ScDict.toKey(key), value);
    }

    public toBytes(): u8[] {
        if (this.dict.size == 0) {
            return [0, 0, 0, 0];
        }
        const keys = this.dict.keys().sort();
        const enc = new wasmtypes.WasmEncoder();
        enc.fixedBytes(wasmtypes.uint32ToBytes(keys.length as u32), wasmtypes.ScUint32Length);
        for (let i = 0; i < keys.length; i++) {
            const k = keys[i];
            const key = ScDict.fromKey(k);
            let val = this.dict.get(k);
            if (val === null) {
                val = [];
            }
            enc.fixedBytes(wasmtypes.uint16ToBytes(key.length as u16), wasmtypes.ScUint16Length);
            enc.fixedBytes(key, key.length as u32);
            enc.fixedBytes(wasmtypes.uint32ToBytes(val.length as u32), wasmtypes.ScUint32Length);
            enc.fixedBytes(val, val.length as u32);
        }
        return enc.buf();
    }
}

export class ScImmutableDict {
    dict: Map<string, u8[]>;

    public constructor(dict: ScDict) {
        this.dict = dict.dict;
    }

    exists(key: u8[]): bool {
        return this.dict.has(ScDict.toKey(key));
    }

    get(key: u8[]): u8[] | null {
        const mapKey = ScDict.toKey(key);
        if (! this.dict.has(mapKey)) {
            return null;
        }
        return this.dict.get(mapKey);
    }
}
