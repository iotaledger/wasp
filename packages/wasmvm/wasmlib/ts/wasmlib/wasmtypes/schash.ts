// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";
import {addressToBytes} from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScHashLength = 32;

export class ScHash {
    id: u8[] = wasmtypes.zeroes(ScHashLength);

    public equals(other: ScHash): bool {
        return wasmtypes.bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return hashToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return hashToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function hashDecode(dec: wasmtypes.WasmDecoder): ScHash {
    return hashFromBytesUnchecked(dec.fixedBytes(ScHashLength));
}

export function hashEncode(enc: wasmtypes.WasmEncoder, value: ScHash): void {
    enc.fixedBytes(value.id, ScHashLength);
}

export function hashFromBytes(buf: u8[]): ScHash {
    if (buf.length == 0) {
        return new ScHash();
    }
    if (buf.length != ScHashLength) {
        panic("invalid Hash length");
    }
    return hashFromBytesUnchecked(buf);
}

export function hashToBytes(value: ScHash): u8[] {
    return value.id;
}

export function hashFromString(value: string): ScHash {
    return hashFromBytes(wasmtypes.base58Decode(value));
}

export function hashToString(value: ScHash): string {
    // TODO standardize human readable string
    return wasmtypes.base58Encode(hashToBytes(value));
}

function hashFromBytesUnchecked(buf: u8[]): ScHash {
    let o = new ScHash();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableHash {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return hashToString(this.value());
    }

    value(): ScHash {
        return hashFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableHash extends ScImmutableHash {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScHash): void {
        this.proxy.set(hashToBytes(value));
    }
}
