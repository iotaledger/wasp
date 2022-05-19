// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

export const ScInt32Length = 4;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function int32Decode(dec: wasmtypes.WasmDecoder): i32 {
    return dec.vliDecode(32) as i32;
}

export function int32Encode(enc: wasmtypes.WasmEncoder, value: i32): void {
    enc.vliEncode(value as i64);
}

export function int32FromBytes(buf: u8[]): i32 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScInt32Length) {
        panic("invalid Int32 length");
    }
    let ret: i32 = buf[3];
    ret = (ret << 8) | buf[2];
    ret = (ret << 8) | buf[1];
    return (ret << 8) | buf[0];
}

export function int32ToBytes(value: i32): u8[] {
    return [
        value as u8,
        (value >> 8) as u8,
        (value >> 16) as u8,
        (value >> 24) as u8,
    ];
}

export function int32FromString(value: string): i32 {
    return wasmtypes.intFromString(value, 32) as i32;
}

export function int32ToString(value: i32): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableInt32 {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
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
