// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {uint32Decode, uint32Encode} from "./scuint32";
import {stringToBytes} from "./scstring";

export interface IKvStore {
    delete(key: u8[]): void;

    exists(key: u8[]): bool;

    get(key: u8[]): u8[] | null;

    set(key: u8[], value: u8[]): void;
}

export class ScProxy {
    protected proxy: Proxy;

    public constructor(proxy: Proxy) {
        this.proxy = proxy;
    }
}

export class Proxy {
    _key: u8[] = [];
    kvStore: IKvStore;

    constructor(kvStore: IKvStore) {
        this.kvStore = kvStore;
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
        // TODO clearPrefix

        // clear the length counter
        this.delete();
    }

    public decoder(): WasmDecoder {
        return this._decoder(this.get());
    }

    protected _decoder(buf: u8[] | null): WasmDecoder {
        return new WasmDecoder(buf);
    }

    delete(): void {
        this.kvStore.delete(this._key);
    }

    protected element(index: u32): Proxy {
        let enc = this.encoder();
        uint32Encode(enc, index);
        return this.sub('#'.charCodeAt(0) as u8, enc.buf());
    }

    public encoder(): WasmEncoder {
        return new WasmEncoder();
    }

    exists(): bool {
        return this.kvStore.exists(this._key);
    }

    public expand(length: u32): void {
        // update the length counter
        let enc = this.encoder();
        uint32Encode(enc, length);
        this.set(enc.buf());
    }

    get(): u8[] | null {
        return this.kvStore.get(this._key);
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
    public key(key: u8[]): Proxy {
        return this.sub('.'.charCodeAt(0) as u8, key);
    }

    // Length returns the number of elements in an Array
    // Never try to access an index >= Length()
    public length(): u32 {
        // get the length counter
        let buf = this.get();
        if (buf == null) {
            return 0;
        }
        return uint32Decode(this._decoder(buf));
    }

    protected proxy(kvStore: IKvStore, key: u8[]): Proxy {
        const res = new Proxy(kvStore);
        res._key = key;
        return res;
    }

    // Root returns a Proxy for an element of a root container (Params/Results/State).
    // The key is always a string.
    public root(key: string): Proxy {
        return this.proxy(this.kvStore, stringToBytes(key));
    }

    set(value: u8[]): void {
        this.kvStore.set(this._key, value);
    }

    // sub returns a proxy for an element of a container.
    // The separator is significant, it prevents potential clashes with other elements.
    // Different separators can be used to indicate different sub-containers
    protected sub(sep: u8, key: u8[]): Proxy {
        if (this._key.length == 0) {
            // this must be a root proxy
            return this.proxy(this.kvStore, key.slice(0));
        }
        return this.proxy(this.kvStore, this._key.concat([sep]).concat(key));
    }
}