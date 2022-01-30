// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

export const ScUint16Length = 2;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function uint16Decode(dec: WasmDecoder): u16 {
    return dec.vluDecode(16) as u16;
}

export function uint16Encode(enc: WasmEncoder, value: u16): void {
    enc.vluEncode(value as u64);
}

export function uint16FromBytes(buf: u8[] | null): u16 {
    if (buf == null) {
        return 0;
    }
    if (buf.length != ScUint16Length) {
        panic("invalid Uint16 length");
    }
    let ret: u16 = buf[1];
    return (ret << 8) | buf[0];
}

export function uint16ToBytes(value: u16): u8[] {
    return [
        value as u8,
        (value >> 8) as u8,
    ];
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
