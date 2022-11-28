// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {uint32Decode, uint32Encode} from "./scuint32";
import {WasmDecoder, WasmEncoder} from "./codec";
import {stringToBytes} from "./scstring";

export interface IKvStore {
    delete(key: Uint8Array): void;

    exists(key: Uint8Array): bool;

    get(key: Uint8Array): Uint8Array;

    set(key: Uint8Array, value: Uint8Array): void;
}

export class ScProxy {
    protected proxy: Proxy;

    public constructor(proxy: Proxy) {
        this.proxy = proxy;
    }
}

export class Proxy {
    _key: Uint8Array;
    kvStore: IKvStore;

    constructor(kvStore: IKvStore) {
        this.kvStore = kvStore;
        this._key = new Uint8Array(0);
    }

    // alternative constructor
    protected proxy(kvStore: IKvStore, key: Uint8Array): Proxy {
        const res = new Proxy(kvStore);
        res._key = key;
        return res;
    }

    // Append returns a Proxy for a newly appended null element
    // Note that this will essentially return the element at Length()
    public append(): Proxy {
        const length = this.length();
        this.expand(length + 1);
        return this.element(length);
    }

    // ClearArray clears an array by deleting all elements
    // TODO Note that this does not delete recursive container elements
    public clearArray(): void {
        for (let length = this.length(); length != 0; length--) {
            this.element(length - 1).delete();
        }

        // clear the length counter
        this.delete();
    }

    // ClearMap clears a map by deleting all elements
    // TODO Note that this does not delete recursive container elements
    public clearMap(): void {
        // TODO clear prefix

        // clear the length counter
        this.delete();
    }

    delete(): void {
        //log(this.id.toString() + ".delete(" + keya(this._key) + ")");
        this.kvStore.delete(this._key);
    }

    protected element(index: u32): Proxy {
        let enc = new WasmEncoder();
        uint32Encode(enc, index);
        // 0x23 is '#'
        return this.sub(0x23, enc.buf());
    }

    exists(): bool {
        //log(this.id.toString() + ".exists(" + keya(this._key) + ")");
        return this.kvStore.exists(this._key);
    }

    //TODO have a Grow function that grows an array?
    protected expand(length: u32): void {
        // update the length counter
        let enc = new WasmEncoder();
        uint32Encode(enc, length);
        this.set(enc.buf());
    }

    get(): Uint8Array {
        const buf = this.kvStore.get(this._key);
        //log(this.id.toString() + ".get(" + keya(this._key) + ") = " + vala(buf));
        return buf;
    }

    // Index gets a Proxy for an element of an Array by its index
    public index(index: u32): Proxy {
        const size = this.length();
        if (index >= size) {
            if (index == size) {
                panic("invalid index: use append");
            }
            panic("invalid index");
        }
        return this.element(index);
    }

    // Key gets a Proxy for an element of a Map by its key
    public key(key: Uint8Array): Proxy {
        // 0x2e is '.'
        return this.sub(0x2e, key);
    }

    // Length returns the number of elements in an Array
    // Never try to access an index >= Length()
    public length(): u32 {
        // get the length counter
        let buf = this.get();
        if (buf.length == 0) {
            return 0;
        }
        const dec = new WasmDecoder(buf)
        return uint32Decode(dec);
    }

    // Root returns a Proxy for an element of a root container (Params/Results/State).
    // The key is always a string.
    public root(key: string): Proxy {
        return this.proxy(this.kvStore, stringToBytes(key));
    }

    set(value: Uint8Array): void {
        //log(this.id.toString() + ".set(" + keya(this._key) + ") = " + vala(value));
        this.kvStore.set(this._key, value);
    }

    // sub returns a proxy for an element of a container.
    // The separator is significant, it prevents potential clashes with other elements.
    // Different separators can be used to indicate different sub-containers
    protected sub(sep: u8, key: Uint8Array): Proxy {
        if (this._key.length == 0) {
            // this must be a root proxy
            return this.proxy(this.kvStore, key.slice(0));
        }
        const subKey = new Uint8Array(this._key.length + 1 + key.length);
        subKey.set(this._key);
        subKey[this._key.length] = sep;
        subKey.set(key, this._key.length+1)
        return this.proxy(this.kvStore, subKey);
    }
}