// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {WasmDecoder, WasmEncoder} from './codec';
import {Proxy} from './proxy';
import {bytesFromUint8Array} from './scbytes';
import {uint16Decode, uint16Encode} from "./scuint16";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function stringDecode(dec: WasmDecoder): string {
    const length = uint16Decode(dec);
    return stringFromBytes(dec.fixedBytes(length as u32));
}

export function stringEncode(enc: WasmEncoder, value: string): void {
    const buf = stringToBytes(value);
    uint16Encode(enc, buf.length as u16);
    enc.fixedBytes(buf, buf.length as u32);
}

export function stringFromBytes(buf: Uint8Array): string {
    return String.UTF8.decodeUnsafe(buf.dataStart, buf.length);
}

export function stringToBytes(value: string): Uint8Array {
    const arrayBuffer = String.UTF8.encode(value);
    const u8Array = Uint8Array.wrap(arrayBuffer);
    return bytesFromUint8Array(u8Array);
}

export function stringFromString(value: string): string {
    return value;
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
