// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

export const ScUint32Length = 4;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function uint32Decode(dec: WasmDecoder): u32 {
    return dec.vluDecode(32) as u32;
}

export function uint32Encode(enc: WasmEncoder, value: u32): void {
    enc.vluEncode(value as u64);
}

export function uint32FromBytes(buf: u8[] | null): u32 {
    if (buf == null) {
        return 0;
    }
    if (buf.length != ScUint32Length) {
        panic("invalid Uint32 length");
    }
    let ret: u32 = buf[3];
    ret = (ret << 8) | buf[2];
    ret = (ret << 8) | buf[1];
    return (ret << 8) | buf[0];
}

export function uint32ToBytes(value: u32): u8[] {
    return [
        value as u8,
        (value >> 8) as u8,
        (value >> 16) as u8,
        (value >> 24) as u8,
    ];
}

export function uint32ToString(value: u32): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableUint32 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return uint32ToString(this.value());
    }

    value(): u32 {
        return uint32FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableUint32 extends ScImmutableUint32 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: u32): void {
        this.proxy.set(uint32ToBytes(value));
    }
}
