// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { panic } from "../sandbox";
import * as wasmtypes from "./index";
import {addressToBytes} from "./index";
import {ScSandboxUtils} from "../sandboxutils";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScChainIDLength = 32;

export class ScChainID {
    id: u8[] = wasmtypes.zeroes(ScChainIDLength);

    public address(): wasmtypes.ScAddress {
        const buf: u8[] = [wasmtypes.ScAddressAlias];
        return wasmtypes.addressFromBytes(buf.concat(this.id));
    }

    public equals(other: ScChainID): bool {
        return wasmtypes.bytesCompare(this.id, other.id) == 0;
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

export function chainIDDecode(dec: wasmtypes.WasmDecoder): ScChainID {
    return chainIDFromBytesUnchecked(dec.fixedBytes(ScChainIDLength));
}

export function chainIDEncode(enc: wasmtypes.WasmEncoder, value: ScChainID): void {
    enc.fixedBytes(value.id, ScChainIDLength);
}

export function chainIDFromBytes(buf: u8[]): ScChainID {
    if (buf.length == 0) {
        return new ScChainID();
    }
    if (buf.length != ScChainIDLength) {
        panic("invalid ChainID length");
    }
    return chainIDFromBytesUnchecked(buf);
}

export function chainIDToBytes(value: ScChainID): u8[] {
    return value.id;
}

export function chainIDFromString(value: string): ScChainID {
    const addr = wasmtypes.addressFromString(value);
    if (addr.id[0] != wasmtypes.ScAddressAlias) {
        panic("invalid ChainID address type");
    }
    return chainIDFromBytes(addr.id.slice(1));
}

export function chainIDToString(value: ScChainID): string {
    return wasmtypes.addressToString(value.address());
}

function chainIDFromBytesUnchecked(buf: u8[]): ScChainID {
    let o = new ScChainID();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableChainID {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
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
