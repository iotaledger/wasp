// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";
import {addressToBytes} from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScNftIDLength = 20;

export class ScNftID {
    id: u8[] = wasmtypes.zeroes(ScNftIDLength);

    public equals(other: ScNftID): bool {
        return wasmtypes.bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return nftIDToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        // TODO standardize human readable string
        return nftIDToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function nftIDDecode(dec: wasmtypes.WasmDecoder): ScNftID {
    return nftIDFromBytesUnchecked(dec.fixedBytes(ScNftIDLength));
}

export function nftIDEncode(enc: wasmtypes.WasmEncoder, value: ScNftID): void {
    enc.fixedBytes(value.id, ScNftIDLength);
}

export function nftIDFromBytes(buf: u8[]): ScNftID {
    if (buf.length == 0) {
        return new ScNftID();
    }
    if (buf.length != ScNftIDLength) {
        panic("invalid NftID length");
    }
    return nftIDFromBytesUnchecked(buf);
}

export function nftIDToBytes(value: ScNftID): u8[] {
    return value.id;
}

export function nftIDFromString(value: string): ScNftID {
    return nftIDFromBytes(wasmtypes.base58Decode(value));
}

export function nftIDToString(value: ScNftID): string {
    // TODO standardize human readable string
    return wasmtypes.base58Encode(nftIDToBytes(value));
}

function nftIDFromBytesUnchecked(buf: u8[]): ScNftID {
    let o = new ScNftID();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableNftID {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return nftIDToString(this.value());
    }

    value(): ScNftID {
        return nftIDFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableNftID extends ScImmutableNftID {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScNftID): void {
        this.proxy.set(nftIDToBytes(value));
    }
}
