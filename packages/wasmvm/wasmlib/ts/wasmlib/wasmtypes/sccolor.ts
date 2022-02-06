// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {base58Encode, WasmDecoder, WasmEncoder, zeroes} from "./codec";
import {Proxy} from "./proxy";
import {bytesCompare} from "./scbytes";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScColorLength = 32;

export class ScColor {
    id: u8[] = zeroes(ScColorLength);

    constructor(fill: u8) {
        this.id.fill(fill);
    }

    public equals(other: ScColor): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return colorToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        // TODO standardize human readable string
        return colorToString(this);
    }
}

// predefined colors
export const IOTA: ScColor = new ScColor(0x00);
export const MINT: ScColor = new ScColor(0xff);

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function colorDecode(dec: WasmDecoder): ScColor {
    return colorFromBytesUnchecked(dec.fixedBytes(ScColorLength));
}

export function colorEncode(enc: WasmEncoder, value: ScColor): void {
    enc.fixedBytes(value.toBytes(), ScColorLength);
}

export function colorFromBytes(buf: u8[]): ScColor {
    if (buf.length == 0) {
        return new ScColor(0);
    }
    if (buf.length != ScColorLength) {
        panic("invalid Color length");
    }
    return colorFromBytesUnchecked(buf);
}

export function colorToBytes(value: ScColor): u8[] {
    return value.id;
}

export function colorToString(value: ScColor): string {
    // TODO standardize human readable string
    return base58Encode(value.id);
}

function colorFromBytesUnchecked(buf: u8[]): ScColor {
    let o = new ScColor(0);
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableColor {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return colorToString(this.value());
    }

    value(): ScColor {
        return colorFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableColor extends ScImmutableColor {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScColor): void {
        this.proxy.set(colorToBytes(value));
    }
}
