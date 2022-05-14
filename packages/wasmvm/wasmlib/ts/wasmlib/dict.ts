// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {log} from "./sandbox";

// returns a hex string representing the byte buffer
function hex(buf: u8[]): string {
    const hexa = "0123456789abcdef";
    let res = "";
    for (let i = 0; i < buf.length; i++) {
        const b = buf[i];
        res += hexa.charAt(b >> 4) + hexa.charAt(b & 0x0f);
    }
    return res;
}

function keya(key: u8[]): string {
    for (let i = 0; i < key.length; i++) {
        if (key[i] == 0x23) {
            return wasmtypes.stringFromBytes(key.slice(0, i + 1)) + hex(key.slice(i + 1));
        }
        if (key[i] < 0x20 || key[i] > 0x7e) {
            return hex(key);
        }
    }
    return wasmtypes.stringFromBytes(key);
}

function vala(val: u8[]): string {
    return hex(val);
}

export class ScDict implements wasmtypes.IKvStore {
    dict: Map<string, u8[]> = new Map();

    static toKey(buf: u8[]): string {
        let key = "";
        for (let i = 0; i < buf.length; i++) {
            key += String.fromCharCode((buf[i] >> 4) + 0x40, (buf[i] & 0x0f) + 0x40);
        }
        return key;
    }

    static fromKey(key: string): u8[] {
        let buf = new Array<u8>(key.length / 2);
        for (let i = 0; i < key.length; i += 2) {
            const b1 = key.charCodeAt(i) as u8;
            const b2 = key.charCodeAt(i + 1) as u8;
            buf[i / 2] = (((b1 - 0x40) << 4) | (b2 - 0x40));
        }
        return buf;
    }

    public constructor(buf: u8[]) {
        if (buf.length != 0) {
            const dec = new wasmtypes.WasmDecoder(buf);
            const size = wasmtypes.uint32FromBytes(dec.fixedBytes(wasmtypes.ScUint32Length));
            for (let i: u32 = 0; i < size; i++) {
                const keyBuf = dec.fixedBytes(wasmtypes.ScUint16Length);
                const keyLen = wasmtypes.uint16FromBytes(keyBuf);
                const key = dec.fixedBytes(keyLen as u32);
                const valBuf = dec.fixedBytes(wasmtypes.ScUint32Length);
                const valLen = wasmtypes.uint32FromBytes(valBuf);
                const val = dec.fixedBytes(valLen);
                this.set(key, val);
            }
        }
    }

    public asProxy(): wasmtypes.Proxy {
        return new wasmtypes.Proxy(this);
    }

    delete(key: u8[]): void {
        // this.dump("delete");
        // log("dict.delete(" + keya(key) + ")");
        this.dict.delete(ScDict.toKey(key));
        // this.dump("Delete")
    }

    protected dump(which: string): void {
        const keys = this.dict.keys()
        for (let i = 0; i < keys.length; i++) {
            log("dict." + which + "." + i.toString() + "." + keya(ScDict.fromKey(keys[i])) + " = " + vala(this.dict.get(keys[i])));
        }
    }

    exists(key: u8[]): bool {
        const mapKey = ScDict.toKey(key);
        const ret = this.dict.has(mapKey);
        // this.dump("exists");
        // log("dict.exists(" + keya(key) + ") = " + ret.toString());
        return ret;
    }

    get(key: u8[]): u8[] {
        // this.dump("get")
        const mapKey = ScDict.toKey(key);
        if (!this.dict.has(mapKey)) {
            // log("dict.get(" + keya(key) + ") = null");
            return [];
        }
        const value = this.dict.get(mapKey);
        // log("dict.get(" + keya(key) + ") = " + vala(value));
        return value;
    }

    public immutable(): ScImmutableDict {
        return new ScImmutableDict(this);
    }

    set(key: u8[], value: u8[]): void {
        // this.dump("set")
        // log("dict.set(" + keya(key) + ", " + vala(value) + ")");
        this.dict.set(ScDict.toKey(key), value);
        // this.dump("Set")
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

    get(key: u8[]): u8[] {
        const mapKey = ScDict.toKey(key);
        if (!this.dict.has(mapKey)) {
            return [];
        }
        return this.dict.get(mapKey);
    }
}
