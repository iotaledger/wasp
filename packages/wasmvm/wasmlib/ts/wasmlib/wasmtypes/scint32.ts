// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

export const ScInt32Length = 4;

import {panic} from '../sandbox';
import {intFromString, WasmDecoder, WasmEncoder} from './codec';
import {Proxy} from './proxy';

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function int32Decode(dec: WasmDecoder): i32 {
    return dec.vliDecode(32) as i32;
}

export function int32Encode(enc: WasmEncoder, value: i32): void {
    enc.vliEncode(value as i32);
}

export function int32FromBytes(buf: Uint8Array): i32 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScInt32Length) {
        panic('invalid Int32 length');
    }
    let ret: i32 = buf[3];
    ret = (ret & 0x80) ? ret - 0x100 : ret;
    ret = (ret << 8) | buf[2];
    ret = (ret << 8) | buf[1];
    ret = (ret << 8) | buf[0];
    return ret;
}

export function int32ToBytes(value: i32): Uint8Array {
    const buf = new Uint8Array(ScInt32Length);
    buf[0] = value as u8;
    buf[1] = (value >> 8) as u8;
    buf[2] = (value >> 16) as u8;
    buf[3] = (value >> 24) as u8;
    return buf;
}

export function int32FromString(value: string): i32 {
    return intFromString(value, 32) as i32;
}

export function int32ToString(value: i32): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableInt32 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return int32ToString(this.value());
    }

    value(): i32 {
        return int32FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableInt32 extends ScImmutableInt32 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: i32): void {
        this.proxy.set(int32ToBytes(value));
    }
}
