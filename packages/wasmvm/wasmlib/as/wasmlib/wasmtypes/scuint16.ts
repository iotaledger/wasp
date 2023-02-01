// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

export const ScUint16Length = 2;

import {panic} from '../sandbox';
import {uintFromString, WasmDecoder, WasmEncoder} from './codec';
import {Proxy} from './proxy';

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function uint16Decode(dec: WasmDecoder): u16 {
    return dec.vluDecode(16) as u16;
}

export function uint16Encode(enc: WasmEncoder, value: u16): void {
    enc.vluEncode(value as u64);
}

export function uint16FromBytes(buf: Uint8Array): u16 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScUint16Length) {
        panic('invalid Uint16 length');
    }
    const ret: u16 = buf[1];
    return (ret << 8) | buf[0];
}

export function uint16ToBytes(value: u16): Uint8Array {
    const buf = new Uint8Array(ScUint16Length);
    buf[0] = value as u8;
    buf[1] = (value >> 8) as u8;
    return buf;
}

export function uint16FromString(value: string): u16 {
    return uintFromString(value, 16) as u16;
}

export function uint16ToString(value: u16): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableUint16 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return uint16ToString(this.value());
    }

    value(): u16 {
        return uint16FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableUint16 extends ScImmutableUint16 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: u16): void {
        this.proxy.set(uint16ToBytes(value));
    }
}
