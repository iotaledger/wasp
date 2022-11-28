// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {hexDecode, hexEncode, WasmDecoder, WasmEncoder, zeroes} from "./codec";
import {Proxy} from "./proxy";
import {bytesCompare} from "./scbytes";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScNftIDLength = 32;

export class ScNftID {
    id: Uint8Array = zeroes(ScNftIDLength);

    public equals(other: ScNftID): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): Uint8Array {
        return nftIDToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return nftIDToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function nftIDDecode(dec: WasmDecoder): ScNftID {
    return nftIDFromBytesUnchecked(dec.fixedBytes(ScNftIDLength));
}

export function nftIDEncode(enc: WasmEncoder, value: ScNftID): void {
    enc.fixedBytes(value.id, ScNftIDLength);
}

export function nftIDFromBytes(buf: Uint8Array | null): ScNftID {
    if (buf === null || buf.length == 0) {
        return new ScNftID();
    }
    if (buf.length != ScNftIDLength) {
        panic("invalid NftID length");
    }
    return nftIDFromBytesUnchecked(buf);
}

export function nftIDToBytes(value: ScNftID): Uint8Array {
    return value.id;
}

export function nftIDFromString(value: string): ScNftID {
    return nftIDFromBytes(hexDecode(value));
}

export function nftIDToString(value: ScNftID): string {
    return hexEncode(nftIDToBytes(value));
}

function nftIDFromBytesUnchecked(buf: Uint8Array): ScNftID {
    let o = new ScNftID();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableNftID {
    proxy: Proxy;

    constructor(proxy: Proxy) {
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
