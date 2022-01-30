// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";
import {bytesCompare} from "./scbytes";
import {uint32FromBytes, uint32ToBytes} from "./scuint32";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScHnameLength = 4;

export class ScHname {
    id: u8[];

    constructor(id: u32) {
        this.id = uint32ToBytes(id);
    }

    public equals(other: ScHname): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return hnameToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return hnameToString(this)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function hnameDecode(dec: WasmDecoder): ScHname {
    return hnameFromBytesUnchecked(dec.fixedBytes(ScHnameLength));
}

export function hnameEncode(enc: WasmEncoder, value: ScHname): void {
    enc.fixedBytes(value.toBytes(), ScHnameLength);
}

export function hnameFromBytes(buf: u8[] | null): ScHname {
    if (buf == null) {
        return new ScHname(0);
    }
    if (buf.length != ScHnameLength) {
        panic("invalid Hname length");
    }
    return hnameFromBytesUnchecked(buf);
}

export function hnameToBytes(value: ScHname): u8[] {
    return value.id;
}

export function hnameToString(value: ScHname): string {
    // TODO standardize human readable string
    const res = uint32FromBytes(value.id).toString(16);
    return "0000000".slice(0, 8 - res.length) + res;
}

function hnameFromBytesUnchecked(buf: u8[]): ScHname {
    let o = new ScHname(0);
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableHname {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return hnameToString(this.value());
    }

    value(): ScHname {
        return hnameFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableHname extends ScImmutableHname {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScHname): void {
        this.proxy.set(hnameToBytes(value));
    }
}
