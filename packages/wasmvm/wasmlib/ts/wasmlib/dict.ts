// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {log} from './sandbox';
import {
    IKvStore,
    Proxy,
    stringFromBytes,
    WasmDecoder,
    WasmEncoder
} from './wasmtypes';

// returns a hex string representing the byte buffer
function hex(buf: Uint8Array): string {
    const hexa = '0123456789abcdef';
    let res = '';
    for (let i = 0; i < buf.length; i++) {
        const b = buf[i];
        res += hexa.charAt(b >> 4) + hexa.charAt(b & 0x0f);
    }
    return res;
}

function keya(key: Uint8Array): string {
    for (let i = 0; i < key.length; i++) {
        if (key[i] == 0x23) {
            return stringFromBytes(key.slice(0, i + 1)) + hex(key.slice(i + 1));
        }
        if (key[i] < 0x20 || key[i] > 0x7e) {
            return hex(key);
        }
    }
    return stringFromBytes(key);
}

function vala(val: Uint8Array): string {
    return hex(val);
}

export class ScDict implements IKvStore {
    dict: Map<string, Uint8Array> = new Map();

    public constructor(buf: Uint8Array | null) {
        if (buf !== null && buf.length != 0) {
            const dec = new WasmDecoder(buf);
           let size = dec.vluDecode(32) as u32;
            for (; size > 0; size--) {
                const keyLen = dec.vluDecode(32);
                const key = dec.fixedBytes(keyLen);
                const valLen = dec.vluDecode(32);
                const val = dec.fixedBytes(valLen);
                this.set(key, val);
            }
        }
    }

    static toKey(buf: Uint8Array): string {
        let key = '';
        for (let i = 0; i < buf.length; i++) {
            key += String.fromCharCode((buf[i] >> 4) + 0x40, (buf[i] & 0x0f) + 0x40);
        }
        return key;
    }

    static fromKey(key: string): Uint8Array {
        const buf = new Uint8Array(key.length / 2);
        for (let i = 0; i < key.length; i += 2) {
            const b1 = key.charCodeAt(i) as u8;
            const b2 = key.charCodeAt(i + 1) as u8;
            buf[i / 2] = (((b1 - 0x40) << 4) | (b2 - 0x40));
        }
        return buf;
    }

    public asProxy(): Proxy {
        return new Proxy(this);
    }

    delete(key: Uint8Array): void {
        // this.dump('delete');
        // log('dict.delete(' + keya(key) + ')');
        this.dict.delete(ScDict.toKey(key));
        // this.dump('Delete')
    }

    exists(key: Uint8Array): bool {
        const mapKey = ScDict.toKey(key);
        const ret = this.dict.has(mapKey);
        // this.dump('exists');
        // log('dict.exists(' + keya(key) + ') = ' + ret.toString());
        return ret;
    }

    get(key: Uint8Array): Uint8Array {
        // this.dump('get')
        const mapKey = ScDict.toKey(key);
        if (!this.dict.has(mapKey)) {
            // log('dict.get(' + keya(key) + ') = null');
            return new Uint8Array(0);
        }
        const value = this.dict.get(mapKey)!;
        // log('dict.get(' + keya(key) + ') = ' + vala(value));
        return value;
    }

    public immutable(): ScImmutableDict {
        return new ScImmutableDict(this);
    }

    set(key: Uint8Array, value: Uint8Array): void {
        // this.dump('set')
        // log('dict.set(' + keya(key) + ', ' + vala(value) + ')');
        this.dict.set(ScDict.toKey(key), value);
        // this.dump('Set')
    }

    public toBytes(): Uint8Array {
        if (this.dict.size == 0) {
            return new Uint8Array(1);
        }
        const keys = [...this.dict.keys()].sort();
        const enc = new WasmEncoder();
        enc.vluEncode(keys.length as u32);
        for (let i = 0; i < keys.length; i++) {
            const k = keys[i];
            const key = ScDict.fromKey(k);
            const val = this.dict.get(k)!;
            enc.vluEncode(key.length as u32);
            enc.fixedBytes(key, key.length as u32);
            enc.vluEncode(val.length as u32);
            enc.fixedBytes(val, val.length as u32);
        }
        return enc.buf();
    }

    protected dump(which: string): void {
        const keys = [...this.dict.keys()];
        for (let i = 0; i < keys.length; i++) {
            log('dict.' + which + '.' + i.toString() + '.' + keya(ScDict.fromKey(keys[i])) + ' = ' + vala(this.dict.get(keys[i])!));
        }
    }
}

export class ScImmutableDict {
    dict: Map<string, Uint8Array>;

    public constructor(dict: ScDict) {
        this.dict = dict.dict;
    }

    exists(key: Uint8Array): bool {
        return this.dict.has(ScDict.toKey(key));
    }

    get(key: Uint8Array): Uint8Array {
        const mapKey = ScDict.toKey(key);
        if (!this.dict.has(mapKey)) {
            return new Uint8Array(0);
        }
        return this.dict.get(mapKey)!;
    }
}
