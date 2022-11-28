// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {hexDecode, hexEncode, WasmDecoder, WasmEncoder, zeroes} from "./codec";
import {Proxy} from "./proxy";
import {bytesCompare} from "./scbytes";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScHashLength = 32;

export class ScHash {
    id: Uint8Array = zeroes(ScHashLength);

    public equals(other: ScHash): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): Uint8Array {
        return hashToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return hashToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function hashDecode(dec: WasmDecoder): ScHash {
    return hashFromBytesUnchecked(dec.fixedBytes(ScHashLength));
}

export function hashEncode(enc: WasmEncoder, value: ScHash): void {
    enc.fixedBytes(value.id, ScHashLength);
}

export function hashFromBytes(buf: Uint8Array | null): ScHash {
    if (buf === null || buf.length == 0) {
        return new ScHash();
    }
    if (buf.length != ScHashLength) {
        panic("invalid Hash length");
    }
    return hashFromBytesUnchecked(buf);
}

export function hashToBytes(value: ScHash): Uint8Array {
    return value.id;
}

export function hashFromString(value: string): ScHash {
    return hashFromBytes(hexDecode(value));
}

export function hashToString(value: ScHash): string {
    return hexEncode(hashToBytes(value));
}

function hashFromBytesUnchecked(buf: Uint8Array): ScHash {
    let o = new ScHash();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableHash {
    proxy: Proxy;

    constructor(proxy: Proxy) {
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
