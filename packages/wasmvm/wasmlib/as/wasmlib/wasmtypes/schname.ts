// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {sandbox} from "../host";
import {FnUtilsHashName, panic} from "../sandbox";
import {uint32FromBytes, uint32ToBytes} from "./scuint32";
import {stringToBytes} from "./scstring";
import {bytesCompare} from "./scbytes";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScHnameLength = 4;

export class ScHname {
    id: Uint8Array;

    constructor(id: u32) {
        this.id = uint32ToBytes(id);
    }

    static fromName(name: string): ScHname {
        return hnameFromBytes(sandbox(FnUtilsHashName, stringToBytes(name)));
    }

    public equals(other: ScHname): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): Uint8Array {
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

export function hnameFromBytes(buf: Uint8Array): ScHname {
    if (buf.length == 0) {
        return new ScHname(0);
    }
    if (buf.length != ScHnameLength) {
        panic("invalid Hname length");
    }
    return hnameFromBytesUnchecked(buf);
}

export function hnameToBytes(value: ScHname): Uint8Array {
    return value.id;
}

export function hnameFromString(value: string): ScHname {
    if (value.length > 8) {
        panic("invalid Hname string");
    }
    return new ScHname(parseInt(value, 16) as u32);
}

export function hnameToString(value: ScHname): string {
    const res = uint32FromBytes(value.id).toString(16);
    return "0000000".slice(0, 8 - res.length) + res;
}

function hnameFromBytesUnchecked(buf: Uint8Array): ScHname {
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
