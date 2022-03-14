// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

export const ScUint8Length = 1;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function uint8Decode(dec: wasmtypes.WasmDecoder): u8 {
    return dec.byte();
}

export function uint8Encode(enc: wasmtypes.WasmEncoder, value: u8): void {
    enc.byte(value);
}

export function uint8FromBytes(buf: u8[]): u8 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScUint8Length) {
        panic("invalid Uint8 length");
    }
    return buf[0];
}

export function uint8ToBytes(value: u8): u8[] {
    return [value];
}

export function uint8ToString(value: u8): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableUint8 {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
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
