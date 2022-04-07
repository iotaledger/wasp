// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScTokenIDLength = 38;

export class ScTokenID {
    id: u8[] = wasmtypes.zeroes(ScTokenIDLength);

    public equals(other: ScTokenID): bool {
        return wasmtypes.bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return tokenIDToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        // TODO standardize human readable string
        return tokenIDToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function tokenIDDecode(dec: wasmtypes.WasmDecoder): ScTokenID {
    return tokenIDFromBytesUnchecked(dec.fixedBytes(ScTokenIDLength));
}

export function tokenIDEncode(enc: wasmtypes.WasmEncoder, value: ScTokenID): void {
    enc.fixedBytes(value.id, ScTokenIDLength);
}

export function tokenIDFromBytes(buf: u8[]): ScTokenID {
    if (buf.length == 0) {
        return new ScTokenID();
    }
    if (buf.length != ScTokenIDLength) {
        panic("invalid TokenID length");
    }
    return tokenIDFromBytesUnchecked(buf);
}

export function tokenIDToBytes(value: ScTokenID): u8[] {
    return value.id;
}

export function tokenIDToString(value: ScTokenID): string {
    // TODO standardize human readable string
    return wasmtypes.base58Encode(value.id);
}

function tokenIDFromBytesUnchecked(buf: u8[]): ScTokenID {
    let o = new ScTokenID();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableTokenID {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return tokenIDToString(this.value());
    }

    value(): ScTokenID {
        return tokenIDFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableTokenID extends ScImmutableTokenID {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScTokenID): void {
        this.proxy.set(tokenIDToBytes(value));
    }
}
