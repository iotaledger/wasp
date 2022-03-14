// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

export const ScUint64Length = 8;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function uint64Decode(dec: wasmtypes.WasmDecoder): u64 {
    return dec.vluDecode(64);
}

export function uint64Encode(enc: wasmtypes.WasmEncoder, value: u64): void {
    enc.vluEncode(value);
}

export function uint64FromBytes(buf: u8[]): u64 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScUint64Length) {
        panic("invalid Uint64 length");
    }
    let ret: u64 = buf[7];
    ret = (ret << 8) | buf[6];
    ret = (ret << 8) | buf[5];
    ret = (ret << 8) | buf[4];
    ret = (ret << 8) | buf[3];
    ret = (ret << 8) | buf[2];
    ret = (ret << 8) | buf[1];
    return (ret << 8) | buf[0];
}

export function uint64ToBytes(value: u64): u8[] {
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

export function uint64ToString(value: u64): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableUint64 {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return uint64ToString(this.value());
    }

    value(): u64 {
        return uint64FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableUint64 extends ScImmutableUint64 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: u64): void {
        this.proxy.set(uint64ToBytes(value));
    }
}
