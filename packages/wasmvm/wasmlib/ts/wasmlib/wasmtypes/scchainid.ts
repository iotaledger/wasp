// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {base58Encode, WasmDecoder, WasmEncoder, zeroes} from "./codec";
import {Proxy} from "./proxy";
import {addressFromBytes, ScAddress} from "./scaddress";
import {bytesCompare} from "./scbytes";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScChainIDLength = 33;

export class ScChainID {
    id: u8[] = zeroes(ScChainIDLength);

    public address(): ScAddress {
        return addressFromBytes(this.id);
    }

    public equals(other: ScChainID): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return chainIDToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return chainIDToString(this)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function chainIDDecode(dec: WasmDecoder): ScChainID {
    return chainIDFromBytesUnchecked(dec.fixedBytes(ScChainIDLength));
}

export function chainIDEncode(enc: WasmEncoder, value: ScChainID): void {
    enc.fixedBytes(value.toBytes(), ScChainIDLength);
}

export function chainIDFromBytes(buf: u8[]): ScChainID {
    if (buf.length == 0) {
        return new ScChainID();
    }
    if (buf.length != ScChainIDLength) {
        panic("invalid ChainID length");
    }
    // must be ledgerstate.AliasAddressType
    if (buf[0] != 2) {
        panic("invalid ChainID: not an alias address");
    }
    return chainIDFromBytesUnchecked(buf);
}

export function chainIDToBytes(value: ScChainID): u8[] {
    return value.id;
}

export function chainIDToString(value: ScChainID): string {
    // TODO standardize human readable string
    return base58Encode(value.id);
}

function chainIDFromBytesUnchecked(buf: u8[]): ScChainID {
    let o = new ScChainID();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableChainID {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return chainIDToString(this.value());
    }

    value(): ScChainID {
        return chainIDFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableChainID extends ScImmutableChainID {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScChainID): void {
        this.proxy.set(chainIDToBytes(value));
    }
}
