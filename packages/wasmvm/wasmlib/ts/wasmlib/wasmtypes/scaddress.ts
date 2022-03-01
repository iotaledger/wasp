// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index"

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScAddressAlias  : u8 = 2;
export const ScAddressEd25519: u8 = 0;
export const ScAddressNFT    : u8 = 1;

export const ScAddressLength = 33;

export class ScAddress {
    id: u8[] = wasmtypes.zeroes(ScAddressLength);

    asAgentID(): wasmtypes.ScAgentID {
        // agentID for address has Hname zero
        return new wasmtypes.ScAgentID(this, new wasmtypes.ScHname(0));
    }

    public equals(other: ScAddress): bool {
        return wasmtypes.bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return addressToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return addressToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function addressDecode(dec: wasmtypes.WasmDecoder): ScAddress {
    return addressFromBytesUnchecked(dec.fixedBytes(ScAddressLength))
}

export function addressEncode(enc: wasmtypes.WasmEncoder, value: ScAddress): void {
    enc.fixedBytes(value.toBytes(), ScAddressLength)
}

export function addressFromBytes(buf: u8[]): ScAddress {
    if (buf.length == 0) {
        return new ScAddress();
    }
    if (buf.length != ScAddressLength) {
        panic("invalid Address length");
    }
    if (buf[0] > ScAddressAlias) {
        panic("invalid Address: address type > 2");
    }
    return addressFromBytesUnchecked(buf);
}

export function addressToBytes(value: ScAddress): u8[] {
    return value.id;
}

export function addressToString(value: ScAddress): string {
    // TODO standardize human readable string
    return wasmtypes.base58Encode(value.id);
}

function addressFromBytesUnchecked(buf: u8[]): ScAddress {
    let o = new ScAddress();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableAddress {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return addressToString(this.value());
    }

    value(): ScAddress {
        return addressFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableAddress extends ScImmutableAddress {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScAddress): void {
        this.proxy.set(addressToBytes(value));
    }
}
