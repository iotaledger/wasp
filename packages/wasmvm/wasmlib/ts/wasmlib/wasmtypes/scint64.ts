// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

export const ScInt64Length = 8;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function int64Decode(dec: WasmDecoder): i64 {
    return dec.vliDecode(64);
}

export function int64Encode(enc: WasmEncoder, value: i64): void {
    enc.vliEncode(value);
}

export function int64FromBytes(buf: u8[] | null): i64 {
    if (buf == null) {
        return 0;
    }
    if (buf.length != ScInt64Length) {
        panic("invalid Int64 length");
    }
    let ret: i64 = buf[7];
    ret = (ret << 8) | buf[6];
    ret = (ret << 8) | buf[5];
    ret = (ret << 8) | buf[4];
    ret = (ret << 8) | buf[3];
    ret = (ret << 8) | buf[2];
    ret = (ret << 8) | buf[1];
    return (ret << 8) | buf[0];
}

export function int64ToBytes(value: i64): u8[] {
    return [
        value as u8,
        (value >> 8) as u8,
        (value >> 16) as u8,
        (value >> 24) as u8,
        (value >> 32) as u8,
        (value >> 40) as u8,
        (value >> 48) as u8,
        (value >> 56) as u8,
    ];
}

export function int64ToString(value: i64): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableInt64 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return int64ToString(this.value());
    }

    value(): i64 {
        return int64FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableInt64 extends ScImmutableInt64 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: i64): void {
        this.proxy.set(int64ToBytes(value));
    }
}
