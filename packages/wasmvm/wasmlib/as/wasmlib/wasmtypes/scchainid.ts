// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder, zeroes} from "./codec";
import {Proxy} from "./proxy";
import {addressFromBytes, addressFromString, addressToString, ScAddress, ScAddressAlias} from "./scaddress";
import {bytesCompare} from "./scbytes";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScChainIDLength = 32;

export class ScChainID {
    id: Uint8Array = zeroes(ScChainIDLength);

    public address(): ScAddress {
        const buf = new Uint8Array(this.id.length + 1);
        buf[0] = ScAddressAlias as u8;
        buf.set(this.id, 1);
        return addressFromBytes(buf);
    }

    public equals(other: ScChainID): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): Uint8Array {
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
    enc.fixedBytes(value.id, ScChainIDLength);
}

export function chainIDFromBytes(buf: Uint8Array | null): ScChainID {
    if (buf === null || buf.length == 0) {
        return new ScChainID();
    }
    if (buf.length != ScChainIDLength) {
        panic("invalid ChainID length");
    }
    return chainIDFromBytesUnchecked(buf);
}

export function chainIDToBytes(value: ScChainID): Uint8Array {
    return value.id;
}

export function chainIDFromString(value: string): ScChainID {
    const addr = addressFromString(value);
    if (addr.id[0] != ScAddressAlias) {
        panic("invalid ChainID address type");
    }
    return chainIDFromBytes(addr.id.slice(1));
}

export function chainIDToString(value: ScChainID): string {
    return addressToString(value.address());
}

function chainIDFromBytesUnchecked(buf: Uint8Array): ScChainID {
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
