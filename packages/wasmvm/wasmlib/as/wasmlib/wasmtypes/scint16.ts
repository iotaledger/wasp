// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from '../sandbox';
import {intFromString, WasmDecoder, WasmEncoder} from './codec';
import {Proxy} from './proxy';

export const ScInt16Length = 2;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function int16Decode(dec: WasmDecoder): i16 {
    return dec.vliDecode(16) as i16;
}

export function int16Encode(enc: WasmEncoder, value: i16): void {
    enc.vliEncode(value as i64);
}

export function int16FromBytes(buf: Uint8Array): i16 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScInt16Length) {
        panic('invalid Int16 length');
    }
    const ret: i16 = buf[1];
    return (ret << 8) | buf[0];
}

export function int16ToBytes(value: i16): Uint8Array {
    const buf = new Uint8Array(ScInt16Length);
    buf[0] = value as u8;
    buf[1] = (value >> 8) as u8;
    return buf;
}

export function int16FromString(value: string): i16 {
    return intFromString(value, 16) as i16;
}

export function int16ToString(value: i16): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableInt16 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return int16ToString(this.value());
    }

    value(): i16 {
        return int16FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableInt16 extends ScImmutableInt16 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: i16): void {
        this.proxy.set(int16ToBytes(value));
    }
}
