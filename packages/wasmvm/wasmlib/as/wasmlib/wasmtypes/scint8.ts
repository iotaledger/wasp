// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {intFromString, WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

export const ScInt8Length = 1;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function int8Decode(dec: WasmDecoder): i8 {
    return dec.byte() as i8;
}

export function int8Encode(enc: WasmEncoder, value: i8): void {
    enc.byte(value as u8);
}

export function int8FromBytes(buf: Uint8Array): i8 {
    if (buf.length == 0) {
        return 0;
    }
    if (buf.length != ScInt8Length) {
        panic("invalid Int8 length");
    }
    return buf[0] as i8;
}

export function int8ToBytes(value: i8): Uint8Array {
    const buf = new Uint8Array(ScInt8Length);
    buf[0] = value as u8;
    return buf;
}

export function int8FromString(value: string): i8 {
    return intFromString(value, 8) as i8;
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
