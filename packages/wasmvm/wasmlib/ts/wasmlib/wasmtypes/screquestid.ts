// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScRequestIDLength = 34;

export class ScRequestID {
    id: u8[] = wasmtypes.zeroes(ScRequestIDLength);

    public equals(other: ScRequestID): bool {
        return wasmtypes.bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return requestIDToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return requestIDToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function requestIDDecode(dec: wasmtypes.WasmDecoder): ScRequestID {
    return requestIDFromBytesUnchecked(dec.fixedBytes(ScRequestIDLength));
}

export function requestIDEncode(enc: wasmtypes.WasmEncoder, value: ScRequestID): void {
    enc.fixedBytes(value.id, ScRequestIDLength);
}

export function requestIDFromBytes(buf: u8[]): ScRequestID {
    if (buf.length == 0) {
        return new ScRequestID();
    }
    if (buf.length != ScRequestIDLength) {
        panic("invalid RequestID length");
    }
    // final uint16 output index must be > ledgerstate.MaxOutputCount
    if (buf[ScRequestIDLength - 2] > 127 || buf[ScRequestIDLength - 1] != 0) {
        panic("invalid RequestID: output index > 127");
    }
    return requestIDFromBytesUnchecked(buf);
}

export function requestIDToBytes(value: ScRequestID): u8[] {
    return value.id;
}

export function requestIDToString(value: ScRequestID): string {
    // TODO standardize human readable string
    return wasmtypes.base58Encode(value.id);
}

function requestIDFromBytesUnchecked(buf: u8[]): ScRequestID {
    let o = new ScRequestID();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableRequestID {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return requestIDToString(this.value());
    }

    value(): ScRequestID {
        return requestIDFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableRequestID extends ScImmutableRequestID {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScRequestID): void {
        this.proxy.set(requestIDToBytes(value));
    }
}
