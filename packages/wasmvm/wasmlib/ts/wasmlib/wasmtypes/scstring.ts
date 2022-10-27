// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function stringDecode(dec: wasmtypes.WasmDecoder): string {
    return stringFromBytes(dec.bytes());
}

export function stringEncode(enc: wasmtypes.WasmEncoder, value: string): void {
    enc.bytes(stringToBytes(value));
}

export function stringFromBytes(buf: u8[]): string {
    return String.UTF8.decodeUnsafe(buf.dataStart, buf.length);
}

export function stringToBytes(value: string): u8[] {
    let arrayBuffer = String.UTF8.encode(value);
    let u8Array = Uint8Array.wrap(arrayBuffer)
    return wasmtypes.bytesFromUint8Array(u8Array);
}

export function stringFromString(value: string): string {
    return value;
}

export function stringToString(value: string): string {
    return value;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableString {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return this.value();
    }

    value(): string {
        return stringFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableString extends ScImmutableString {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: string): void {
        this.proxy.set(stringToBytes(value));
    }
}
