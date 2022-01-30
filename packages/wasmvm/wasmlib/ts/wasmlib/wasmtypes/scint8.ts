// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

export const ScInt8Length = 1;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function int8Decode(dec: WasmDecoder): i8 {
    return dec.byte() as i8;
}

export function int8Encode(enc: WasmEncoder, value: i8): void {
    enc.byte(value as u8);
}

export function int8FromBytes(buf: u8[] | null): i8 {
    if (buf == null) {
        return 0;
    }
    if (buf.length != ScInt8Length) {
        panic("invalid Int8 length");
    }
    return buf[0] as i8;
}

export function int8ToBytes(value: i8): u8[] {
    return [value as u8];
}

export function int8ToString(value: i8): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableInt8 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return int8ToString(this.value());
    }

    value(): i8 {
        return int8FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableInt8 extends ScImmutableInt8 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: i8): void {
        this.proxy.set(int8ToBytes(value));
    }
}
