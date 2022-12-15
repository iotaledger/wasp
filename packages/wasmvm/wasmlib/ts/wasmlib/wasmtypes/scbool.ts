// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

export const ScBoolLength = 1;
export const ScBoolFalse = 0x00
export const ScBoolTrue = 0x01

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function boolDecode(dec: WasmDecoder): bool {
    return dec.byte() != ScBoolFalse;
}

export function boolEncode(enc: WasmEncoder, value: bool): void {
    enc.byte((value ? ScBoolTrue : ScBoolFalse) as u8);
}

export function boolFromBytes(buf: Uint8Array): bool {
    if (buf.length == 0) {
        return false;
    }
    if (buf.length != ScBoolLength) {
        panic("invalid Bool length");
    }
    if (buf[0] == ScBoolFalse) {
        return false;
    }
    if (buf[0] != ScBoolTrue) {
        panic("invalid Bool value");
    }
    return true;
}

export function boolToBytes(value: bool): Uint8Array {
    const buf = new Uint8Array(ScBoolLength);
    buf[0] = (value ? ScBoolTrue : ScBoolFalse) as u8;
    return buf;
}

export function boolFromString(value: string): bool {
    if (value == "0") {
        return false;
    }
    if (value == "1") {
        return true;
    }
    panic("invalid bool string");
    return false;
}

export function boolToString(value: bool): string {
    return value ? "1" : "0";
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableBool {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return boolToString(this.value());
    }

    value(): bool {
        return boolFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableBool extends ScImmutableBool {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: bool): void {
        this.proxy.set(boolToBytes(value));
    }
}
