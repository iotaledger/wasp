// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from '../sandbox';
import {uintFromString, WasmDecoder, WasmEncoder} from './codec';
import {Proxy} from './proxy';

export const ScUint8Length = 1;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function uint8Decode(dec: WasmDecoder): u8 {
    return dec.byte();
}

export function uint8Encode(enc: WasmEncoder, value: u8): void {
    enc.byte(value);
}

export function uint8FromBytes(buf: Uint8Array): u8 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScUint8Length) {
        panic('invalid Uint8 length');
    }
    return buf[0];
}

export function uint8ToBytes(value: u8): Uint8Array {
    const buf = new Uint8Array(ScUint8Length);
    buf[0] = value as u8;
    return buf;
}

export function uint8FromString(value: string): u8 {
    return uintFromString(value, 8) as u8;
}

export function uint8ToString(value: u8): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableUint8 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return uint8ToString(this.value());
    }

    value(): u8 {
        return uint8FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableUint8 extends ScImmutableUint8 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: u8): void {
        this.proxy.set(uint8ToBytes(value));
    }
}
