// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

export const ScInt16Length = 2;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function int16Decode(dec: wasmtypes.WasmDecoder): i16 {
    return dec.vliDecode(16) as i16;
}

export function int16Encode(enc: wasmtypes.WasmEncoder, value: i16): void {
    enc.vliEncode(value as i64);
}

export function int16FromBytes(buf: u8[]): i16 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScInt16Length) {
        panic("invalid Int16 length");
    }
    let ret: i16 = buf[1];
    return (ret << 8) | buf[0];
}

export function int16ToBytes(value: i16): u8[] {
    return [
        value as u8,
        (value >> 8) as u8,
    ];
}

export function int16ToString(value: i16): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableInt16 {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
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
