// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function stringDecode(dec: WasmDecoder): string {
    return stringFromBytes(dec.bytes());
}

export function stringEncode(enc: WasmEncoder, value: string): void {
    enc.bytes(stringToBytes(value));
}

export function stringFromBytes(buf: u8[] | null): string {
    return buf == null ? "" : String.UTF8.decodeUnsafe(buf.dataStart, buf.length);
}

export function stringToBytes(value: string): u8[] {
    let arrayBuffer = String.UTF8.encode(value);
    let u8Array = Uint8Array.wrap(arrayBuffer)
    let ret: u8[] = new Array(u8Array.length);
    for (let i = 0; i < ret.length; i++) {
        ret[i] = u8Array[i];
    }
    return ret;
}

export function stringToString(value: string): string {
    return value;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableString {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return this.value().toString();
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
