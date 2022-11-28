// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {hexDecode, hexEncode, WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

export function bytesCompare(lhs: Uint8Array, rhs: Uint8Array): i32 {
    const size = (lhs.length < rhs.length) ? lhs.length : rhs.length;
    for (let i = 0; i < size; i++) {
        if (lhs[i] != rhs[i]) {
            return (lhs[i] > rhs[i]) ? 1 : -1;
        }
    }
    return (lhs.length > size) ? 1 : (rhs.length > size) ? -1 : 0;
}

export function bytesEquals(lhs: Uint8Array, rhs: Uint8Array): bool {
    if (lhs.length != rhs.length) {
        return false;
    }
    for (let i = 0; i < lhs.length; i++) {
        if (lhs[i] != rhs[i]) {
            return false;
        }
    }
    return true;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function bytesDecode(dec: WasmDecoder): Uint8Array {
    return dec.bytes();
}

export function bytesEncode(enc: WasmEncoder, value: Uint8Array): void {
    enc.bytes(value);
}

export function bytesFromBytes(buf: Uint8Array): Uint8Array {
    return buf;
}

export function bytesFromUint8Array(arr: Uint8Array): Uint8Array {
    return arr.slice();
}

export function bytesToBytes(buf: Uint8Array): Uint8Array {
    return buf;
}

export function bytesToUint8Array(buf: Uint8Array): Uint8Array {
    return buf.slice();
}

export function bytesFromString(value: string): Uint8Array {
    return hexDecode(value);
}

export function bytesToString(value: Uint8Array): string {
    return hexEncode(value);
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableBytes {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return bytesToString(this.value());
    }

    value(): Uint8Array {
        return bytesFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableBytes extends ScImmutableBytes {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: Uint8Array): void {
        this.proxy.set(bytesToBytes(value));
    }
}
